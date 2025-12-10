package checks

import (
	"os"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/mod/modfile"
)

var _ = Describe("Checks", func() {

	It("should used forked klog instead of klog", func() {
		// regular klog writes to disk, so make sure we used a forked version that doesn't write to
		// disk, which is a problem with hardened containers with root only file systems.

		gomod, err := exec.Command("go", "env", "GOMOD").CombinedOutput()
		Expect(err).NotTo(HaveOccurred())
		gomodfile := strings.TrimSpace(string(gomod))
		data, err := os.ReadFile(gomodfile)
		Expect(err).NotTo(HaveOccurred())

		modFile, err := modfile.Parse(gomodfile, data, nil)
		Expect(err).NotTo(HaveOccurred())

		for _, dep := range modFile.Require {
			// let's not depend on these in go utils, as go utils is used in a lot of places
			// and we want to keep it lean
			Expect(dep.Mod.Path).NotTo(HavePrefix("k8s.io/"))
			Expect(dep.Mod.Path).NotTo(HavePrefix("sigs.k8s.io/controller-runtime/"))
		}

	})

})
