// Write a program that takes a string as input and checks whether it is a palindrome.
package main

import (
	"fmt"
)

func palindrome(a string) bool {
	for i := 0; i < len(a)/2; i++ {
		if a[i] != a[len(a)-i-1] {
			return false
		}
	}
	return true
}

func main() {
	word := "noon"
	if palindrome(word) {
		fmt.Printf("%s is a palindrome.\n", word)
	} else {
		fmt.Printf("%s is not a palindrome.\n", word)
	}
}
