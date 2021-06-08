package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/coreyog/sslfsr"
	"github.com/coreyog/statux"
	flags "github.com/jessevdk/go-flags"
)

const (
	expectedStateChanges = math.MaxUint8
)

type Arguments struct {
	Start int `short:"s" default:"2"`
	End   int `short:"e" default:"255"`
}

func (a *Arguments) Validate() (err error) {
	if a == nil {
		return errors.New("nil arguments")
	}

	if a.Start < 2 || a.Start > expectedStateChanges {
		return fmt.Errorf("Start parameter must be between 2 and %d: %d", expectedStateChanges, a.Start)
	}

	if a.End < a.Start || a.End > expectedStateChanges {
		return fmt.Errorf("End parameter must be between Start (%d) and %d: %d", a.Start, expectedStateChanges, a.End)
	}

	if a.Start == a.End {
		a.End++
	}

	return nil
}

func (a *Arguments) IsFullRange() bool {
	return a != nil && a.Start == 2 && a.End == expectedStateChanges
}

var args = &Arguments{}

func main() {
	_, err := flags.Parse(args)
	if err != nil {
		if flags.WroteHelp(err) {
			os.Exit(0)
		}
		panic(err)
	}

	err = args.Validate()
	if err != nil {
		panic(err)
	}

	// time execution
	start := time.Now()
	defer func() {
		fmt.Println()
		fmt.Printf("DONE: %s\n", time.Since(start))
	}()

	// prepare multiplexed logging
	cpus := runtime.NumCPU()
	cpus = int(math.Min(float64(cpus), float64(args.End-args.Start)))

	stat, err := statux.New(cpus)
	if err != nil {
		panic(err)
	}

	lines := stat.BuildLineWriters()

	// setup plumbing
	intervals := make(chan int, cpus*2)
	working := make(chan int, cpus)
	wg := &sync.WaitGroup{}
	wg.Add(cpus)

	// start workers
	for i := 0; i < cpus; i++ {
		go worker(lines[i], wg, intervals, working)
	}

	// gather results
	results := []int{}
	go func() {
		for w := range working {
			results = append(results, w)
		}
	}()

	// prepare for interruptions
	ctrlc := make(chan os.Signal, 1)
	signal.Notify(ctrlc, os.Interrupt)
	safety := sync.WaitGroup{} // after ctrl+c, this will stop main thread
	go func() {
		<-ctrlc       // notice interrupt
		safety.Add(1) // stop main thread
		stat.Finish() // dispose of multiplex logging
		fmt.Println()
		wrapup(results)
		os.Exit(1) // drop everything and quit
	}()

	// provide input for the workers
	for interval := args.Start; interval < args.End; interval++ {
		intervals <- interval
	}

	// indicate to workers that no more input is coming, they will close
	close(intervals)
	safety.Wait()
	wg.Wait() // wait for workers to finish their final tasks
	safety.Wait()
	close(working) // stop gathering results
	safety.Wait()

	stat.Finish() // dispose of multiplex logging
	fmt.Println()

	wrapup(results)
}

func wrapup(results []int) {
	sort.Ints(results) // cleaner output

	var out io.Writer // tee output results

	outfile, err := os.Create("results8.txt")
	if err != nil {
		out = os.Stdout
	} else {
		out = io.MultiWriter(os.Stdout, outfile)
	}

	bufout := bufio.NewWriter(out)

	_, _ = bufout.WriteString(fmt.Sprintf("tested intervals: [%d, %d)\n", args.Start, args.End))
	_, _ = bufout.WriteString(fmt.Sprintf("%v\n", results))
	_, _ = bufout.WriteString(fmt.Sprintf("working count: %d\n", len(results)))
	if args.IsFullRange() {
		match := reflect.DeepEqual(sslfsr.Intervals8Bits(), results)
		_, _ = bufout.WriteString(fmt.Sprintf("matches expected results: %t\n", match))
	}

	bufout.Flush()
}

func worker(logger io.StringWriter, wg *sync.WaitGroup, todo <-chan int, working chan<- int) {
	for interval := range todo {
		_, _ = logger.WriteString(fmt.Sprintf("%d - Building Calculator...", interval))

		calculator := build8BitCalculator(interval)

		value := uint8(1)
		count := 1

		_, _ = logger.WriteString(fmt.Sprintf("%d - Calculatoring...", interval))

		visited := map[uint8]struct{}{}
		visited[value] = struct{}{}

		value = calculator[value]
		visited[value] = struct{}{}
		var ok bool
		for count < expectedStateChanges {
			count++
			value = calculator[value]
			_, ok = visited[value]
			if ok {
				break
			}
			visited[value] = struct{}{}
		}

		if count == expectedStateChanges {
			working <- interval
		}
	}
	_, _ = logger.WriteString("DONE")
	wg.Done()
}

func build8BitCalculator(interval int) (calc map[uint8]uint8) {
	calc = map[uint8]uint8{}

	for i := 1; i <= math.MaxUint8; i++ {
		b := uint8(i)

		for j := 0; j < interval; j++ {
			b = shift8Bits(b)
		}

		calc[uint8(i)] = subshift8Bits(b)
	}

	return calc
}

func shift8Bits(value uint8) uint8 {
	bit := sslfsr.GetBit(value, 0) != sslfsr.GetBit(value, 2) != sslfsr.GetBit(value, 3) != sslfsr.GetBit(value, 4)
	value = value >> 1
	if bit {
		value = value | 0x80
	}

	return value
}

func subshift8Bits(value uint8) uint8 {
	bit := sslfsr.GetBit(value, 0) != sslfsr.GetBit(value, 1)
	higher := value & 0xF0
	lower := value & 0x0F
	lower = lower >> 1
	if bit {
		lower = lower | 0x8
	}

	return lower | higher
}
