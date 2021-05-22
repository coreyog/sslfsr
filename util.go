package sslfsr

// getBit checks a bit at 1<<loc and returns if that bit is a 1
func GetBit(x byte, loc uint) (bit bool) {
	return x&(1<<loc) != 0
}

func GetBit16(x uint16, loc uint) (bit bool) {
	return x&(1<<loc) != 0
}

// contains checks if an array of integers containes a specific integer
func contains(nums []int, num int) (itDoes bool) {
	for _, n := range nums {
		if num == n {
			return true
		}
	}
	return false
}
