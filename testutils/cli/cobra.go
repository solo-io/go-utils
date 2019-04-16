package cli

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func ExecuteCli(command *cobra.Command, args string) error {
	command.SetArgs(strings.Split(args, " "))
	return command.Execute()
}

func ExecuteCliOut(command *cobra.Command, args string) (string, error) {
	stdOut := os.Stdout
	stdErr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w
	os.Stderr = w

	command.SetArgs(strings.Split(args, " "))
	err = command.Execute()

	outC := make(chan string)

	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	os.Stdout = stdOut // restoring the real stdout
	os.Stderr = stdErr
	out := <-outC

	return strings.TrimSuffix(out, "\n"), err
}
