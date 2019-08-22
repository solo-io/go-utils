package clusterlock_test

import (
	"testing"

	"github.com/solo-io/go-utils/testutils/clusterlock"
	"github.com/solo-io/go-utils/testutils/kube"
	"github.com/solo-io/go-utils/testutils/runners/consul"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClusterlock(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clusterlock Suite")
}

var (
	consulFactory *consul.ConsulFactory
	kubeClient    kubernetes.Interface
)

var _ = BeforeSuite(func() {
	kubeClient = kube.MustKubeClient()
	var err error
	consulFactory, err = consul.NewConsulFactory()
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	_ = consulFactory.Clean()
	kubeClient.CoreV1().ConfigMaps("default").Delete(clusterlock.LockResourceName, &v1.DeleteOptions{})
})
