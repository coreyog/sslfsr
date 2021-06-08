package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/coreyog/sslfsr"
)

const (
	expectedStateChanges = math.MaxUint16
)

func main() {
	// time execution
	start := time.Now()
	defer func() {
		fmt.Printf("\n\nDONE: %s\n", time.Since(start))
	}()

	// prepare for interruptions
	ctrlc := make(chan os.Signal, 1)
	signal.Notify(ctrlc, os.Interrupt)
	go func() {
		<-ctrlc                                      // notice interrupt
		end := time.Now()                            // prepare for timing output
		fmt.Printf("\n\nDONE: %s\n", end.Sub(start)) // time execution output
		os.Exit(1)                                   // drop everything and quit
	}()

	intervals := processArgs()

	// calculations
	for _, inter := range intervals {
		working := verifyInterval(inter)

		fmt.Printf("%d - %t\n", inter, working)
	}
}

func processArgs() []int {
	args := os.Args

	if len(args) > 1 {
		args = args[1:] // remove $0, the binary
	} else {
		fmt.Println("please provide intervals to verify")
		os.Exit(1)
	}

	intervals := make([]int, 0, len(args))
	for _, arg := range args {
		parts := strings.Split(arg, "-") // check for ranges: 2-100
		if len(parts) == 1 {             // not a range, just a number: 42
			num, err := strconv.Atoi(arg)
			if err != nil {
				invalidArg(arg, err)
				continue
			}

			if num < 2 || num > expectedStateChanges {
				invalidArg(arg, fmt.Errorf("must be >2 and <%d", expectedStateChanges))
				continue
			}

			// add the number
			intervals = append(intervals, num)
		} else if len(parts) == 2 { // range!
			// parse parts
			low, err := strconv.Atoi(parts[0])
			if err != nil {
				invalidArg(arg, err)
				continue
			}
			high, err := strconv.Atoi(parts[1])
			if err != nil {
				invalidArg(arg, err)
				continue
			}
			// sanity checks
			if low < 2 && low < expectedStateChanges {
				invalidArg(arg, fmt.Errorf("must be >2 and <%d", expectedStateChanges))
				continue
			}
			if high < 2 && high < expectedStateChanges {
				invalidArg(arg, fmt.Errorf("must be >2 and <%d", expectedStateChanges))
				continue
			}
			if low >= high {
				invalidArg(arg, err)
				continue
			}
			for i := low; i <= high; i++ { // do it
				intervals = append(intervals, i)
			}
		} else {
			invalidArg(arg, fmt.Errorf("unsupported number of dashes"))
			continue
		}
	}

	// deduplicate intervals
	// https://github.com/golang/go/wiki/SliceTricks#in-place-deduplicate-comparable
	sort.Ints(intervals)
	j := 0
	for i := 1; i < len(intervals); i++ {
		if intervals[j] == intervals[i] {
			continue
		}
		j++
		intervals[j] = intervals[i]
	}
	intervals = intervals[:j+1]

	return intervals
}

func verifyInterval(inter int) bool {
	const startValue = uint16(1)
	// initialize
	register := startValue
	visited := make([]bool, expectedStateChanges+1)

	// calculate less number of calculations
	for i := 0; i < expectedStateChanges; i++ {
		for j := 0; j < inter; j++ {
			// `interval` number of shifts...
			register = shift16Bits(register)
		}
		// .. then a subshift
		register = subshift16Bits(register)
		// check visisted
		ok := visited[register]
		if ok {
			break
		}
		// mark visited
		visited[register] = true
	}

	// verify ...
	working := register == startValue              // ... final state === starting state
	for i := 1; working && i < len(visited); i++ { // ... all states (excepte 0) were marked as visited
		working = working && visited[i]
	}

	return working
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

func invalidArg(arg string, err error) {
	fmt.Printf("invalid arg: %s - %s\n", arg, err)
}
