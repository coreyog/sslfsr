package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"strconv"
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
		fmt.Println()
		fmt.Printf("DONE: %s\n", time.Since(start))
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

	args := os.Args

	if len(args) > 1 {
		args = args[1:]
	} else {
		fmt.Println("please provide intervals to verify")
		os.Exit(1)
	}

	intervals := make([]int, 0, len(args))
	for _, arg := range args {
		num, err := strconv.Atoi(arg)
		if err != nil {
			fmt.Printf("invalid arg: %s - %s\n", arg, err)
			continue
		}

		intervals = append(intervals, num)
	}

	const startValue = uint16(1)
	for _, inter := range intervals {
		register := startValue
		visited := make([]bool, expectedStateChanges+1)

		for i := 0; i < expectedStateChanges; i++ {
			for j := 0; j < inter; j++ {
				register = shift16Bits(register)
			}
			register = subshift16Bits(register)
			visited[register] = true
		}

		working := register == startValue
		for i := 1; working && i < len(visited); i++ {
			working = working && visited[i]
		}

		fmt.Printf("%d - %t\n", inter, working)
	}
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
