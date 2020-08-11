package debugutils

import (
	"context"
	"errors"
	"os"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/debugutils/test"
	corev1 "k8s.io/api/core/v1"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	"k8s.io/client-go/testing"
)

var _ = Describe("logs unit tests", func() {
	var (
		ctx            context.Context
		podFinder      *MockPodFinder
		requestBuilder *LogRequestBuilder
		podList        *corev1.PodList
	)

	BeforeEach(func() {
		ctrl, ctx = gomock.WithContext(context.TODO(), T)
		podFinder = NewMockPodFinder(ctrl)
		podList = test.GeneratePodList()
		fakeCoreV1 := &fakecorev1.FakeCoreV1{Fake: &testing.Fake{}}
		requestBuilder = NewLogRequestBuilder(fakeCoreV1, podFinder)
	})
	AfterEach(func() {
		ctrl.Finish()
	})

	Context("Log request builder", func() {
		It("works from unstructured", func() {
			resources, err := manifests.ResourceList()
			Expect(err).NotTo(HaveOccurred())
			podFinder.EXPECT().GetPods(ctx, resources).Return([]*corev1.PodList{podList}, nil).Times(1)
			requests, err := requestBuilder.LogsFromUnstructured(ctx, resources)
			Expect(err).NotTo(HaveOccurred())
			Expect(requests).To(HaveLen(4))
		})
		It("fails if pod list fails", func() {
			resources, err := manifests.ResourceList()
			Expect(err).NotTo(HaveOccurred())
			fakeErr := eris.New("this is a fake error")
			podFinder.EXPECT().GetPods(ctx, resources).Return(nil, fakeErr).Times(1)
			_, err = requestBuilder.LogsFromUnstructured(ctx, resources)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, fakeErr)).To(BeTrue())
		})
		It("can build requests for a list of pods", func() {
			requests := requestBuilder.RetrieveLogs(podList)
			Expect(requests).To(HaveLen(4))
		})

		It("can build log requests given a valid pod", func() {
			requests := requestBuilder.buildLogsRequest(*test.GatewayPod())
			Expect(requests).To(HaveLen(1))
			Expect(requests[0].ContainerName).To(Equal("gateway"))
		})

		It("returns 0 given no containers or init containers", func() {
			pod := test.GatewayPod()
			pod.Spec.Containers = []corev1.Container{}
			requests := requestBuilder.buildLogsRequest(*pod)
			Expect(requests).To(HaveLen(0))
		})

		It("works for init containers as well as standard containers", func() {
			initContainerName := "init-container"
			pod := test.GatewayPod()
			pod.Spec.InitContainers = []corev1.Container{{Name: initContainerName}}
			requests := requestBuilder.buildLogsRequest(*pod)
			Expect(requests).To(HaveLen(2))
			Expect(requests[0].ContainerName).To(Equal("gateway"))
			Expect(requests[1].ContainerName).To(Equal(initContainerName))
		})

	})

	Context("Log collector", func() {
		var (
			lc *logCollector
			sc *MockStorageClient
		)

		BeforeEach(func() {
			lc = NewLogCollector(requestBuilder)
			sc = NewMockStorageClient(ctrl)
		})

		It("fails to save logs if a single request fails", func() {
			fakeErr := eris.New("this is a fake error")
			sucessfulRequest, failingRequest := NewMockResponseWrapper(ctrl), NewMockResponseWrapper(ctrl)
			sucessfulRequest.EXPECT().Stream(gomock.Any()).Times(2).Return(&os.File{}, nil)
			failingRequest.EXPECT().Stream(gomock.Any()).Times(1).Return(nil, fakeErr)
			sc.EXPECT().Save(gomock.Any(), gomock.Any()).AnyTimes()
			logRequests := []*LogsRequest{
				{Request: sucessfulRequest},
				{Request: failingRequest},
				{Request: sucessfulRequest},
			}
			err := lc.SaveLogs(ctx, sc, "", logRequests)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, fakeErr)).To(BeTrue())
		})

		It("succeeds and calls storage client for all requests", func() {
			fakeLocation := "location"
			fakeReaderCloser := &os.File{}
			sucessfulRequest := NewMockResponseWrapper(ctrl)
			sucessfulRequest.EXPECT().Stream(gomock.Any()).Times(1).Return(fakeReaderCloser, nil)
			logRequests := []*LogsRequest{
				{
					Request: sucessfulRequest,
					LogMeta: LogMeta{
						ContainerName: "request1",
					},
				},
			}
			sc.EXPECT().Save(fakeLocation, &StorageObject{
				Resource: fakeReaderCloser,
				Name:     logRequests[0].ResourceId(),
			}).Return(nil).Times(1)
			err := lc.SaveLogs(ctx, sc, fakeLocation, logRequests)
			Expect(err).NotTo(HaveOccurred())
		})

	})

})
