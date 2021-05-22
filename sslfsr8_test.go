package sslfsr

import (
	"math"
	"testing"
)

func Test8BitShift(t *testing.T) {
	x := NewSSLFSR8(0)
	start := x.register
	for i := 0; i < 255; i++ {
		x.Shift()
	}
	if x.register != start {
		t.Error("254 shifts should result in starting state")
	}
}

func Test8BitSubShift(t *testing.T) {
	x := NewSSLFSR8(0)
	start := x.register
	for i := 0; i < 15; i++ {
		x.SubShift()
	}
	if x.register != start {
		t.Error("14 subshifts should result in starting state")
	}
}

func Test8BitNext(t *testing.T) {
	x := NewSSLFSR8(11)
	start := x.register

	x.Next()
	count := 1
	for x.register != start || x.counter != 0 {
		x.Next()
		count++
	}
	expected := (math.Pow(2, 8) - 1) * 12
	if float64(count) != expected {
		t.Error("should return to start state in ((2^8)-1)(11+1) state changes, expected:", expected, "actual:", count)
	}
}

func TestIntervals8Bits(t *testing.T) {
	intervals := Intervals8Bits()
	for _, interval := range intervals {
		x := NewSSLFSR8(interval)
		start := x.register
		x.Next()
		count := 1
		for x.register != start || x.counter != 0 {
			x.Next()
			count++
		}
		expected := x.CalculateExpectedMaximalLength()
		if count != expected {
			t.Error("interval", interval, "does not work, expected state count:", expected, "actual:", count)
		}
	}
}

func TestNonIntervals8Bits(t *testing.T) {
	intervals := Intervals8Bits()
	for i := 2; i <= 255; i++ {
		if contains(intervals, i) {
			continue
		}
		x := NewSSLFSR8(i)
		start := x.register
		x.Next()
		count := 1
		for x.register != start || x.counter != 0 {
			x.Next()
			count++
		}
		expected := x.CalculateExpectedMaximalLength()
		if count == expected {
			t.Error("interval", i, "works, state count should not be:", expected, "actual:", count)
		}
	}
}
