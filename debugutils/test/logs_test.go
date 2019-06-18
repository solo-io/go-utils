package test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/debugutils"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("logs", func() {
	var (
		fs afero.Fs

		deployedPods = []*debugutils.LogsRequest{
			{
				PodMeta: metav1.ObjectMeta{
					Name:      "gateway",
					Namespace: "gloo-system",
				},
				ContainerName: "gateway",
			},
			{
				PodMeta: metav1.ObjectMeta{
					Name:      "gateway-proxy",
					Namespace: "gloo-system",
				},
				ContainerName: "gateway-proxy",
			},
			{
				PodMeta: metav1.ObjectMeta{
					Name:      "gloo",
					Namespace: "gloo-system",
				},
				ContainerName: "gloo",
			},
			{
				PodMeta: metav1.ObjectMeta{
					Name:      "discovery",
					Namespace: "gloo-system",
				},
				ContainerName: "discovery",
			},
		}

		mustRequestBuilder = func() *debugutils.LogRequestBuilder {
			requestBuilder, err := debugutils.DefaultLogRequestBuilder()
			Expect(err).NotTo(HaveOccurred())
			return requestBuilder
		}
	)

	Context("request builder", func() {

		It("can properly build the requests from the gloo manifest", func() {
			requestBuilder := mustRequestBuilder()
			resources, err := manifests.ResourceList()
			Expect(err).NotTo(HaveOccurred())
			requests, err := requestBuilder.LogsFromUnstructured(resources)
			Expect(err).NotTo(HaveOccurred())
			Expect(requests).To(HaveLen(4))
			for _, deployedPod := range deployedPods {
				found := false
				for _, request := range requests {
					if request.ContainerName == deployedPod.ContainerName &&
						strings.HasPrefix(request.PodMeta.Name, deployedPod.PodMeta.Name) {
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
			lc     debugutils.LogCollector
			sc     StorageClient
			tmpDir string
		)

		It("can properly store all logs from gloo manifest to files", func() {
			var err error
			fs = afero.NewOsFs()
			sc = NewFileStorageClient(fs)
			tmpDir, err = afero.TempDir(fs, "", "")
			Expect(err).NotTo(HaveOccurred())
			lc, err = debugutils.DefaultLogCollector()
			Expect(err).NotTo(HaveOccurred())
			unstructured, err := manifests.ResourceList()
			Expect(err).NotTo(HaveOccurred())
			requests, err := lc.GetLogRequests(unstructured)
			Expect(requests).To(HaveLen(4))
			Expect(err).NotTo(HaveOccurred())
			err = lc.SaveLogs(sc, tmpDir, requests)
			Expect(err).NotTo(HaveOccurred())
			files, err := afero.ReadDir(fs, tmpDir)
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(HaveLen(4))
			for _, deployedPod := range deployedPods {
				found := false
				for _, file := range files {
					fileName := file.Name()
					prefix := fmt.Sprintf("%s_%s", deployedPod.PodMeta.Namespace, deployedPod.PodMeta.Name)
					suffix := fmt.Sprintf("%s.log", deployedPod.ContainerName)
					if strings.HasPrefix(fileName, prefix) && strings.HasSuffix(fileName, suffix) {
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
