package benchmarking

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	. "github.com/onsi/gomega"
)

// ONLY COMPILES ON LINUX
// most often you will want to wrap the code being tested (i.e., f()) in a for loop to get more sizeable readings.
// use this function over built-in benchmarker (https://onsi.github.io/ginkgo/#benchmark_tests) to measure time
// using the OS user-time specific to the executing thread. This _should_ help with noisy neighbors in ci/locally.
//
// other things to note while benchmarking:
//   - I've noticed the google cloudbuild hardware matters a lot, using nonstandard hardware (e.g. `N1_HIGHCPU_8`)
//     has helped improve reliability of benchmarks
//   - we could still explore running tests with nice and/or ionice
//   - could also further explore running the tests with docker flags --cpu-shares set
func ExpectFuncToComplete(f func(), runtimeThresholdInSeconds float64) {
	var rusage1 syscall.Rusage
	var rusage2 syscall.Rusage
	runtime.LockOSThread()                                    // important to lock OS thread to ensure we are the only goroutine being benchmarked
	prevGc := debug.SetGCPercent(-1)                          // might just be paranoid, but disable gc while benchmarking
	err := syscall.Getrusage(syscall.RUSAGE_THREAD, &rusage1) // RUSAGE_THREAD system call only works/compiles on linux
	if err != nil {
		Expect(err).ToNot(HaveOccurred())
	}
	f()
	err = syscall.Getrusage(syscall.RUSAGE_THREAD, &rusage2)
	if err != nil {
		Expect(err).ToNot(HaveOccurred())
	}
	debug.SetGCPercent(prevGc)
	runtime.UnlockOSThread()
	duration1 := time.Duration(rusage1.Utime.Nano())
	duration2 := time.Duration(rusage2.Utime.Nano())
	userRuntime := duration2 - duration1
	fmt.Printf("utime: %f\n", userRuntime.Seconds())
	Expect(userRuntime.Seconds()).Should(And(
		BeNumerically("<=", runtimeThresholdInSeconds),
		BeNumerically(">", 0)))
}

func TimeForFuncToComplete(f func()) float64{
	var rusage1 syscall.Rusage
	var rusage2 syscall.Rusage
	runtime.LockOSThread()                                    // important to lock OS thread to ensure we are the only goroutine being benchmarked
	prevGc := debug.SetGCPercent(-1)                          // might just be paranoid, but disable gc while benchmarking
	err := syscall.Getrusage(syscall.RUSAGE_THREAD, &rusage1) // RUSAGE_THREAD system call only works/compiles on linux
	if err != nil {
		Expect(err).ToNot(HaveOccurred())
	}
	f()
	err = syscall.Getrusage(syscall.RUSAGE_THREAD, &rusage2)
	if err != nil {
		Expect(err).ToNot(HaveOccurred())
	}
	debug.SetGCPercent(prevGc)
	runtime.UnlockOSThread()
	duration1 := time.Duration(rusage1.Utime.Nano())
	duration2 := time.Duration(rusage2.Utime.Nano())
	userRuntime := duration2 - duration1
	fmt.Printf("utime: %f\n", userRuntime.Seconds())
	Expect(userRuntime.Seconds()).Should(
		BeNumerically(">", 0))
	return userRuntime.Seconds()
}
