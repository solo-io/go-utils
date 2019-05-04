package helper

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/go-utils/testutils"
	kube2 "github.com/solo-io/go-utils/testutils/kube"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("test container tests", func() {

	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}

	var (
		namespace string
		kube      kubernetes.Interface
	)

	BeforeSuite(func() {
		namespace = testutils.RandString(8)
		kube = kube2.MustKubeClient()
		err := kubeutils.CreateNamespacesInParallel(kube, namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterSuite(func() {
		err := kubeutils.DeleteNamespacesInParallelBlocking(kube, namespace)
		Expect(err).NotTo(HaveOccurred())
	})
	Context("test runner", func() {

		var (
			testRunner *TestRunner
		)
		BeforeEach(func() {
			var err error
			testRunner, err = NewTestRunner(namespace)
			Expect(err).NotTo(HaveOccurred())
			err = testRunner.Deploy(time.Minute*2)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := testRunner.Terminate()
			Expect(err).NotTo(HaveOccurred())
		})

		It("can install and uninstall the testrunner", func() {
			// responseString := fmt.Sprintf(`"%s":"%s.%s.svc.cluster.local:%v"`,
			// 	linkerd.HeaderKey, helper.HttpEchoName, testHelper.InstallNamespace, helper.HttpEchoPort)
			host := fmt.Sprintf("%s.%s.svc.cluster.local:%v", TestrunnerName, namespace, TestRunnerPort)
			testRunner.CurlEventuallyShouldRespond(CurlOpts{
				Protocol:          "http",
				Path:              "/",
				Method:            "GET",
				Host:              host,
				Service:           TestrunnerName,
				Port:              TestRunnerPort,
				ConnectionTimeout: 10,
			}, SimpleHttpResponse, 1, 120*time.Second)
		})


	})

	Context("http ehco", func() {

		var (
			httpEcho *HttpEcho
		)
		BeforeEach(func() {
			var err error
			httpEcho, err = NewEchoHttp(namespace)
			Expect(err).NotTo(HaveOccurred())
			err = httpEcho.deploy(time.Minute)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := httpEcho.Terminate()
			Expect(err).NotTo(HaveOccurred())
		})
		It("can install and uninstall the http echo pod", func() {
			responseString := fmt.Sprintf(`"host":"%s.%s.svc.cluster.local:%v"`,
				HttpEchoName, namespace, HttpEchoPort)
			host := fmt.Sprintf("%s.%s.svc.cluster.local:%v", HttpEchoName, namespace, HttpEchoPort)
			httpEcho.CurlEventuallyShouldRespond(CurlOpts{
				Protocol:          "http",
				Path:              "/",
				Method:            "GET",
				Host:              host,
				Service:           HttpEchoName,
				Port:              HttpEchoPort,
				ConnectionTimeout: 10,
			}, responseString, 1, 120*time.Second)
		})
	})
})
