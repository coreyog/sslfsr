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

type StateMap [math.MaxUint8 + 1]uint8

type WorkItem struct {
	LUT      StateMap
	Interval int
}

var memoShift8Bits StateMap
var memoSubshift8Bits StateMap

func init() {
	for i := range math.MaxUint8 + 1 {
		memoShift8Bits[i] = sslfsr.Shift8Bits(uint8(i))
		memoSubshift8Bits[i] = sslfsr.SubShift8Bits(uint8(i))
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
		fmt.Println()
		wrapup(results)
	}()

	build8BitLUTs(luts)

	// indicate to workers that no more input is coming, they will close
	close(luts)
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

	_, _ = bufout.WriteString(fmt.Sprintf("tested intervals: [%d, %d)\n", 0, math.MaxUint8))
	_, _ = bufout.WriteString(fmt.Sprintf("%v\n", results))
	_, _ = bufout.WriteString(fmt.Sprintf("working count: %d\n", len(results)))
	match := reflect.DeepEqual(sslfsr.Intervals8Bits(), results)
	_, _ = bufout.WriteString(fmt.Sprintf("matches expected results: %t\n", match))

	bufout.Flush()

	if !match {
		os.Exit(1)
	}
}

func worker(logger io.StringWriter, todo <-chan *WorkItem, working chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()

	for work := range todo {
		_, _ = logger.WriteString(strconv.Itoa(work.Interval))

		value := uint8(1)
		count := 0

		visited := [math.MaxUint8 + 1]bool{}
		var fail bool

		for range math.MaxUint8 {
			count++
			value = work.LUT[value] // shift, shift, ..., subshift = 1 interval

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

func build8BitLUTs(c chan *WorkItem) {
	var shuttle StateMap
	for i := range shuttle {
		shuttle[i] = uint8(i)
	}

	for interval := 1; interval < math.MaxUint8; interval++ {
		var lut StateMap
		for i := 1; i <= math.MaxUint8; i++ {
			s := &shuttle[i]

			*s = memoShift8Bits[*s]

			lut[i] = memoSubshift8Bits[*s]
		}

		c <- &WorkItem{
			LUT:      lut,
			Interval: interval,
		}
	}
}
