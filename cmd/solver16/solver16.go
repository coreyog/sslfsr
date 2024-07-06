package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/coreyog/sslfsr"
	"github.com/coreyog/statux"
)

//go:generate go build "-gcflags=all=-N -l" .

type StateMap [math.MaxUint16 + 1]uint16

type WorkItem struct {
	Lut      StateMap
	Interval int
}

var memoShift16Bits StateMap
var memoSubshift16Bits StateMap

func init() {
	for i := range math.MaxUint16 + 1 {
		memoShift16Bits[i] = sslfsr.Shift16Bits(uint16(i))
		memoSubshift16Bits[i] = sslfsr.SubShift16Bits(uint16(i))
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--wfd" {
		fmt.Println("waiting for debugger...")
		debugger := true
		for debugger {
			time.Sleep(100 * time.Millisecond) // breakpoint here
		}
	}

	// time execution
	start := time.Now()
	defer func() {
		fmt.Println()
		fmt.Printf("DONE: %s\n", time.Since(start))
	}()

	// prepare multiplexed logging
	cpus := runtime.NumCPU()

	stat, err := statux.New(cpus)
	if err != nil {
		panic(err)
	}
	lines := stat.BuildLineWriters()

	// setup plumbing
	luts := make(chan *WorkItem, cpus*2)
	working := make(chan int, cpus)
	wg := &sync.WaitGroup{}
	wg.Add(cpus)

	// start workers
	for i := 0; i < cpus; i++ {
		go worker(lines[i], luts, working, wg)
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
		<-ctrlc // notice interrupt
		fmt.Printf("INTERRUPT: %s\n", time.Since(start))
		safety.Add(1) // stop main thread
		stat.Finish() // dispose of multiplex logging
		fmt.Println() // easy to read output
		wrapup(results)
	}()

	build16BitLUTs(luts)

	// indicate to workers that no more input is coming, they will close
	close(luts)
	safety.Wait()
	wg.Wait() // wait for workers to finish their final tasks
	safety.Wait()
	close(working) // stop gathering results
	safety.Wait()

	safety.Wait()

	stat.Finish() // dispose of multiplex logging
	fmt.Println()

	wrapup(results)
}

func wrapup(results []int) {
	sort.Ints(results) // cleaner output

	var out io.Writer // tee output results

	outfile, err := os.Create("results16.txt")
	if err != nil {
		out = os.Stdout
	} else {
		out = io.MultiWriter(os.Stdout, outfile)
	}

	bufout := bufio.NewWriter(out)

	_, _ = bufout.WriteString(fmt.Sprintf("tested intervals: [%d, %d]\n", 0, math.MaxUint16))
	_, _ = bufout.WriteString(fmt.Sprintf("%v\n", results))
	_, _ = bufout.WriteString(fmt.Sprintf("working count: %d\n", len(results)))
	match := reflect.DeepEqual(sslfsr.Intervals16Bits(), results)
	_, _ = bufout.WriteString(fmt.Sprintf("matches expected results: %t\n", match))

	bufout.Flush()

	if !match {
		// non-zero exit code indicates not all intervals were verified
		os.Exit(1)
	}
}

func worker(logger io.StringWriter, todo <-chan *WorkItem, working chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()

	for work := range todo {
		_, _ = logger.WriteString(strconv.Itoa(work.Interval))

		value := uint16(1)
		count := 0

		visited := [math.MaxUint16 + 1]bool{}
		var fail bool

		for range math.MaxUint16 {
			count++
			value = work.Lut[value] // shift, shift, ..., subshift = 1 interval

			if visited[value] {
				fail = true
				break
			}

			visited[value] = true
		}

		if !fail {
			working <- work.Interval
		}
	}

	_, _ = logger.WriteString("DONE")
}

func build16BitLUTs(c chan *WorkItem) {
	var shuttle StateMap
	for i := range shuttle {
		shuttle[i] = uint16(i)
	}

	for interval := 1; interval < math.MaxUint16; interval++ {
		var lut StateMap
		for i := 1; i <= math.MaxUint16; i++ {
			s := &shuttle[i]

			*s = memoShift16Bits[*s]

			lut[i] = memoSubshift16Bits[*s]
		}

		c <- &WorkItem{
			Lut:      lut,
			Interval: interval,
		}
	}
}
