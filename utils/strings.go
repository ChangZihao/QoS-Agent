package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func StringToFloat64(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err != nil {
		return 0
	} else {
		return f
	}
}

func StrList2lines(strlist []string) string {
	if len(strlist) > 0 {
		return strings.Join(strlist, "\\n")
	} else {
		return ""
	}
}

func UInt2BitsStr(mask uint, maxWay int) string {
	width := maxWay / 4
	return fmt.Sprintf("%0[1]*x", width, mask)

}