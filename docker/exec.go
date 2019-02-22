package docker

import (
	"io"
	"os/exec"
)


func Command(args ...string) *containerCmd {
	return &containerCmd{
		args:     args,
	}
}

// containerCmd implements exec.Cmd for docker containers
type containerCmd struct {
	nameOrID string // the container name or ID
	args     []string
	env      []string
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
}

func (c *containerCmd) Run() error {
	var args []string

	// set env
	for _, env := range c.env {
		args = append(args, "-e", env)
	}
	args = append(
		args,
		// finally, with the caller args
		c.args...,
	)

	cmd := exec.Command("docker", args...)
	if c.stdin != nil {
		cmd.Stdin = c.stdin
	}
	if c.stderr != nil {
		cmd.Stderr = c.stderr
	}
	if c.stdout != nil {
		cmd.Stdout = c.stdout
	}
	return cmd.Run()
}
