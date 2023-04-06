package exec

import (
	"bytes"
	"context"
	"github.com/onsi/ginkgo"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"

	"github.com/onsi/ginkgo/v2"
	"github.com/pkg/errors"
)

func RunCommand(workingDir string, verbose bool, args ...string) error {
	_, err := RunCommandOutput(workingDir, verbose, args...)
	return err
}

func RunCommandOutput(workingDir string, verbose bool, args ...string) (string, error) {
	return RunCommandInputOutput("", workingDir, verbose, args...)
}

func RunCommandInput(input, workingDir string, verbose bool, args ...string) error {
	_, err := RunCommandInputOutput(input, workingDir, verbose, args...)
	return err
}

func RunCommandInputOutput(input, workingDir string, verbose bool, args ...string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = workingDir
	cmd.Env = os.Environ()
	if len(input) > 0 {
		cmd.Stdin = bytes.NewBuffer([]byte(input))
	}
	buf := &bytes.Buffer{}
	var out io.Writer
	if verbose {
		out = io.MultiWriter(buf, ginkgo.GinkgoWriter)
	} else {
		out = buf
	}
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		return "", errors.Wrapf(err, "%v failed: %s", cmd.Args, buf.String())
	}

	return buf.String(), nil
}
