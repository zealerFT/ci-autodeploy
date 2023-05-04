package clean

import (
	"testing"
)

func TestBasic(t *testing.T) {
	i, j := 0, 0
	func() {
		Register(func() { i += 1 })
		Register(func() { j += 2 })
		Run()
		Run() // multiple runs will be OK.
	}()
	if i != 1 || j != 2 {
		t.Errorf("Run() incorrect, want i = 1, j = 2, got i = %d, j = %d", i, j)
	}
}
