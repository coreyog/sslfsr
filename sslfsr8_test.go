package sslfsr

import (
	"math"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSettersAndGetters8Bits(t *testing.T) {
	t.Parallel()

	reg := BuildSSLFSR8(1, 2, 3)

	assert.Equal(t, uint8(1), reg.GetRegister())
	assert.Equal(t, uint8(2), reg.GetInterval())
	assert.Equal(t, uint8(3), reg.GetCounter())
}

func Test8BitShift(t *testing.T) {
	t.Parallel()

	for i := range math.MaxUint8 + 1 {
		reg := NewSSLFSR8(0) // not using interval
		reg.register = uint8(i)

		for range math.MaxUint8 {
			reg.Shift()
		}

		assert.Equal(t, uint8(i), reg.register, "255 shifts should result in starting state")
	}
}

func Test8BitSubShift(t *testing.T) {
	t.Parallel()

	for i := range math.MaxUint8 + 1 {
		reg := NewSSLFSR8(0)
		reg.register = uint8(i)

		for range math.MaxUint8 {
			reg.SubShift()
		}

		assert.Equal(t, uint8(i), reg.register, "255 subshifts should result in starting state")
	}
}

func TestIntervals8Bits(t *testing.T) {
	t.Parallel()
	intervals := Intervals8Bits()

	for _, interval := range intervals {
		reg := NewSSLFSR8(uint8(interval))
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

func TestNonIntervals8Bits(t *testing.T) {
	t.Parallel()
	intervals := Intervals8Bits()

	for i := 1; i <= math.MaxUint8; i++ {
		if slices.Contains(intervals, i) {
			continue
		}

		reg := NewSSLFSR8(uint8(i))
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
