package utils

import (
	"strconv"
	"time"
)

// UintToInt converts Uint value to Int value
func UintToInt(value uint) int {
	return int(value)
}

// IntToString converts Int to String
func IntToString(value int) string {
	return strconv.Itoa(value)
}

// UintToString converts Uint to String
func UintToString(value uint) string {
	return IntToString(UintToInt(value))
}

// IntToDuration converts value from Int to Duration
func IntToDuration(value int, multiplier time.Duration) time.Duration {
	return time.Duration(value) * multiplier
}

// UintToDuration converts value from Uint to Duration
func UintToDuration(value uint, multiplier time.Duration) time.Duration {
	return IntToDuration(UintToInt(value), multiplier)
}

// IntToDurationString converts value from Int to Duration string
func IntToDurationString(value int, multiplier time.Duration) string {
	return IntToDuration(value, multiplier).String()
}

// UintToDurationString converts value from Uint to Duration string
func UintToDurationString(value uint, multiplier time.Duration) string {
	return UintToDuration(value, multiplier).String()
}
