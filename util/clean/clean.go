// Package cleanup provides a single point for registering clean up functions.
// This is similar to Finalizer, except that the clean up functions are
// guaranteed to be called if the process terminates normally.
//
// Usage:
//
// In mypackage.go
//
//	cleanup.Register(func(){
//	  // Arbitrary clean up function, most likely close goroutine, etc.
//	})
//
// In main.go
//
//	func main() {
//	  flag.Parse()
//	  defer cleanup.Run()
//	}
package clean

import (
	"sync"

	"github.com/golang/glog"
)

var (
	mu  sync.Mutex
	fns []func()
)

// Register adds a function to the cleanup queue.
func Register(f func()) {
	mu.Lock()
	defer mu.Unlock()
	fns = append(fns, f)
}

// Run runs all the cleanup functions registered.
func Run() {
	glog.Infof("Cleanup: performing %d cleanups", len(fns))
	mu.Lock()
	cur := fns // make sure the functions are executed precisely once.
	fns = nil
	mu.Unlock()
	for _, f := range cur {
		f()
	}
	glog.Infof("Cleanup: all cleanup done.")
	// Make sure all the logs produced during clean up is written to file.
	glog.Flush()
}
