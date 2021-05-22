package sslfsr

// Intervals4Bits returns a list of known optimum intervals
func Intervals4Bits() (working []int) {
	return []int{
		7,
		11,
		13,
	}
}

// SSLFSR4 holds an 8 bit register
type SSLFSR4 struct {
	register byte
	interval int
	counter  int
}

// NewSSLFSR4 constructs an SSLFSR8 with a given interval
func NewSSLFSR4(interval int) (sslfsr SSLFSR4) {
	return SSLFSR4{
		register: 1,
		interval: interval,
		counter:  0,
	}
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

// Shift modifies Register by applying a standard LFSR shift to it
func (sslfsr *SSLFSR4) Shift() {
	reg := sslfsr.register
	bit := GetBit(reg, 0) != GetBit(reg, 1)
	reg = reg >> 1
	if bit {
		reg = reg | 0x08
	}
	sslfsr.register = reg
}

// SubShift modifies Register by applying a standard LFSR shift to just it's lower bits
func (sslfsr *SSLFSR4) SubShift() {
	reg := sslfsr.register
	bit := GetBit(reg, 0) != GetBit(reg, 1)
	higher := reg & 0x0C
	lower := reg & 0x03
	lower = lower >> 1
	if bit {
		lower = lower | 0x02
	}
	sslfsr.register = lower | higher
}

// CalculateExpectedMaximalLength calculates the total state count if the SSLFSRs Interval were an optimal Interval
func (sslfsr *SSLFSR4) CalculateExpectedMaximalLength() (stateCount int) {
	return 15 * (sslfsr.interval + 1)
}
