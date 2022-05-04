package utils

import (
	"github.com/davecgh/go-spew/spew"
	"os"
)

func D(values ...interface{}) {
	spew.Dump(values...)
}

func DD(values ...interface{}) {
	D(values...)
	os.Exit(0)
}
