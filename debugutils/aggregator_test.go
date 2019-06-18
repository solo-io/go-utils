package debugutils

import (
	"path/filepath"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/debugutils/mocks"
	"github.com/solo-io/go-utils/stringutils"
	"github.com/spf13/afero"
)

var _ = Describe("aggregator test", func() {
	var (
		aggregator *Aggregator
	)

	Context("unit", func() {
		var (
			resourceCollector *mocks.MockResourceCollector
			logCollector      *mocks.MockLogCollector
			storageClient     *mocks.MockStorageClient
			fs                afero.Fs
			tmpd              string
		)
		BeforeEach(func() {
			var err error
			ctrl = gomock.NewController(T)
			logCollector = mocks.NewMockLogCollector(ctrl)
			resourceCollector = mocks.NewMockResourceCollector(ctrl)
			storageClient = mocks.NewMockStorageClient(ctrl)
			fs = afero.NewMemMapFs()
			tmpd, err = afero.TempDir(fs, "", "")
			Expect(err).NotTo(HaveOccurred())
			aggregator = NewAggregator(resourceCollector, logCollector, storageClient, fs, tmpd)
		})

		It("can properly create subdirectories", func() {
			directories := []string{"resources", "logs"}
			err := aggregator.createSubResourceDirectories()
			Expect(err).NotTo(HaveOccurred())
			files, err := afero.ReadDir(fs, tmpd)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(HaveLen(2))
			for _, v := range files {
				Expect(stringutils.ContainsString(filepath.Base(v.Name()), directories))
			}
		})

		It("properly sets all filepaths", func() {
			namespace := "ns"
			filename := "/hello/world/test.tgz"
			resourceCollector.EXPECT().RetrieveResources(gomock.Any(), namespace, gomock.Any()).Return(nil, nil).Times(1)
			resourceCollector.EXPECT().SaveResources(storageClient, filepath.Join(tmpd, "resources"), nil).Return(nil).Times(1)
			logCollector.EXPECT().GetLogRequests(gomock.Any()).Return(nil, nil).Times(1)
			logCollector.EXPECT().SaveLogs(storageClient, filepath.Join(tmpd, "logs"), nil).Times(1)
			storageClient.EXPECT().Save(filepath.Dir(filename), gomock.Any()).Return(nil).Times(1)

			err := aggregator.StreamFromManifest(manifests, namespace, filename)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			ctrl.Finish()
		})
	})
})
