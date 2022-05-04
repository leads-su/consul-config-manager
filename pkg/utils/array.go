package utils

import (
	"strings"
)

// StringsArrayContains checks if array of strings contains term
func StringsArrayContains(array []string, term string) bool {
	for _, s := range array {
		if strings.TrimSpace(s) == strings.TrimSpace(term) {
			return true
		}
	}
	return false
}
