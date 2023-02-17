package main

import (
	"fmt"
	"strings"
)

func digits(name string) {
	// The string to be converted
	input := "rubin"

	// Map each letter to its corresponding digit on a telephone keypad
	mapping := map[rune]rune{
		'A': '2', 'B': '2', 'C': '2',
		'D': '3', 'E': '3', 'F': '3',
		'G': '4', 'H': '4', 'I': '4',
		'J': '5', 'K': '5', 'L': '5',
		'M': '6', 'N': '6', 'O': '6',
		'P': '7', 'Q': '7', 'R': '7', 'S': '7',
		'T': '8', 'U': '8', 'V': '8',
		'W': '9', 'X': '9', 'Y': '9', 'Z': '9',
	}

	// Convert the string to the corresponding digits
	output := make([]rune, 0, len(name))
	for _, char := range strings.ToUpper(input) {
		if digit, ok := mapping[char]; ok {
			output = append(output, digit)
		} else {
			output = append(output, char)
		}
	}

	// Print the result
	fmt.Println("Digits:", string(output))
}
