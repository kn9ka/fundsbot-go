package utils

import (
	"fmt"
	"strconv"
)

func ConvertRate(rate string) string {
	// Преобразование строки в число с плавающей точкой
	num, err := strconv.ParseFloat(rate, 64)
	if err != nil {
		fmt.Printf("Failed to parse float: %s, while parsing: %s", err, rate)
		return rate
	}

	// Округление числа до двух знаков после точки
	roundedNum := fmt.Sprintf("%.2f", num)

	return roundedNum
}
