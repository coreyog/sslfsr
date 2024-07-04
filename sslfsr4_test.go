package sslfsr

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSettersAndGetters4Bits(t *testing.T) {
	t.Parallel()

	reg := BuildSSLFSR4(1, 2, 3)

	assert.Equal(t, uint8(1), reg.GetRegister())
	assert.Equal(t, uint8(2), reg.GetInterval())
	assert.Equal(t, uint8(3), reg.GetCounter())
}

func Test4BitShift(t *testing.T) {
	t.Parallel()

	for i := range MaxUint4 + 1 {
		reg := NewSSLFSR4(0) // not using interval
		reg.register = uint8(i)

		for range MaxUint4 {
			reg.Shift()
		}

		assert.Equal(t, uint8(i), reg.register, "15 shifts should result in starting state")
	}
}

func Test4BitSubShift(t *testing.T) {
	t.Parallel()

	for i := range MaxUint4 + 1 {
		reg := NewSSLFSR4(0)
		reg.register = uint8(i)

		for range MaxUint4 {
			reg.SubShift()
		}

		assert.Equal(t, uint8(i), reg.register, "15 subshifts should result in starting state")
	}
}

func TestIntervals4Bits(t *testing.T) {
	t.Parallel()
	intervals := Intervals4Bits()

	for _, interval := range intervals {
		reg := NewSSLFSR4(uint8(interval))
		start := reg.register

		reg.Next()
		count := 1
		for reg.register != start || reg.counter != 0 {
			reg.Next()
			count++
		}

		assert.Equal(t, reg.CalculateExpectedMaximalLength(), count)
	}
}

func TestNonIntervals4Bits(t *testing.T) {
	t.Parallel()
	intervals := Intervals4Bits()

	for i := 1; i <= MaxUint4; i++ {
		if slices.Contains(intervals, i) {
			continue
		}

		reg := NewSSLFSR4(uint8(i))
		start := reg.register

		reg.Next()
		count := 1
		for reg.register != start || reg.counter != 0 {
			reg.Next()
			count++
		}

		assert.NotEqual(t, reg.CalculateExpectedMaximalLength(), count, "interval %d", i)
	}
}
