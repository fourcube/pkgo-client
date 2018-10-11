package xlog

import (
	"fmt"
	"os"
)

func Fatal(s string, args ...interface{}) {
	fmt.Printf(fmt.Sprintf("FATAL %s\n", s), args...)
	os.Exit(1)
}

func Print(s string, args ...interface{}) {
	fmt.Printf(fmt.Sprintf("%s\n", s), args...)
}
