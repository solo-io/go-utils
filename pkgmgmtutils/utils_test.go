package pkgmgmtutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("package management utils", func() {
	It("can extract shas from build file artifacts", func() {
		testData := map[string]struct {
			filename   string
			sha        string
			binaryName string
		}{
			"darwin": {
				filename:   "glooctl-darwin-amd64.sha256",
				sha:        "5b741927c40f4a430bcf42905901e204e955a5a5a3a3bbb9e67680df3a188f99",
				binaryName: "glooctl-darwin-amd64",
			},
			"linux": {
				filename:   "glooctl-linux-amd64.sha256",
				sha:        "9e173748ca85c5505909c4c7abe0cb03a58aef3b56dc134ece62d2306433347f",
				binaryName: "glooctl-linux-amd64",
			},
			"windows": {
				filename:   "glooctl-windows-amd64.sha256",
				sha:        "031434d831a394af2b7882b6f1a220e34efc91c4e4ef807a530fc8ec7990d2ca",
				binaryName: "glooctl-windows-amd64.exe",
			},
		}

		dirTmp, err := ioutil.TempDir("", "_output")
		Expect(err).To(BeNil())
		defer os.RemoveAll(dirTmp)

		for _, v := range testData {
			data := fmt.Sprintf("%s %s", v.sha, v.binaryName)
			err = ioutil.WriteFile(filepath.Join(dirTmp, v.filename), []byte(data), 0644)
			Expect(err).To(BeNil())
		}

		shas, err := getLocalBinarySha256(dirTmp)
		Expect(err).To(BeNil())

		Expect(shas.darwinSha).To(Equal([]byte(testData["darwin"].sha)))
		Expect(shas.linuxSha).To(Equal([]byte(testData["linux"].sha)))
		Expect(shas.windowsSha).To(Equal([]byte(testData["windows"].sha)))
	})

	It("can no shas found when extract shas from build file artifacts", func() {
		testData := map[string]struct {
			filename   string
			sha        string
			binaryName string
		}{
			"vms": {
				filename:   "glooctl-vax-vax.sha256",
				sha:        "5b741927c40f4a430bcf42905901e204e955a5a5a3a3bbb9e67680df3a188f99",
				binaryName: "glooctl-vms-vax",
			},
			"as400": {
				filename:   "glooctl-as400-i570.sha256",
				sha:        "9e173748ca85c5505909c4c7abe0cb03a58aef3b56dc134ece62d2306433347f",
				binaryName: "glooctl-as400-i570",
			},
			"hpux": {
				filename:   "glooctl-windows-sx1000.sha256",
				sha:        "031434d831a394af2b7882b6f1a220e34efc91c4e4ef807a530fc8ec7990d2ca",
				binaryName: "glooctl-windows-sx1000.exe",
			},
		}

		dirTmp, err := ioutil.TempDir("", "_output")
		Expect(err).To(BeNil())
		defer os.RemoveAll(dirTmp)

		for _, v := range testData {
			data := fmt.Sprintf("%s %s", v.sha, v.binaryName)
			err = ioutil.WriteFile(filepath.Join(dirTmp, v.filename), []byte(data), 0644)
			Expect(err).To(BeNil())
		}

		shas, err := getLocalBinarySha256(dirTmp)
		Expect(err).To(Equal(ErrNoSha256sFound))
		Expect(shas).To(BeNil())
	})

	It("can extract sha from a file", func() {
		testSha := "5b741927c40f4a430bcf42905901e204e955a5a5a3a3bbb9e67680df3a188f99"
		testData := testSha + " glooctl-darwin-amd64"

		file, err := ioutil.TempFile("", "glooctl-darwin-amd64.sha256")
		Expect(err).To(BeNil())
		defer os.Remove(file.Name())

		err = ioutil.WriteFile(file.Name(), []byte(testData), 0644)
		Expect(err).To(BeNil())

		b, err := extractShaFromFile(file.Name())
		Expect(err).To(BeNil())
		Expect(b).To(Equal([]byte(testSha)))
	})

	It("fail when extract sha from an empty file", func() {
		file, err := ioutil.TempFile("", "glooctl-darwin-amd64.sha256")
		Expect(err).To(BeNil())
		defer os.Remove(file.Name())

		b, err := extractShaFromFile(file.Name())
		Expect(err).To(Equal(ErrNoShaDataFound))
		Expect(b).To(BeEmpty())
	})

	It("can extract sha from a file", func() {
		testData := "Some Random Test String"

		file, err := ioutil.TempFile("", "glooctl-darwin-amd64.sha256")
		Expect(err).To(BeNil())
		defer os.Remove(file.Name())

		err = ioutil.WriteFile(file.Name(), []byte(testData), 0644)
		Expect(err).To(BeNil())

		b, err := extractShaFromFile(file.Name())
		Expect(err).ToNot(BeNil())
		Expect(b).To(BeNil())
	})

	It("can replace submatch", func() {
		regex := `version\s*"([0-9.]+)"`
		testData := `Some other data version "1.2.3" and even more data`

		b := replaceSubmatch([]byte(testData), []byte("4.5.6"), regexp.MustCompile(regex))

		Expect(b).To(Equal([]byte(`Some other data version "4.5.6" and even more data`)))
	})

	It("can replace submatch fail", func() {
		regex := `version\s*"([a-z.]+)"`
		testData := `Some other data version "1.2.3" and version "4.5.6" and even more data`

		b := replaceSubmatch([]byte(testData), []byte("7.8.9"), regexp.MustCompile(regex))
		Expect(b).To(Equal([]byte(testData)))
	})

	It("can replace submatch fail", func() {
		regex := `version\s*"[0-9.]+"`
		testData := `Some other data version "1.2.3" and version "4.5.6" and even more data`

		Expect(func() {
			_ = replaceSubmatch([]byte(testData), []byte("7.8.9"), regexp.MustCompile(regex))
		}).To(Panic())
	})
})
