package kubeinstallutils_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeinstallutils"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/go-utils/testutils/kube"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var _ = Describe("InstallKubeManifest", func() {
	var (
		namespace  string
		kubeClient kubernetes.Interface
	)
	BeforeEach(func() {
		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			Skip("use RUN_KUBE_TESTS to run this test")
		}
		namespace = "install-kube-manifest-" + testutils.RandString(8)
		kubeClient = kube.MustKubeClient()
		err := kubeutils.CreateNamespacesInParallel(kubeClient, namespace)
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		err := kubeutils.DeleteNamespacesInParallelBlocking(kubeClient, namespace)
		Expect(err).NotTo(HaveOccurred())
	})
	It("installs arbitrary kube manifests", func() {
		err := deployNginx(namespace)
		Expect(err).NotTo(HaveOccurred())

		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		kube, err := kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())

		svcs, err := kube.CoreV1().Services(namespace).List(v1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		deployments, err := kube.ExtensionsV1beta1().Deployments(namespace).List(v1.ListOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(svcs.Items).To(HaveLen(1))
		Expect(deployments.Items).To(HaveLen(1))

	})
})

func deployNginx(namespace string) error {
	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return err
	}
	kube, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	apiext, err := clientset.NewForConfig(cfg)
	if err != nil {
		return err
	}

	installer := kubeinstallutils.NewKubeInstaller(kube, apiext, namespace)

	kubeObjs, err := kubeinstallutils.ParseKubeManifest(testutils.NginxYaml)
	if err != nil {
		return err
	}

	for _, kubeOjb := range kubeObjs {
		if err := installer.Create(kubeOjb); err != nil {
			return err
		}
	}
	return nil
}
