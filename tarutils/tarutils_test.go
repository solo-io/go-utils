package tarutils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/tarutils"
	"github.com/spf13/afero"
)

var _ = Describe("tarutils", func() {
	Context("File system", func() {
		It("can tar/untar files", func() {
			fs := afero.NewOsFs()
			tmpDir := mustWriteTestDir(fs)
			mustAddTestFiles(tmpDir, fs)
			tmp, err := afero.TempFile(fs, "", "tar-zipped-file-")
			Expect(err).NotTo(HaveOccurred())
			err = tarutils.Tar(tmpDir, fs, tmp)
			Expect(err).NotTo(HaveOccurred())

			newTmpDir := mustWriteTestDir(fs)
			err = tarutils.Untar(newTmpDir, tmp.Name(), fs)
			Expect(err).NotTo(HaveOccurred())

			mustFindOriginalFiles(newTmpDir, fs)
		})
	})

	Context("Mem map", func() {
		It("can tar/untar files", func() {
			fs := afero.NewMemMapFs()
			tmpDir := mustWriteTestDir(fs)
			mustAddTestFiles(tmpDir, fs)
			tmp, err := afero.TempFile(fs, "", "tar-zipped-file-")
			Expect(err).NotTo(HaveOccurred())
			err = tarutils.Tar(tmpDir, fs, tmp)
			Expect(err).NotTo(HaveOccurred())

			newTmpDir := mustWriteTestDir(fs)
			err = tarutils.Untar(newTmpDir, tmp.Name(), fs)
			Expect(err).NotTo(HaveOccurred())

			mustFindOriginalFiles(newTmpDir, fs)
		})
	})
})

func mustFindOriginalFiles(newTmpDir string, fs afero.Fs) {
	files, err := afero.ReadDir(fs, newTmpDir)
	Expect(err).NotTo(HaveOccurred())
	for _, v := range files {
		if !v.IsDir() {
			_, err := afero.ReadFile(fs, newTmpDir+"/"+v.Name())
			Expect(err).NotTo(HaveOccurred())
		}
	}
}

func mustWriteTestDir(fs afero.Fs) string {
	tmpDir, err := afero.TempDir(fs, "", "tar-test-")
	Expect(err).NotTo(HaveOccurred())
	return tmpDir
}

func mustAddTestFiles(tmpdir string, fs afero.Fs) {
	file, err := afero.TempFile(fs, tmpdir, "tar-test-file-")
	Expect(err).NotTo(HaveOccurred())
	err = afero.WriteFile(fs, file.Name(), []byte("first file"), 0777)
	Expect(err).NotTo(HaveOccurred())
	file, err = afero.TempFile(fs, tmpdir, "tar-test-file-")
	Expect(err).NotTo(HaveOccurred())
	err = afero.WriteFile(fs, file.Name(), []byte("second file"), 0777)
	Expect(err).NotTo(HaveOccurred())
}
