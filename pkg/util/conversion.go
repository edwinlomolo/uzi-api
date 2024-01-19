package util

import "strconv"

func FloatToString(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}
