package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/coreyog/sslfsr"
	"github.com/coreyog/statux"
)

const (
	expectedStateChanges = math.MaxUint16
)

func main() {
	start := time.Now()
	defer func() {
		fmt.Println()
		fmt.Printf("DONE: %s\n", time.Since(start))
	}()

	cpus := runtime.NumCPU()
	stat, err := statux.New(cpus)
	if err != nil {
		panic(err)
	}

	lines := stat.BuildLineWriters()

	intervals := make(chan int, cpus*2)
	working := make(chan int, cpus)
	wg := &sync.WaitGroup{}
	wg.Add(cpus)

	for i := 0; i < cpus; i++ {
		go worker(lines[i], wg, intervals, working)
	}

	results := []int{}
	go func() {
		for w := range working {
			results = append(results, w)
		}
	}()

	ctrlc := make(chan os.Signal)
	signal.Notify(ctrlc, os.Interrupt)
	safety := sync.WaitGroup{}
	go func() {
		<-ctrlc
		safety.Add(1)
		stat.Finish()
		sort.Ints(results)
		fmt.Println(results)
		fmt.Println()
		fmt.Printf("DONE: %s\n", time.Since(start))
		os.Exit(1)
	}()

	for interval := 2; interval < math.MaxUint16; interval++ {
		intervals <- interval
	}

	close(intervals)
	safety.Wait()
	wg.Wait()
	safety.Wait()
	close(working)
	safety.Wait()

	sort.Ints(results)
	safety.Wait()

	stat.Finish()

	var out io.Writer

	outfile, err := os.Open("results.txt")
	if err != nil {
		out = os.Stdout
	} else {
		out = io.MultiWriter(os.Stdout, outfile)
	}

	_, _ = out.Write([]byte(fmt.Sprintf("%v\n", results)))
	_, _ = out.Write([]byte(fmt.Sprintf("working count: %d\n", len(results))))
}

func worker(logger io.StringWriter, wg *sync.WaitGroup, todo <-chan int, working chan<- int) {
	for interval := range todo {
		_, _ = logger.WriteString(fmt.Sprintf("%d - Building Calculator...", interval))

		calculator := build16BitCalculator(interval)

		value := uint16(1)
		count := 1

		_, _ = logger.WriteString(fmt.Sprintf("%d - Calculatoring...", interval))

		visited := map[uint16]struct{}{}
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

func build16BitCalculator(interval int) (calc map[uint16]uint16) {
	calc = map[uint16]uint16{}

	for i := 1; i <= math.MaxUint16; i++ {
		s := uint16(i)

		for j := 0; j < interval; j++ {
			s = shift16Bits(s)
		}

		calc[uint16(i)] = subshift16Bits(s)
	}

	return calc
}

func shift16Bits(value uint16) uint16 {
	bit := sslfsr.GetBit16(value, 0) != sslfsr.GetBit16(value, 1) != sslfsr.GetBit16(value, 3) != sslfsr.GetBit16(value, 12)
	value = value >> 1
	if bit {
		value = value | 0x8000
	}

	return value
}

func subshift16Bits(value uint16) uint16 {
	bit := sslfsr.GetBit16(value, 0) != sslfsr.GetBit16(value, 2) != sslfsr.GetBit16(value, 3) != sslfsr.GetBit16(value, 4)
	higher := value & 0xFF00
	lower := value & 0x00FF
	lower = lower >> 1
	if bit {
		lower = lower | 0x80
	}

	return lower | higher
}

// func build8BitCalculator(interval int) (calc map[uint8]uint8) {
// 	calc = map[uint8]uint8{}

// 	for i := 1; i <= math.MaxUint8; i++ {
// 		b := uint8(i)

// 		for j := 0; j < interval; j++ {
// 			b = shift8Bits(b)
// 		}

// 		calc[uint8(i)] = subshift8Bits(b)
// 	}

// 	return calc
// }

// func shift8Bits(value byte) byte {
// 	bit := sslfsr.GetBit(value, 0) != sslfsr.GetBit(value, 2) != sslfsr.GetBit(value, 3) != sslfsr.GetBit(value, 4)
// 	value = value >> 1
// 	if bit {
// 		value = value | 0x80
// 	}

// 	return value
// }

// func subshift8Bits(value byte) byte {
// 	bit := sslfsr.GetBit(value, 0) != sslfsr.GetBit(value, 1)
// 	higher := value & 0xF0
// 	lower := value & 0x0F
// 	lower = lower >> 1
// 	if bit {
// 		lower = lower | 0x8
// 	}

// 	return lower | higher
// }
