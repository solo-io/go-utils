package logger

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"time"

	"github.com/k0kubun/pp"
)

var debugMode = os.Getenv("DEBUG") == "1"

// Deprecated:
var DefaultOut io.Writer = os.Stdout

func init() {
	if os.Getenv("DISABLE_COLOR") == "1" {
		pp.ColoringEnabled = false
	}
}

var rxp = regexp.MustCompile(".*/src/")

// Deprecated:
func Sprintf(format string, a ...interface{}) string {
	return pp.Sprintf("%v\t"+format, append([]interface{}{line()}, a...)...)
}

// Deprecated:
func GreyPrintf(format string, a ...interface{}) {
	fmt.Fprintf(DefaultOut, "%v\t"+format+"\n", append([]interface{}{line()}, a...)...)
}

// Deprecated:
func Printf(format string, a ...interface{}) {
	fmt.Fprintln(DefaultOut, Sprintf(format, a...))
}

// Deprecated:
func Warnf(format string, a ...interface{}) {
	fmt.Fprintln(DefaultOut, Sprintf("WARNING:\n"+format+"\n", a...))
}

// Deprecated:
func Debugf(format string, a ...interface{}) {
	if debugMode {
		fmt.Fprintln(DefaultOut, Sprintf(format, a...))
	}
}

// Deprecated:
func Fatalf(format string, a ...interface{}) {
	fmt.Fprintln(DefaultOut, Sprintf(format, a...))
	os.Exit(1)
}

func line() string {
	_, file, line, _ := runtime.Caller(3)
	file = rxp.ReplaceAllString(file, "")
	return fmt.Sprintf("%v: %v:%v", time.Now().Format(time.RFC1123), file, line)
}
