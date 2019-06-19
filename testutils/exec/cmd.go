package exec

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"github.com/onsi/ginkgo"
	"github.com/solo-io/go-utils/errors"
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
	cmd := exec.Command(args[0], args[1:]...)
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
