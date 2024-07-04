package sslfsr

import (
	"math"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test16BitShift(t *testing.T) {
	t.Parallel()
	for i := range math.MaxUint16 + 1 {
		reg := NewSSLFSR16(0) // not using interval
		reg.register = uint16(i)

		for range math.MaxUint16 {
			reg.Shift()
		}

		assert.Equal(t, uint16(i), reg.register, "65535 shifts should result in starting state")
	}
}

func Test16BitSubShift(t *testing.T) {
	t.Parallel()
	for i := range math.MaxUint16 {
		reg := NewSSLFSR16(0)
		reg.register = uint16(i)

		for range math.MaxUint8 {
			reg.SubShift()
		}

		assert.Equal(t, uint16(i), reg.register, "255 subshifts should result in starting state")
	}
}

func Test16BitIntervals(t *testing.T) {
	t.Skip("This test takes too long to run")

	for _, interval := range Intervals16Bits() {
		reg := NewSSLFSR16(uint16(interval))
		start := reg.register

		reg.Next()
		count := 1
		for reg.register != start || reg.counter != 0 {
			reg.Next()
			count++
		}

		assert.Equal(t, reg.CalculateExpectedMaximalLength(), count, "should return to start state in ((2^16)-1)(%d+1) state changes", interval)
	}
}

func TestNonIntervals16Bits(t *testing.T) {
	t.Skip("This test takes too long to run")

	intervals := Intervals16Bits()
	for i := 2; i <= math.MaxUint16; i++ {
		if slices.Contains(intervals, i) {
			continue
		}

		reg := NewSSLFSR16(uint16(i))
		start := reg.register
		reg.Next()
		count := 1
		for reg.register != start || reg.counter != 0 {
			reg.Next()
			count++
		}

		assert.NotEqual(t, reg.CalculateExpectedMaximalLength(), count)
	}
}
