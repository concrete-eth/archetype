package utils

import (
	"math"

	"golang.org/x/exp/constraints"
)

// Min returns the smaller of a and b.
func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max returns the larger of a and b.
func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Abs returns the absolute value of x.
func Abs[T constraints.Signed](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

// Clamp returns x clamped to the range [floor, ceil].
func Clamp[T constraints.Ordered](x, floor, ceil T) T {
	return Min(Max(x, floor), ceil)
}

// Sign returns -1 if x < 0, 1 if x > 0, and 0 if x == 0.
func Sign[T constraints.Signed](x T) T {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}

// SafeAddInt8 adds a and b, returning math.MaxInt8 if the result would be greater than math.MaxInt8.
func SafeAddUint8(a, b uint8) uint8 {
	if a > math.MaxUint8-b {
		return math.MaxUint8
	}
	return a + b
}

// SafeSubInt8 subtracts b from a, returning 0 if the result would be less than 0.
func SafeSubUint8(a, b uint8) uint8 {
	if a < b {
		return 0
	}
	return a - b
}

// SafeAddInt16 adds a and b, returning math.MaxInt16 if the result would be greater than math.MaxInt16.
func SafeAddUint16(a, b uint16) uint16 {
	if a > math.MaxUint16-b {
		return math.MaxUint16
	}
	return a + b
}

// SafeSubInt16 subtracts b from a, returning 0 if the result would be less than 0.
func SafeSubUint16(a, b uint16) uint16 {
	if a < b {
		return 0
	}
	return a - b
}

// Pow returns x**y.
func Pow[T constraints.Signed](x, y T) T {
	return T(math.Pow(float64(x), float64(y)))
}
