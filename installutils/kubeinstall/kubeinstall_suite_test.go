package kubeinstall_test

import (
	"os"
	"testing"

	"github.com/solo-io/go-utils/testutils/kube"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/testutils/clusterlock"
)

func TestKubeinstall(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubeinstall Suite")
}

var (
	lock *clusterlock.TestClusterLocker
	err  error
)

var _ = BeforeSuite(func() {
	lock, err = clusterlock.NewTestClusterLocker(kube.MustKubeClient(), clusterlock.Options{
		IdPrefix: os.ExpandEnv("supergloo-helm-{$BUILD_ID}-"),
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(lock.AcquireLock()).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
})
