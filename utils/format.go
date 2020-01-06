package utils

import (
	"strings"
)

func GetpqosFormat(lines []string) []float64 {
	data := make([]float64, 5)
	for _, line := range lines {
		if line == "" {
			return nil
		} else {
			nums := strings.Fields(line)[2:]
			nums[1] = nums[1][:len(nums[1])-1]
			//fmt.Println(nums)

			for i, num := range nums {
				data[i] += StringToFloat64(num)
			}
		}
	}
	return data
}
