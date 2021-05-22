package sslfsr

import (
	"math"
	"testing"
)

func Test4BitShift(t *testing.T) {
	x := NewSSLFSR4(0)
	start := x.register
	for i := 0; i < 15; i++ {
		x.Shift()
	}
	if x.register != start {
		t.Error("15 shifts should result in starting state")
	}
}

func Test4BitSubShift(t *testing.T) {
	x := NewSSLFSR4(0)
	start := x.register
	for i := 0; i < 3; i++ {
		x.SubShift()
		t.Log(x.register)
	}
	if x.register != start {
		t.Error("4 subshifts should result in starting state")
	}
}

func Test4BitNext(t *testing.T) {
	x := NewSSLFSR4(7)
	start := x.register

	x.Next()
	count := 1
	for x.register != start || x.counter != 0 {
		x.Next()
		count++
	}
	expected := (math.Pow(2, 4) - 1) * 8
	if float64(count) != expected {
		t.Error("should return to start state in ((2^4)-1)(7+1) state changes, expected:", expected, "actual:", count)
	}
}

func TestIntervals4Bits(t *testing.T) {
	intervals := Intervals4Bits()
	for _, interval := range intervals {
		x := NewSSLFSR4(interval)
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

func TestNonIntervals4Bits(t *testing.T) {
	intervals := Intervals4Bits()
	for i := 2; i <= 15; i++ {
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
