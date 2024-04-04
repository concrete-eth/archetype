package utils

import (
	"math"

	"golang.org/x/exp/constraints"
)

func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func Abs[T constraints.Signed](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

func Clamp[T constraints.Ordered](x, floor, ceil T) T {
	return Min(Max(x, floor), ceil)
}

func Sign[T constraints.Signed](x T) T {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}

func SafeAddUint8(a, b uint8) uint8 {
	if a > math.MaxUint8-b {
		return math.MaxUint8
	}
	return a + b
}

func SafeSubUint8(a, b uint8) uint8 {
	if a < b {
		return 0
	}
	return a - b
}

func SafeAddUint16(a, b uint16) uint16 {
	if a > math.MaxUint16-b {
		return math.MaxUint16
	}
	return a + b
}

func SafeSubUint16(a, b uint16) uint16 {
	if a < b {
		return 0
	}
	return a - b
}

func Pow[T constraints.Signed](x, y T) T {
	return T(math.Pow(float64(x), float64(y)))
}
