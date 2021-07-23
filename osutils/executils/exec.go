package executils

import (
	"os/exec"
	"syscall"
)

// Runs cmd.CombinedOutput and returns the status code of the run
func CombinedOutputWithStatus(cmd *exec.Cmd) ([]byte, int, error) {
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				return out, status.ExitStatus(), err
			}
		}
	}
	return out, 0, err
}
