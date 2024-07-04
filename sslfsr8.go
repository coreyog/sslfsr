package sslfsr

import (
	"math"
	"math/bits"
)

// SSLFSR8 holds an 8 bit register, it's interval, and a counter
type SSLFSR8 struct {
	register uint8
	interval uint8
	counter  uint8
}

// Intervals8Bits returns a list of known optimum intervals
func Intervals8Bits() (working []int) {
	return []int{
		1, //?
		11,
		29,
		63,
		68,
		83,
		104,
		106,
		129,
		134,
		150,
		155,
		170,
		176,
		177,
		192,
		195,
		202,
		225,
		237,
		253,
	}
}

// NewSSLFSR8 constructs an SSLFSR8 with a given interval
func NewSSLFSR8(interval uint8) (sslfsr SSLFSR8) {
	return SSLFSR8{
		register: 1,
		interval: interval,
		counter:  0,
	}
}

// BuildSSLFSR8 constructs an SSLFSR8 with a given register, interval, and counter
func BuildSSLFSR8(register uint8, interval uint8, counter uint8) (sslfsr SSLFSR4) {
	return SSLFSR4{
		register: register,
		interval: interval,
		counter:  counter,
	}
}

// GetRegister returns the current register value
func (sslfsr *SSLFSR8) GetRegister() uint8 {
	return sslfsr.register
}

// GetInterval returns the interval this SSLFSR8 was constructed with
func (sslfsr *SSLFSR8) GetInterval() uint8 {
	return sslfsr.interval
}

// GetCounter returns the current counter value
func (sslfsr *SSLFSR8) GetCounter() uint8 {
	return sslfsr.counter
}

// Next Shifts or SubShifts according to the Counter and Interval and updates Counter accordingly
func (sslfsr *SSLFSR8) Next() {
	if sslfsr.counter == sslfsr.interval {
		sslfsr.SubShift()
		sslfsr.counter = 0
	} else {
		sslfsr.Shift()
		sslfsr.counter++
	}
}

// Shift modifies register by applying a standard LFSR shift to it
func (sslfsr *SSLFSR8) Shift() {
	sslfsr.register = Shift8Bits(sslfsr.register)
}

// Shift modifies register by applying a standard LFSR shift to it
func Shift8Bits(register uint8) uint8 {
	taps := byte(0b00011101)
	bit := bits.OnesCount8(register&taps)%2 == 1

	register = register >> 1
	if bit {
		register = register | 0x80
	}

	return register
}

// SubShift modifies register by applying a standard LFSR shift to just it's lower bits
func (sslfsr *SSLFSR8) SubShift() {
	sslfsr.register = SubShift8Bits(sslfsr.register)
}

// SubShift modifies register by applying a standard LFSR shift to just it's lower bits
func SubShift8Bits(register uint8) uint8 {
	taps := byte(0b00000011)
	bit := bits.OnesCount8(register&taps)%2 == 1
	higher := register & 0xF0
	lower := register & 0x0F
	lower = lower >> 1
	if bit {
		lower = lower | 0x8
	}

	return lower | higher
}

// CalculateExpectedMaximalLength calculates the total state count if the SSLFSRs Interval were an optimal Interval
func (sslfsr *SSLFSR8) CalculateExpectedMaximalLength() (stateCount int) {
	return CalculateExpectedMaximalLength8Bits(sslfsr.interval)
}

// CalculateExpectedMaximalLength8Bits calculates the total state count if the SSLFSRs Interval were an optimal Interval
func CalculateExpectedMaximalLength8Bits(interval uint8) (stateCount int) {
	return math.MaxUint8 * (int(interval) + 1)
}
