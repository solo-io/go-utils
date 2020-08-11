package clusterlock_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/solo-io/go-utils/testutils/runners/consul"

	"github.com/avast/retry-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/go-utils/testutils/clusterlock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("kube cluster lock test", func() {

	var (
		ctx       context.Context
		namespace string
	)

	BeforeEach(func() {
		ctx = context.Background()
		namespace = testutils.RandString(8)
		err := kubeutils.CreateNamespacesInParallel(ctx, kubeClient, namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := kubeutils.DeleteNamespacesInParallelBlocking(ctx, kubeClient, namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can handle a single locking scenario", func() {
		lock, err := clusterlock.NewTestClusterLocker(kubeClient, clusterlock.Options{
			Namespace: namespace,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
		Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
	})

	It("can handle synchronous requests", func() {
		for idx := 0; idx < 5; idx++ {
			lock, err := clusterlock.NewTestClusterLocker(kubeClient, clusterlock.Options{
				Namespace: namespace,
			})
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
				lock, err := clusterlock.NewTestClusterLocker(kubeClient, clusterlock.Options{
					Namespace: namespace,
				})
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
		lock, err := clusterlock.NewTestClusterLocker(kubeClient, clusterlock.Options{
			Namespace: namespace,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
		lock2, err := clusterlock.NewTestClusterLocker(kubeClient, clusterlock.Options{
			Namespace: namespace,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(lock2.AcquireLock(retry.Delay(time.Millisecond), retry.Attempts(3))).To(HaveOccurred())
		Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
	})

	It("Take back timed out lock", func() {
		lock, err := clusterlock.NewTestClusterLocker(kubeClient, clusterlock.Options{
			Namespace: namespace,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
		cfgMap, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(context.Background(), clusterlock.LockResourceName, v1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		cfgMap.Annotations[clusterlock.LockTimeoutAnnotationKey] = time.Now().Add(time.Duration(-1) * time.Minute).Format(clusterlock.DefaultTimeFormat)
		_, err = kubeClient.CoreV1().ConfigMaps(namespace).Update(context.Background(), cfgMap, v1.UpdateOptions{})
		Expect(err).NotTo(HaveOccurred())
		lock2, err := clusterlock.NewTestClusterLocker(kubeClient, clusterlock.Options{
			Namespace: namespace,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(lock2.AcquireLock()).NotTo(HaveOccurred())
		Expect(lock2.ReleaseLock()).NotTo(HaveOccurred())
	})

	It("fails to release if a different lock requester tries to release it", func() {
		lock, err := clusterlock.NewTestClusterLocker(kubeClient, clusterlock.Options{
			Namespace: namespace,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
		lock2, err := clusterlock.NewTestClusterLocker(kubeClient, clusterlock.Options{
			Namespace: namespace,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(clusterlock.IsNotLockOwnerError(lock2.ReleaseLock())).To(BeTrue())
		Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
	})

	It("releases lock after timeout if heartbeat stops", func() {
		seperateContext, cancel := context.WithCancel(context.Background())
		opts := clusterlock.Options{
			Context: seperateContext,
		}
		lock, err := clusterlock.NewTestClusterLocker(kubeClient, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
		cancel()
		lock2, err := clusterlock.NewTestClusterLocker(kubeClient, clusterlock.Options{
			Namespace: namespace,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(lock2.AcquireLock()).NotTo(HaveOccurred())
	})
})

var _ = Describe("consul cluster lock test", func() {
	var (
		consulClient   *api.Client
		consulInstance *consul.ConsulInstance
		keyPrefix      = testutils.RandString(6)
	)
	BeforeEach(func() {
		var err error

		consulInstance, err = consulFactory.NewConsulInstance()
		Expect(err).NotTo(HaveOccurred())
		err = consulInstance.Run()
		Expect(err).NotTo(HaveOccurred())

		cfg := api.DefaultConfig()
		cfg.Address = fmt.Sprintf("127.0.0.1:%v", consulInstance.Ports.HttpPort)

		consulClient, err = api.NewClient(cfg)
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		_ = consulInstance.Clean()
	})

	It("can handle a single locking scenario", func() {
		lock, err := clusterlock.NewConsulClusterLocker(context.Background(), keyPrefix, consulClient)
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
		Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
	})

	It("can handle synchronous requests", func() {
		for idx := 0; idx < 5; idx++ {
			lock, err := clusterlock.NewConsulClusterLocker(context.Background(), keyPrefix, consulClient)
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
			idx := idx
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer GinkgoRecover()
				lock, err := clusterlock.NewConsulClusterLocker(context.Background(), fmt.Sprintf("%v-", idx), consulClient)
				Expect(err).NotTo(HaveOccurred())
				Expect(lock.AcquireLock(retry.Delay(time.Second))).NotTo(HaveOccurred())
				Expect(*sharedString).To(Equal(""))
				*sharedString = fmt.Sprintf("%v", idx)
				time.Sleep(time.Second)
				*sharedString = ""
				Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
			}()
		}
		wg.Wait()
	})

	It("errors out if lock isn't free after a set amount of time", func() {
		lock, err := clusterlock.NewConsulClusterLocker(context.Background(), keyPrefix, consulClient)
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
		lock2, err := clusterlock.NewConsulClusterLocker(context.Background(), keyPrefix, consulClient)
		Expect(err).NotTo(HaveOccurred())
		Expect(err).NotTo(HaveOccurred())
		Expect(lock2.AcquireLock(retry.Delay(time.Millisecond), retry.Attempts(3))).To(HaveOccurred())
		Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
	})

})
