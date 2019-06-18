package debugutils

import (
	"os"

	"github.com/ghodss/yaml"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/errors"
	corev1 "k8s.io/api/core/v1"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	"k8s.io/client-go/testing"
)

var _ = Describe("logs unit tests", func() {
	var (
		podFinder      *MockPodFinder
		requestBuilder *LogRequestBuilder
		podList        *corev1.PodList
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)
		podFinder = NewMockPodFinder(ctrl)
		podList = generatePodList()
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
			podFinder.EXPECT().GetPods(resources).Return([]*corev1.PodList{podList}, nil).Times(1)
			requests, err := requestBuilder.LogsFromUnstructured(resources)
			Expect(err).NotTo(HaveOccurred())
			Expect(requests).To(HaveLen(4))
		})
		It("fails if pod list fails", func() {
			resources, err := manifests.ResourceList()
			Expect(err).NotTo(HaveOccurred())
			fakeErr := errors.New("this is a fake error")
			podFinder.EXPECT().GetPods(resources).Return(nil, fakeErr).Times(1)
			_, err = requestBuilder.LogsFromUnstructured(resources)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, fakeErr)).To(BeTrue())
		})
		It("can build requests for a list of pods", func() {
			requests := requestBuilder.RetrieveLogs(podList)
			Expect(requests).To(HaveLen(4))
		})

		It("can build log requests given a valid pod", func() {
			requests := requestBuilder.buildLogsRequest(*gatewayPod())
			Expect(requests).To(HaveLen(1))
			Expect(requests[0].ContainerName).To(Equal("gateway"))
		})

		It("returns 0 given no containers or init containers", func() {
			pod := gatewayPod()
			pod.Spec.Containers = []corev1.Container{}
			requests := requestBuilder.buildLogsRequest(*pod)
			Expect(requests).To(HaveLen(0))
		})

		It("works for init containers as well as standard containers", func() {
			initContainerName := "init-container"
			pod := gatewayPod()
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
			fakeErr := errors.New("this is a fake error")
			sucessfulRequest, failingRequest := NewMockResponseWrapper(ctrl), NewMockResponseWrapper(ctrl)
			sucessfulRequest.EXPECT().Stream().Times(2).Return(&os.File{}, nil)
			failingRequest.EXPECT().Stream().Times(1).Return(nil, fakeErr)
			sc.EXPECT().Save(gomock.Any(), gomock.Any()).AnyTimes()
			logRequests := []*LogsRequest{
				{Request: sucessfulRequest},
				{Request: failingRequest},
				{Request: sucessfulRequest},
			}
			err := lc.SaveLogs(sc, "", logRequests)
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, fakeErr)).To(BeTrue())
		})

		It("succeeds and calls storage client for all requests", func() {
			fakeLocation := "location"
			fakeReaderCloser := &os.File{}
			sucessfulRequest := NewMockResponseWrapper(ctrl)
			sucessfulRequest.EXPECT().Stream().Times(1).Return(fakeReaderCloser, nil)
			logRequests := []*LogsRequest{
				{
					Request:       sucessfulRequest,
					ContainerName: "request1",
				},
			}
			sc.EXPECT().Save(fakeLocation, &StorageObject{
				Resource: fakeReaderCloser,
				Name:     logRequests[0].ResourceId(),
			}).Return(nil).Times(1)
			err := lc.SaveLogs(sc, fakeLocation, logRequests)
			Expect(err).NotTo(HaveOccurred())
		})

	})

})

func generatePodList() *corev1.PodList {
	podList := &corev1.PodList{
		Items: []corev1.Pod{*gatewayPod(), *gatewayProxyPod(), *glooPod(), *discoveryPod()},
	}
	return podList
}

var (
	gatewayPod = func() *corev1.Pod {
		return convertPod([]byte(gatewayPodYaml))
	}
	discoveryPod = func() *corev1.Pod {
		return convertPod([]byte(discoveryPodYaml))
	}
	gatewayProxyPod = func() *corev1.Pod {
		return convertPod([]byte(gatewayProxyPodYaml))
	}
	glooPod = func() *corev1.Pod {
		return convertPod([]byte(glooPodYaml))
	}
)

func convertPod(podYaml []byte) *corev1.Pod {
	var pod corev1.Pod
	Expect(yaml.Unmarshal(podYaml, &pod)).NotTo(HaveOccurred())
	return &pod
}

var discoveryPodYaml = `
apiVersion: v1
kind: Pod
metadata:
  annotations:
    prometheus.io/path: /metrics
    prometheus.io/port: "9091"
    prometheus.io/scrape: "true"
  labels:
    gloo: discovery
    pod-template-hash: 688d68bb6b
  name: discovery-688d68bb6b-4c4cp
  namespace: gloo-system
spec:
  containers:
  - env:
    - name: POD_NAMESPACE
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.namespace
    - name: START_STATS_SERVER
      value: "true"
    image: quay.io/solo-io/discovery:0.13.33
    imagePullPolicy: Always
    name: discovery
    resources: {}
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL
      procMount: Default
      readOnlyRootFilesystem: true
      runAsNonRoot: true
      runAsUser: 10101
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-9xmkt
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  nodeName: minikube
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  volumes:
  - name: default-token-9xmkt
    secret:
      defaultMode: 420
      secretName: default-token-9xmkt
  phase: Running
`

var glooPodYaml = `
apiVersion: v1
kind: Pod
metadata:
  annotations:
    prometheus.io/path: /metrics
    prometheus.io/port: "9091"
    prometheus.io/scrape: "true"
  labels:
    gloo: gloo
    pod-template-hash: d97b6bd64
  name: gloo-d97b6bd64-6sgrm
  namespace: gloo-system
spec:
  containers:
  - env:
    - name: POD_NAMESPACE
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.namespace
    - name: START_STATS_SERVER
      value: "true"
    image: quay.io/solo-io/gloo:0.13.33
    imagePullPolicy: Always
    name: gloo
    ports:
    - containerPort: 9977
      name: grpc
      protocol: TCP
    resources:
      requests:
        cpu: 500m
        memory: 256Mi
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL
      procMount: Default
      readOnlyRootFilesystem: true
      runAsNonRoot: true
      runAsUser: 10101
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-9xmkt
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  nodeName: minikube
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  phase: Running

`

var gatewayPodYaml = `
apiVersion: v1
kind: Pod
metadata:
  annotations:
    prometheus.io/path: /metrics
    prometheus.io/port: "9091"
    prometheus.io/scrape: "true"
  labels:
    gloo: gateway
    pod-template-hash: 5d77dd7d4b
  name: gateway-5d77dd7d4b-c8gk2
  namespace: gloo-system
spec:
  containers:
  - env:
    - name: POD_NAMESPACE
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.namespace
    - name: START_STATS_SERVER
      value: "true"
    image: quay.io/solo-io/gateway:0.13.33
    imagePullPolicy: Always
    name: gateway
    resources: {}
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL
      procMount: Default
      readOnlyRootFilesystem: true
      runAsNonRoot: true
      runAsUser: 10101
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-9xmkt
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  nodeName: minikube
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  volumes:
  - name: default-token-9xmkt
    secret:
      defaultMode: 420
      secretName: default-token-9xmkt
  phase: Running
`

var gatewayProxyPodYaml = `
apiVersion: v1
kind: Pod
metadata:
  annotations:
    prometheus.io/path: /metrics
    prometheus.io/port: "8081"
    prometheus.io/scrape: "true"
  labels:
    gloo: gateway-proxy
    pod-template-hash: 655765b499
  name: gateway-proxy-655765b499-wqxhm
  namespace: gloo-system
spec:
  containers:
  - args:
    - --disable-hot-restart
    env:
    - name: POD_NAMESPACE
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.namespace
    - name: POD_NAME
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.name
    image: quay.io/solo-io/gloo-envoy-wrapper:0.13.33
    imagePullPolicy: Always
    name: gateway-proxy
    ports:
    - containerPort: 8080
      name: http
      protocol: TCP
    - containerPort: 8443
      name: https
      protocol: TCP
    resources: {}
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        add:
        - NET_BIND_SERVICE
        drop:
        - ALL
      procMount: Default
      readOnlyRootFilesystem: true
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /etc/envoy
      name: envoy-config
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-9xmkt
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  nodeName: minikube
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  volumes:
  - configMap:
      defaultMode: 420
      name: gateway-proxy-envoy-config
    name: envoy-config
  - name: default-token-9xmkt
    secret:
      defaultMode: 420
      secretName: default-token-9xmkt
  phase: Running

`
