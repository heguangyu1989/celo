package p

import (
	"testing"
)

func TestPrint(t *testing.T) {
	Error("Hello, kitty")
	Info("hello kitty")
	Success("hello kitty")
}
