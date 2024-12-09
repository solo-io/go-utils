package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"runtime"
	"time"

	"github.com/k0kubun/pp"
)

var debugMode = os.Getenv("DEBUG") == "1"
var timeFormat = time.RFC1123

var DefaultOut io.Writer = os.Stdout

func init() {
	if os.Getenv("DISABLE_COLOR") == "1" {
		pp.ColoringEnabled = false
	}
	if os.Getenv("LOG_TIME_FORMAT") != "" {
		timeFormat = os.Getenv("LOG_TIME_FORMAT")
		log.Printf("Setting time format to %v, this is global and discouraged in production formats", timeFormat)
	}
}

var rxp = regexp.MustCompile(".*/src/")

func Sprintf(format string, a ...interface{}) string {
	return pp.Sprintf("%v\t"+format, append([]interface{}{line()}, a...)...)
}

func GreyPrintf(format string, a ...interface{}) {
	fmt.Fprintf(DefaultOut, "%v\t"+format+"\n", append([]interface{}{line()}, a...)...)
}

func Printf(format string, a ...interface{}) {
	fmt.Fprintln(DefaultOut, Sprintf(format, a...))
}

func Warnf(format string, a ...interface{}) {
	fmt.Fprintln(DefaultOut, Sprintf("WARNING:\n"+format+"\n", a...))
}

func Debugf(format string, a ...interface{}) {
	if debugMode {
		fmt.Fprintln(DefaultOut, Sprintf(format, a...))
	}
}

func Fatalf(format string, a ...interface{}) {
	fmt.Fprintln(DefaultOut, Sprintf(format, a...))
	os.Exit(1)
}

func line() string {
	_, file, line, _ := runtime.Caller(3)
	file = rxp.ReplaceAllString(file, "")
	return fmt.Sprintf("%v: %v:%v", time.Now().Format(timeFormat), file, line)
}
