package util

import (
	"strconv"
	"unicode"
)

// ValidateLuhn checks if a string of digits passes the Luhn algorithm check
func ValidateLuhn(number string) bool {
	// Check if the number contains only digits
	for _, c := range number {
		if !unicode.IsDigit(c) {
			return false
		}
	}

	// Convert string to slice of integers
	var digits []int
	for _, c := range number {
		digit, _ := strconv.Atoi(string(c))
		digits = append(digits, digit)
	}

	// Calculate sum according to Luhn algorithm
	sum := 0
	parity := len(digits) % 2

	for i, digit := range digits {
		// Double every second digit starting from the right
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	// Number is valid if sum is divisible by 10
	return sum%10 == 0
}