package test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/debugutils"
	"github.com/solo-io/go-utils/tarutils"
	"github.com/spf13/afero"
)

var _ = Describe("aggregator test", func() {
	var (
		aggregator *debugutils.Aggregator
		fs         afero.Fs
	)

	Context("e2e", func() {
		BeforeEach(func() {
			fs = afero.NewOsFs()
			storageClient := debugutils.NewFileStorageClient(fs)
			logCollector, err := debugutils.DefaultLogCollector()
			Expect(err).NotTo(HaveOccurred())
			resourceCollector, err := debugutils.DefaultResourceCollector()
			Expect(err).NotTo(HaveOccurred())
			tmpd, err := afero.TempDir(fs, "", "")
			Expect(err).NotTo(HaveOccurred())
			aggregator = debugutils.NewAggregator(resourceCollector, logCollector, storageClient, fs, tmpd)
		})
		It("can properly tar up all resources", func() {
			tmpf, err := afero.TempFile(fs, "", "")
			defer fs.Remove(tmpf.Name())
			Expect(err).NotTo(HaveOccurred())
			err = aggregator.StreamFromManifest(manifests, "gloo-system", tmpf.Name())
			Expect(err).NotTo(HaveOccurred())
			tmpd, err := afero.TempDir(fs, "", "")
			Expect(err).NotTo(HaveOccurred())
			defer fs.RemoveAll(tmpd)
			err = tarutils.Untar(tmpd, tmpf.Name(), fs)
			Expect(err).NotTo(HaveOccurred())
		})
	})

})
