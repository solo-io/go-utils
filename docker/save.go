

package docker

import (
	"sigs.k8s.io/kind/pkg/exec"
)

// Save saves image to dest, as in `docker save`
func Save(image, dest string) error {
	return exec.Command("docker", "save", "-o", dest, image).Run()
}
