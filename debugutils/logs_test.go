package debugutils

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("logs", func() {
	var (
		fs             afero.Fs

		deployedPods = []*LogsRequest{
			{
				podMeta: metav1.ObjectMeta{
					Name: "gateway",
					Namespace: "gloo-system",
				},
				containerName: "gateway",
			},
			{
				podMeta: metav1.ObjectMeta{
					Name: "gateway-proxy",
					Namespace: "gloo-system",
				},
				containerName: "gateway-proxy",
			},
			{
				podMeta: metav1.ObjectMeta{
					Name: "gloo",
					Namespace: "gloo-system",
				},
				containerName: "gloo",
			},
			{
				podMeta: metav1.ObjectMeta{
					Name: "discovery",
					Namespace: "gloo-system",
				},
				containerName: "discovery",
			},
		}

		mustRequestBuilder = func() *LogRequestBuilder {
			requestBuilder, err := NewLogRequestBuilder()
			Expect(err).NotTo(HaveOccurred())
			return requestBuilder
		}
	)

	Context("request builder", func() {

		It("can properly build the requests from the gloo manifest", func() {
			requestBuilder := mustRequestBuilder()
			requests, err := requestBuilder.LogsFromManifest(manifests)
			Expect(err).NotTo(HaveOccurred())
			Expect(requests).To(HaveLen(4))
			for _, deployedPod := range deployedPods {
				found := false
				for _, request := range requests {
					if request.containerName == deployedPod.containerName &&
						strings.HasPrefix(request.podMeta.Name, deployedPod.podMeta.Name) {
						found = true
						continue
					}
				}
				Expect(found).To(BeTrue())
			}
		})
	})

	Context("log file storage", func() {
		var (
			lfs    *LogStorageClient
			tmpDir string
		)

		It("can properly store all logs from gloo manifest to files", func() {
			var err error
			fs = afero.NewOsFs()
			tmpDir, err = afero.TempDir(fs, "", "")
			Expect(err).NotTo(HaveOccurred())
			lfs = NewLogFileStorage(fs, tmpDir)
			requestBuilder := mustRequestBuilder()
			requests, err := requestBuilder.LogsFromManifest(manifests)
			Expect(requests).To(HaveLen(4))
			Expect(err).NotTo(HaveOccurred())
			err = lfs.FetchLogs(requests)
			Expect(err).NotTo(HaveOccurred())
			files, err := afero.ReadDir(fs, lfs.Dir())
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(HaveLen(4))
			for _, deployedPod := range deployedPods {
				found := false
				for _, file := range files {
					fileName := file.Name()
					prefix := fmt.Sprintf("%s_%s", deployedPod.podMeta.Namespace, deployedPod.podMeta.Name)
					suffix := fmt.Sprintf("%s.log", deployedPod.containerName)
					if strings.HasPrefix(fileName, prefix) && strings.HasSuffix(fileName, suffix){
						found = true
						continue
					}
				}
				Expect(found).To(BeTrue())
			}
		})
		AfterEach(func() {
			fs.Remove(tmpDir)
		})
	})
})
