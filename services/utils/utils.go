package utils

import (
	"fmt"
	"strconv"
)

func ConvertRate(rate string) string {
	// Преобразование строки в число с плавающей точкой
	num, err := strconv.ParseFloat(rate, 64)
	if err != nil {
		fmt.Println("Failed to parse float:", err)
		return rate
	}

	// Округление числа до двух знаков после точки
	roundedNum := fmt.Sprintf("%.2f", num)

	return roundedNum
}
