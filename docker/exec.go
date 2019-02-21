

package docker

import (
	"io"

	"sigs.k8s.io/kind/pkg/exec"
)

type Cmder interface {
	// command, args..., just like os/exec.Cmd
	Command(string, ...string) Cmd
}

// Cmd abstracts over running a command somewhere, this is useful for testing
type Cmd interface {
	Run() error
	// Each entry should be of the form "key=value"
	SetEnv(...string)
	SetStdin(io.Reader)
	SetStdout(io.Writer)
	SetStderr(io.Writer)
}
// containerCmder implements exec.Cmder for docker containers
type containerCmder struct {
	nameOrID string
}

// ContainerCmder creates a new exec.Cmder against a docker container
func ContainerCmder(containerNameOrID string) exec.Cmder {
	return &containerCmder{
		nameOrID: containerNameOrID,
	}
}

func (c *containerCmder) Command(command string, args ...string) exec.Cmd {
	return &containerCmd{
		nameOrID: c.nameOrID,
		command:  command,
		args:     args,
	}
}

// containerCmd implements exec.Cmd for docker containers
type containerCmd struct {
	nameOrID string // the container name or ID
	command  string
	args     []string
	env      []string
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
}

func (c *containerCmd) Run() error {
	args := []string{
		"exec",
		// run with priviliges so we can remount etc..
		// this might not make sense in the most general sense, but it is
		// important to many kind commands
		"--privileged",
	}
	if c.stdin != nil {
		args = append(args,
			"-i", // interactive so we can supply input
		)
	}
	if c.stderr != nil || c.stdout != nil {
		args = append(args,
			"-t", // use a tty so we can get output
		)
	}
	// set env
	for _, env := range c.env {
		args = append(args, "-e", env)
	}
	// specify the container and command, after this everything will be
	// args the the command in the container rather than to docker
	args = append(
		args,
		c.nameOrID, // ... against the container
		c.command,  // with the command specified
	)
	args = append(
		args,
		// finally, with the caller args
		c.args...,
	)
	cmd := exec.Command("docker", args...)
	if c.stdin != nil {
		cmd.SetStdin(c.stdin)
	}
	if c.stderr != nil {
		cmd.SetStderr(c.stderr)
	}
	if c.stdout != nil {
		cmd.SetStdout(c.stdout)
	}
	return cmd.Run()
}

func (c *containerCmd) SetEnv(env ...string) {
	c.env = env
}

func (c *containerCmd) SetStdin(r io.Reader) {
	c.stdin = r
}

func (c *containerCmd) SetStdout(w io.Writer) {
	c.stdout = w
}

func (c *containerCmd) SetStderr(w io.Writer) {
	c.stderr = w
}
