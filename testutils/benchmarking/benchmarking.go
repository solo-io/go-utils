package benchmarking

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"

	errors "github.com/rotisserie/eris"

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
//
// Deprecated: use Measure instead.
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

// TimeForFuncToComplete returns the time the given function spend executing in user mode.
// Deprecated: use Measure instead.
func TimeForFuncToComplete(f func()) float64 {
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

// Result represents the result of measuring a function's execution time.
type Result struct {
	// Time spent in user mode
	Utime time.Duration
	// Time spent in kernel mode
	Stime time.Duration
	// Time spent in user mode + kernel mode
	Total time.Duration
}

// Measure returns the time it took to execute the given function. It only compiles on Linux.
// Most often you will want to run the code you want to test in a loop to get more sizeable readings, e.g.:
//
//	results := benchmarking.Measure(func() {
//		for i := 0; i < 100; i++ {
//			funcToTest()
//		}
//	})
//
// Measure should be preferred over the Gomega benchmark utils (https://pkg.go.dev/github.com/onsi/gomega/gmeasure)
// as it takes some additional steps to ensure we get accurate measurements.
//
// Further ideas for improvement:
//   - we could explore running tests with nice and/or ionice
//   - could also further explore running the tests with docker flags --cpu-shares set
//   - consider setting GOMAXPROCS when running the tests to ensure we run in a single thread
func Measure(f func()) (Result, error) {

	before, after, err := doMeasure(f)
	if err != nil {
		return Result{}, err
	}

	res := Result{
		Utime: time.Duration(after.Utime.Nano() - before.Utime.Nano()),
		Stime: time.Duration(after.Stime.Nano() - before.Stime.Nano()),
		Total: time.Duration(after.Utime.Nano() + after.Stime.Nano() - before.Utime.Nano() - before.Stime.Nano()),
	}

	// Time spent in user + kernel modes can't reasonably be 0ns, so err on the side of caution and fail.
	if res.Total == 0 {
		return Result{}, errors.New("total execution time was 0 ns")
	}

	return res, nil
}

func doMeasure(f func()) (before syscall.Rusage, after syscall.Rusage, err error) {

	// Important: lock OS thread to ensure we are the only goroutine being benchmarked
	runtime.LockOSThread()

	// Might just be paranoid, but disable garbage collection while benchmarking
	prevGc := debug.SetGCPercent(-1)

	defer func() {
		debug.SetGCPercent(prevGc)
		runtime.UnlockOSThread()
	}()

	// getrusage system call only works/compiles on linux
	if err = syscall.Getrusage(syscall.RUSAGE_THREAD, &before); err != nil {
		return
	}

	f()

	err = syscall.Getrusage(syscall.RUSAGE_THREAD, &after)
	return
}
