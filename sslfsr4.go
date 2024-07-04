package sslfsr

import "math/bits"

const MaxUint4 = 1<<4 - 1

// Intervals4Bits returns a list of known optimum intervals
func Intervals4Bits() (working []int) {
	return []int{
		1,
		7,
		11,
		13,
	}
}

// SSLFSR4 manages a 4 bit register
type SSLFSR4 struct {
	register uint8
	interval uint8
	counter  uint8
}

// NewSSLFSR4 constructs an SSLFSR4 with a given interval
func NewSSLFSR4(interval uint8) (sslfsr SSLFSR4) {
	return SSLFSR4{
		register: 1,
		interval: interval,
		counter:  0,
	}
}

// BuildSSLFSR4 constructs an SSLFSR4 with a given register, interval, and counter
func BuildSSLFSR4(register uint8, interval uint8, counter uint8) (sslfsr SSLFSR4) {
	return SSLFSR4{
		register: register,
		interval: interval,
		counter:  counter,
	}
}

// GetRegister returns the current register value
func (sslfsr *SSLFSR4) GetRegister() uint8 {
	return sslfsr.register
}

// GetInterval returns the interval this SSLFSR4 was constructed with
func (sslfsr *SSLFSR4) GetInterval() uint8 {
	return sslfsr.interval
}

// GetCounter returns the current counter value
func (sslfsr *SSLFSR4) GetCounter() uint8 {
	return sslfsr.counter
}

// Next Shifts or SubShifts according to the Counter and Interval and updates Counter accordingly
func (sslfsr *SSLFSR4) Next() {
	if sslfsr.counter == sslfsr.interval {
		sslfsr.SubShift()
		sslfsr.counter = 0
	} else {
		sslfsr.Shift()
		sslfsr.counter++
	}
}

// Shift modifies register by applying a standard LFSR shift to it
func (sslfsr *SSLFSR4) Shift() {
	sslfsr.register = Shift4Bits(sslfsr.register)
}

// Shift modifies register by applying a standard LFSR shift to it
func Shift4Bits(register uint8) uint8 {
	taps := byte(0b0011)
	bit := bits.OnesCount8(register&taps)%2 == 1

	register = register >> 1
	if bit {
		register = register | 0b1000
	}

	return register
}

// SubShift modifies register by applying a standard LFSR shift to just it's lower bits
func (sslfsr *SSLFSR4) SubShift() {
	sslfsr.register = SubShift4Bits(sslfsr.register)
}

// SubShift modifies register by applying a standard LFSR shift to just it's lower bits
func SubShift4Bits(register uint8) uint8 {
	taps := byte(0b0011)
	bit := bits.OnesCount8(register&taps)%2 == 1

	higher := register & 0x0C
	lower := register & 0x03

	lower = lower >> 1
	if bit {
		lower = lower | 0x02
	}

	register = higher | lower

	return register
}

// CalculateExpectedMaximalLength calculates the total state count if the SSLFSRs Interval were an optimal Interval
func (sslfsr *SSLFSR4) CalculateExpectedMaximalLength() (stateCount int) {
	return CalculateExpectedMaximalLength4Bits(sslfsr.interval)
}

// CalculateExpectedMaximalLength4Bits calculates the total state count if the SSLFSRs Interval were an optimal Interval
func CalculateExpectedMaximalLength4Bits(interval uint8) (stateCount int) {
	return 15 * (int(interval) + 1) // (2^4-1)*(interval+1)
}
