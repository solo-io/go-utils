package debugutils

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/tarutils"
	"github.com/spf13/afero"
)

var _ = Describe("aggregator test", func() {
	var (
		aggregator *Aggregator
	)

	Context("e2e", func() {
		BeforeEach(func() {
			var err error
			aggregator, err = NewDefaultAggregator()
			Expect(err).NotTo(HaveOccurred())
		})
		It("can properly tar up all resources", func() {
			tmpf, err := afero.TempFile(aggregator.fs, "", "")
			defer aggregator.fs.Remove(tmpf.Name())
			Expect(err).NotTo(HaveOccurred())
			err = aggregator.StreamFromManifest(manifests, "gloo-system", tmpf.Name())
			Expect(err).NotTo(HaveOccurred())
			tmpd, err := afero.TempDir(aggregator.fs, "", "")
			Expect(err).NotTo(HaveOccurred())
			defer aggregator.fs.RemoveAll(tmpd)
			err = tarutils.Untar(tmpd, tmpf.Name(), aggregator.fs)
			Expect(err).NotTo(HaveOccurred())

		})
	})
})