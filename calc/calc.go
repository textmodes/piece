package calc

func DivMod(a, b int) (d, m int) {
	d = a / b
	m = a % b
	return
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
