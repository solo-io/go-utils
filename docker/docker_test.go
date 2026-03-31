package docker_test

import (
	"context"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/docker"
)

const (
	validImage   = "soloio/gloo:0.7.0"
	invalidImage = "soloio/gloo:0.1.0"
)

var _ = BeforeSuite(func() {
	exec.Command("docker", "image", "rm", validImage).Run()
})

var _ = Describe("Docker", func() {

	pullValidImage := func() {
		_, err := docker.PullIfNotPresent(context.Background(), validImage, 1)
		Expect(err).NotTo(HaveOccurred())
	}

	pullInvalidImage := func() {
		_, err := docker.PullIfNotPresent(context.Background(), invalidImage, 1)
		Expect(err).To(HaveOccurred())
	}

	Context("Pull", func() {
		It("can pull a valid container", func() {
			// caching in github actions leads to docker image being present from run to run and
			// failing the previous check.
			// previous: Expect(pullValidImage()).To(BeTrue())
			// docker.PullIfNotPresent returns a boolean to represent whether or not it pulled the image
			// and the caching stops it from being pulled
			pullValidImage()
		})

		It("cannot pull an invalid container", func() {
			pullInvalidImage()
		})

	})

	Context("Save", func() {
		It("can save a valid, present container", func() {
			pullValidImage()
			file, err := os.CreateTemp("", "docker_test")
			Expect(err).NotTo(HaveOccurred())
			err = docker.Save(validImage, file.Name())
			Expect(err).NotTo(HaveOccurred())
		})

		It("cannot save an invalid container", func() {
			file, err := os.CreateTemp("", "docker_test")
			Expect(err).NotTo(HaveOccurred())
			err = docker.Save(invalidImage, file.Name())
			Expect(err).To(HaveOccurred())
		})
	})

})
