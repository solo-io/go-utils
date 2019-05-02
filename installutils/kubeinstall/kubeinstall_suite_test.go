package kubeinstall_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKubeinstall(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubeinstall Suite")
}

// var (
// 	lock *clusterlock.TestClusterLocker
// 	err  error
// )
//
// var _ = SynchronizedBeforeSuite(func() []byte {
// 	lock, err = clusterlock.NewTestClusterLocker(kube.MustKubeClient(), clusterlock.Options{
// 		IdPrefix: os.ExpandEnv("supergloo-helm-{$BUILD_ID}-"),
// 	})
// 	Expect(err).NotTo(HaveOccurred())
// 	Expect(lock.AcquireLock()).NotTo(HaveOccurred())
// 	return nil
// }, func(data []byte) {})
//
// var _ = SynchronizedAfterSuite(func() {}, func() {
// 	Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
// })
