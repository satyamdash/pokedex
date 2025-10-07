package main

import (
	"strings"
)

func cleanInput(text string) []string {
	lowercase := strings.ToLower(text)
	return strings.Fields(lowercase)
}
