package clusterlock_test

import (
	"context"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/testutils/clusterlock"
	"sync"
	"time"

	"github.com/avast/retry-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("cluster lock test", func() {

	var (
		kubeClient kubernetes.Interface
		ctx        = context.Background()
	)

	var _ = BeforeSuite(func() {
		kubeClient = MustKubeClient()
	})

	var _ = AfterSuite(func() {
		kubeClient.CoreV1().ConfigMaps("default").Delete(clusterlock.LockResourceName, &v1.DeleteOptions{})
	})

	It("can handle a single locking scenario", func() {
		lock, err := clusterlock.NewTestClusterLocker(ctx, kubeClient, "default")
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
		Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
	})

	It("can handle synchronous requests", func() {
		for idx := 0; idx < 5; idx++ {
			lock, err := clusterlock.NewTestClusterLocker(ctx, kubeClient, "default")
			Expect(err).NotTo(HaveOccurred())
			Expect(lock.AcquireLock()).NotTo(HaveOccurred())
			Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
		}
	})

	It("can handle concurrent requests", func() {
		x := ""
		sharedString := &x
		wg := sync.WaitGroup{}
		for idx := 0; idx < 5; idx++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer GinkgoRecover()
				lock, err := clusterlock.NewTestClusterLocker(ctx, kubeClient, "default")
				Expect(err).NotTo(HaveOccurred())
				Expect(lock.AcquireLock(retry.Delay(time.Second))).NotTo(HaveOccurred())
				Expect(*sharedString).To(Equal(""))
				*sharedString = "hello"
				time.Sleep(time.Second)
				*sharedString = ""
				Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
			}()
		}
		wg.Wait()
	})

	It("errors out if lock isn't free after a set amount of time", func() {
		lock, err := clusterlock.NewTestClusterLocker(ctx, kubeClient, "default")
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
		lock2, err := clusterlock.NewTestClusterLocker(ctx, kubeClient, "default")
		Expect(err).NotTo(HaveOccurred())
		Expect(lock2.AcquireLock(retry.Delay(time.Millisecond), retry.Attempts(3))).To(HaveOccurred())
		Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
	})

	It("Take back timed out lock", func() {
		lock, err := clusterlock.NewTestClusterLocker(ctx, kubeClient, "default")
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
		cfgMap, err := kubeClient.CoreV1().ConfigMaps("default").Get(clusterlock.LockResourceName, v1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		cfgMap.Annotations[clusterlock.LockTimeoutAnnotationKey] = time.Now().Add(time.Duration(-1) * time.Minute).Format(clusterlock.DefaultTimeFormat)
		_, err = kubeClient.CoreV1().ConfigMaps("default").Update(cfgMap)
		Expect(err).NotTo(HaveOccurred())
		lock2, err := clusterlock.NewTestClusterLocker(ctx, kubeClient, "default")
		Expect(err).NotTo(HaveOccurred())
		Expect(lock2.AcquireLock()).NotTo(HaveOccurred())
		Expect(lock2.ReleaseLock()).NotTo(HaveOccurred())
	})

	It("fails to release if a different lock requester tries to release it", func() {
		lock, err := clusterlock.NewTestClusterLocker(ctx, kubeClient, "default")
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
		lock2, err := clusterlock.NewTestClusterLocker(ctx, kubeClient, "default")
		Expect(err).NotTo(HaveOccurred())
		Expect(clusterlock.IsNotLockOwnerError(lock2.ReleaseLock())).To(BeTrue())
	})

	It("releases lock after timeout if heartbeat stops", func() {
		seperateContext, cancel := context.WithCancel(context.Background())
		lock, err := clusterlock.NewTestClusterLocker(seperateContext, kubeClient, "default")
		Expect(err).NotTo(HaveOccurred())
		cancel()
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
		lock2, err := clusterlock.NewTestClusterLocker(ctx, kubeClient, "default")
		Expect(err).NotTo(HaveOccurred())
		Expect(lock2.AcquireLock()).NotTo(HaveOccurred())
	})
})

func MustKubeClient() kubernetes.Interface {
	restConfig, err := kubeutils.GetConfig("", "")
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return kubeClient
}
