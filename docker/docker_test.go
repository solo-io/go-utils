package docker_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/docker"
	"io/ioutil"
	"os/exec"
)

const (
	validImage   = "soloio/gloo:0.7.0"
	invalidImage = "soloio/gloo:0.1.0"
)

var _ = BeforeSuite(func() {
	exec.Command("docker", "image", "prune", "-af").Run()
})

var _ = Describe("Docker", func() {

	pullValidImage := func() bool {
		ok, err := docker.PullIfNotPresent(validImage,1)
		Expect(err).NotTo(HaveOccurred())
		return ok
	}

	pullInvalidImage := func() {
		_, err := docker.PullIfNotPresent(invalidImage,1)
		Expect(err).To(HaveOccurred())
	}

	Context("Pull", func() {
		It("can pull a valid container", func() {
			Expect(pullValidImage()).To(Equal(true))
			Expect(pullValidImage()).To(Equal(false))
		})

		It("cannot pull an invalid container", func() {
			pullInvalidImage()
		})

	})

	Context("Save", func() {
		It("can save a valid, present container", func() {
			pullValidImage()
			file, err := ioutil.TempFile("", "docker_test")
			Expect(err).NotTo(HaveOccurred())
			err = docker.Save(validImage, file.Name())
			Expect(err).NotTo(HaveOccurred())
		})

		It("cannot save an invalid container", func() {
			file, err := ioutil.TempFile("", "docker_test")
			Expect(err).NotTo(HaveOccurred())
			err = docker.Save(invalidImage, file.Name())
			Expect(err).To(HaveOccurred())
		})
	})

})
