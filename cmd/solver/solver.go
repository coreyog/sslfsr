package main

import (
	"fmt"
	"math"
	"reflect"
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

	working := []int{}

	stat, err := statux.New(1)
	if err != nil {
		panic(err)
	}

	lines := stat.BuildLineWriters()
	line := lines[0]

	for interval := 2; interval < 100; interval++ {
		line.WriteString(fmt.Sprintf("%d - Building Calculator...", interval))

		calculator := build16BitCalculator(interval)

		value := uint16(1)
		count := 1

		line.WriteString(fmt.Sprintf("%d - Calculatoring...", interval))

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
			working = append(working, interval)
		}
	}

	line.WriteString("FINISHED")
	stat.Finish()

	fmt.Println(working)
	fmt.Printf("working count: %d\n", len(working))
	match := reflect.DeepEqual(working, []int{7, 66, 99})
	fmt.Printf("Match: %t\n", match)
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

func shift8Bits(value byte) byte {
	bit := sslfsr.GetBit(value, 0) != sslfsr.GetBit(value, 2) != sslfsr.GetBit(value, 3) != sslfsr.GetBit(value, 4)
	value = value >> 1
	if bit {
		value = value | 0x80
	}

	return value
}

func subshift8Bits(value byte) byte {
	bit := sslfsr.GetBit(value, 0) != sslfsr.GetBit(value, 1)
	higher := value & 0xF0
	lower := value & 0x0F
	lower = lower >> 1
	if bit {
		lower = lower | 0x8
	}

	return lower | higher
}
