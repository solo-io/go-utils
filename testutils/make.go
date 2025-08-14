package testutils

import (
	"fmt"
	"os/exec"
	"strings"

	. "github.com/onsi/gomega"
)

func MustMake(dir, args string) {
	make := exec.Command("make", strings.Split(args, " ")...)
	make.Dir = dir
	out, err := make.CombinedOutput()
	if err != nil {
		fmt.Printf("%s", out)
	}
	Expect(err).NotTo(HaveOccurred())
}
