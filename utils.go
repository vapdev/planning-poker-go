package main

import (
	"math/rand"
)

func generateRandomString(n int) string {
	var letters = []rune("1234567890")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
