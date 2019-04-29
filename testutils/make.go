package testutils

import (
	"os/exec"
	"strings"

	. "github.com/onsi/gomega"
)

func MustMake(dir, args string) {
	make := exec.Command("make", strings.Split(args, " ")...)
	make.Dir = dir
	err := make.Run()
	Expect(err).NotTo(HaveOccurred())
}
