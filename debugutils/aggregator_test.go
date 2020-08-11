package debugutils

import (
	"context"
	"path/filepath"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/stringutils"
	"github.com/spf13/afero"
)

var _ = Describe("aggregator test", func() {
	var (
		aggregator *Aggregator
	)

	Context("unit", func() {
		var (
			resourceCollector *MockResourceCollector
			logCollector      *MockLogCollector
			storageClient     *MockStorageClient
			fs                afero.Fs
			tmpd              string
		)
		BeforeEach(func() {
			var err error
			ctrl = gomock.NewController(T)
			logCollector = NewMockLogCollector(ctrl)
			resourceCollector = NewMockResourceCollector(ctrl)
			storageClient = NewMockStorageClient(ctrl)
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
			resourceCollector.EXPECT().RetrieveResources(gomock.Any(), gomock.Any(), namespace, gomock.Any()).Return(nil, nil).Times(1)
			resourceCollector.EXPECT().SaveResources(gomock.Any(), storageClient, filepath.Join(tmpd, "resources"), nil).Return(nil).Times(1)
			logCollector.EXPECT().GetLogRequests(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
			logCollector.EXPECT().SaveLogs(gomock.Any(), storageClient, filepath.Join(tmpd, "logs"), nil).Times(1)
			storageClient.EXPECT().Save(filepath.Dir(filename), gomock.Any()).Return(nil).Times(1)

			err := aggregator.StreamFromManifest(context.Background(), manifests, namespace, filename)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			ctrl.Finish()
		})
	})
})
