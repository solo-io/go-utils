package benchmarking

import (
	"fmt"
	"runtime"
	"syscall"
	"time"

	. "github.com/onsi/gomega"
)

// ONLY COMPILES ON LINUX
// most often you will want to wrap the code being tested (i.e., f()) in a for loop to get more sizeable readings.
// use this function over built-in benchmarker (https://onsi.github.io/ginkgo/#benchmark_tests) to measure time
// using the OS user-time specific to the executing thread. This helps with noisy neighbors in ci/locally
func ExpectFuncToComplete(f func(), runtimeThresholdInSeconds float64) {
	var rusage1 syscall.Rusage
	var rusage2 syscall.Rusage
	runtime.LockOSThread()                                    // important to lock OS thread to ensure we are only one goroutine being benchmarked
	err := syscall.Getrusage(syscall.RUSAGE_THREAD, &rusage1) // RUSAGE_THREAD system call only works/compiles on linux
	if err != nil {
		Expect(err).ToNot(HaveOccurred())
	}
	f()
	err = syscall.Getrusage(syscall.RUSAGE_THREAD, &rusage2)
	if err != nil {
		Expect(err).ToNot(HaveOccurred())
	}
	runtime.UnlockOSThread()
	duration1 := time.Duration(rusage1.Utime.Nano())
	duration2 := time.Duration(rusage2.Utime.Nano())
	userRuntime := duration2 - duration1
	fmt.Printf("\nutime: %f\n", userRuntime.Seconds())
	Expect(userRuntime.Seconds()).Should(And(
		BeNumerically("<=", runtimeThresholdInSeconds),
		BeNumerically(">", 0)))
}
