package kubeinstall

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/solo-io/go-utils/kubeerrutils"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/solo-io/go-utils/testutils/clusterlock"
	batchv1 "k8s.io/api/batch/v1"
	kubev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/solo-io/go-utils/testutils/kube"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/solo-io/go-utils/installutils/kuberesource"
	appsv1 "k8s.io/api/apps/v1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/go-utils/kubeutils"

	"github.com/solo-io/go-utils/testutils"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var istioCrd = apiextensions.CustomResourceDefinition{}

var _ = XDescribe("KubeInstaller", func() {
	var (
		ns         string
		lock       *clusterlock.TestClusterLocker
		kubeClient kubernetes.Interface
	)

	BeforeSuite(func() {
		var err error
		idPrefix := fmt.Sprintf("kube-installer-%s-", os.Getenv("BUILD_ID"))
		lock, err = clusterlock.NewTestClusterLocker(kube.MustKubeClient(), clusterlock.Options{
			IdPrefix: idPrefix,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(lock.AcquireLock()).NotTo(HaveOccurred())
	})

	AfterSuite(func() {
		Expect(lock.ReleaseLock()).NotTo(HaveOccurred())
	})

	BeforeEach(func() {
		kubeClient = kube.MustKubeClient()
		// wait for all services in the previous namespace to be torn down
		// important because of a race caused by nodeport conflcit
		if ns != "" {
			kube.WaitForNamespaceTeardown(ns)
		}
		ns = "test" + testutils.RandString(5)
		err := kubeutils.CreateNamespacesInParallel(kubeClient, ns)
		Expect(err).NotTo(HaveOccurred())

	})
	AfterEach(func() {
		err := kubeutils.DeleteNamespacesInParallelBlocking(kubeClient, ns)
		Expect(err).NotTo(HaveOccurred())
		kube.TeardownClusterResourcesWithPrefix(kubeClient, "istio")
		kube.TeardownClusterResourcesWithPrefix(kubeClient, "prometheus")
		kube.WaitForNamespaceTeardown(ns)
	})

	Context("updating resource from cache", func() {
		It("does nothing if the resource hasn't changed", func() {
			unique := "unique"
			randomLabel := testutils.RandString(8)
			ownerLabels := map[string]string{
				unique: randomLabel,
			}
			// cache a resource
			cache := NewCache()
			cache.access.Unlock()
			resource := makeConfigmap(unique, randomLabel, ns)
			cache.resources = make(kuberesource.UnstructuredResourcesByKey)

			cache.Set(resource.DeepCopy())

			resource.SetAnnotations(nil)

			restCfg, err := kubeutils.GetConfig("", "")
			Expect(err).NotTo(HaveOccurred())

			dynamicClient, err := client.New(restCfg, client.Options{})
			Expect(err).NotTo(HaveOccurred())

			written := resource.DeepCopy()
			err = dynamicClient.Create(context.TODO(), written)
			Expect(err).NotTo(HaveOccurred())

			inst, err := NewKubeInstaller(restCfg, cache, nil)
			Expect(err).NotTo(HaveOccurred())

			err = inst.ReconcileResources(context.TODO(), NewReconcileParams(ns, kuberesource.UnstructuredResources{resource}, ownerLabels, false))
			Expect(err).NotTo(HaveOccurred())

			afterReconcile := resource.DeepCopy()
			err = dynamicClient.Get(context.TODO(), client.ObjectKey{Name: afterReconcile.GetName(), Namespace: afterReconcile.GetNamespace()}, afterReconcile)
			Expect(err).NotTo(HaveOccurred())

			Expect(afterReconcile.GetResourceVersion()).To(Equal(written.GetResourceVersion()))

			// update so it gets written
			an := resource.GetAnnotations()
			an["hi"] = "bye"
			resource.SetAnnotations(an)

			err = inst.ReconcileResources(context.TODO(), NewReconcileParams(ns, kuberesource.UnstructuredResources{resource}, ownerLabels, false))
			Expect(err).NotTo(HaveOccurred())

			err = dynamicClient.Get(context.TODO(), client.ObjectKey{Name: afterReconcile.GetName(), Namespace: afterReconcile.GetNamespace()}, afterReconcile)
			Expect(err).NotTo(HaveOccurred())
			Expect(afterReconcile.GetResourceVersion()).NotTo(Equal(written.GetResourceVersion()))

		})
	})
	Context("create manifest", func() {
		It("creates resources from a helm chart", func() {
			values := `
mixer:
  enabled: true #should install mixer

`
			manifests, err := helmchart.RenderManifests(
				context.TODO(),
				"https://storage.googleapis.com/supergloo-charts/istio-1.0.6.tgz",
				values,
				"aaa",
				ns,
				"",
			)
			Expect(err).NotTo(HaveOccurred())

			restCfg, err := kubeutils.GetConfig("", "")
			Expect(err).NotTo(HaveOccurred())
			cache := NewCache()
			err = cache.Init(context.TODO(), restCfg)
			Expect(err).NotTo(HaveOccurred())
			inst, err := NewKubeInstaller(restCfg, cache, nil)
			Expect(err).NotTo(HaveOccurred())

			resources, err := manifests.ResourceList()
			Expect(err).NotTo(HaveOccurred())

			uniqueLabels := map[string]string{"unique": "setoflabels"}
			err = inst.ReconcileResources(context.TODO(), NewReconcileParams(ns, resources, uniqueLabels, false))
			Expect(err).NotTo(HaveOccurred())

			genericClient, err := client.New(restCfg, client.Options{})
			Expect(err).NotTo(HaveOccurred())
			// expect each resource to exist
			for _, resource := range resources {
				err := genericClient.Get(context.TODO(), client.ObjectKey{resource.GetNamespace(), resource.GetName()}, resource)
				Expect(err).NotTo(HaveOccurred())
				if resource.Object["kind"] == "Deployment" || resource.Object["kind"] == "Job" {
					// ensure all deployments have at least one ready replica and all jobs are complete
					structured, err := kuberesource.ConvertUnstructured(resource)
					Expect(err).NotTo(HaveOccurred())
					switch t := structured.(type) {
					case *appsv1.Deployment:
						Expect(t.Status.ReadyReplicas).To(BeNumerically(">=", 1))
					case *extensionsv1beta1.Deployment:
						Expect(t.Status.ReadyReplicas).To(BeNumerically(">=", 1))
					case *appsv1beta2.Deployment:
						Expect(t.Status.ReadyReplicas).To(BeNumerically(">=", 1))
					case *batchv1.Job:
						Expect(t.Status.CompletionTime).NotTo(BeNil(), "no complete time for job %v", t.Name)
						var completeCondition batchv1.JobCondition
						for _, condition := range t.Status.Conditions {
							if condition.Type == batchv1.JobComplete {
								completeCondition = condition
								break
							}
						}
						Expect(completeCondition).NotTo(BeNil(), "no complete condition for job %v", t.Name)
					}
				}

			}

			Expect(ListAllCachedValues(context.TODO(), "unique", inst)).To(ConsistOf("setoflabels"))
			Expect(ListAllCachedValues(context.TODO(), "unknown", inst)).To(BeEmpty())

			// expect the mixer deployments to be created
			_, err = kubeClient.AppsV1().Deployments(ns).Get("istio-policy", v1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			_, err = kubeClient.AppsV1().Deployments(ns).Get("istio-telemetry", v1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())

			err = inst.PurgeResources(context.TODO(), uniqueLabels)
			Expect(err).NotTo(HaveOccurred())

			// uninstalled
			_, err = kubeClient.AppsV1().Deployments(ns).Get("istio-policy", v1.GetOptions{})
			Expect(err).To(HaveOccurred())
			_, err = kubeClient.AppsV1().Deployments(ns).Get("istio-telemetry", v1.GetOptions{})
			Expect(err).To(HaveOccurred())

			// pods deleted
			Eventually(func() []kubev1.Pod {
				pods, err := kubeClient.CoreV1().Pods(ns).List(v1.ListOptions{LabelSelector: labels.SelectorFromSet(labels.Set{"app": "telemetry"}).String()})
				Expect(err).NotTo(HaveOccurred())
				return pods.Items
			}, time.Minute).Should(HaveLen(0))

			Expect(ListAllCachedValues(context.TODO(), "unique", inst)).To(BeEmpty())
		})
	})

	Context("create a manifest when a resource already exists", func() {
		var (
			res  *unstructured.Unstructured
			inst *KubeInstaller
		)

		getResourceVersion := func() string {
			objectKey := client.ObjectKey{Namespace: res.GetNamespace(), Name: res.GetName()}
			resCopy := res.DeepCopy()
			err := inst.client.Get(context.TODO(), objectKey, resCopy)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			return resCopy.GetResourceVersion()
		}

		makeInstaller := func(policy CreationPolicy) *KubeInstaller {
			cache := NewCache()
			cache.access.Unlock()
			cache.resources = make(kuberesource.UnstructuredResourcesByKey)

			restCfg, err := kubeutils.GetConfig("", "")
			Expect(err).NotTo(HaveOccurred())
			inst, err := NewKubeInstaller(restCfg, cache, &KubeInstallerOptions{CreationPolicy: policy})
			Expect(err).NotTo(HaveOccurred())
			return inst
		}

		Context("with CreationPolicy_UpdateOnExists", func() {
			BeforeEach(func() {
				inst = makeInstaller(CreationPolicy_UpdateOnExists)
			})
			Context("resource has no immutable fields", func() {
				BeforeEach(func() {
					res = makeConfigmap("a", "b", ns)
				})
				It("updates the resource instead", func() {
					ctx := context.TODO()

					// call creation function to create object
					err := inst.getCreationFunction(ctx, res)()
					Expect(err).NotTo(HaveOccurred())

					// record resource version after create
					oldRv := getResourceVersion()

					// modify object
					res.SetLabels(func() map[string]string {
						labels := res.GetLabels()
						labels["new"] = "label"
						return labels
					}())

					// create again to update
					err = inst.getCreationFunction(ctx, res)()
					Expect(err).NotTo(HaveOccurred())

					// record resource version after update
					newRv := getResourceVersion()

					Expect(oldRv).NotTo(Equal(newRv))
				})
			})
			Context("resource has immutable fields", func() {
				BeforeEach(func() {
					res = makeService(ns)
				})
				It("returns an ImmutableResource error", func() {
					ctx := context.TODO()

					// call creation function to create object
					err := inst.getCreationFunction(ctx, res)()
					Expect(err).NotTo(HaveOccurred())

					// modify object
					res.SetLabels(func() map[string]string {
						labels := res.GetLabels()
						if labels == nil {
							labels = map[string]string{}
						}
						labels["new"] = "label"
						return labels
					}())

					// create again to update
					err = inst.getCreationFunction(ctx, res)()
					Expect(err).To(HaveOccurred())
					Expect(kubeerrutils.IsImmutableErr(err)).To(BeTrue())
				})
			})
		})
		Context("with CreationPolicy_ReturnErrors", func() {
			BeforeEach(func() {
				inst = makeInstaller(CreationPolicy_ReturnErrors)
			})
			It("returns AlreadyExists error", func() {
				res = makeConfigmap("any", "thing", ns)
				ctx := context.TODO()

				// call creation function to create object
				err := inst.getCreationFunction(ctx, res)()
				Expect(err).NotTo(HaveOccurred())

				// modify object
				res.SetLabels(func() map[string]string {
					labels := res.GetLabels()
					if labels == nil {
						labels = map[string]string{}
					}
					labels["new"] = "label"
					return labels
				}())

				// create again to update
				err = inst.getCreationFunction(ctx, res)()
				Expect(err).To(HaveOccurred())
				Expect(errors.IsAlreadyExists(err)).To(BeTrue())
			})
		})
		Context("with CreationPolicy_ForceUpdateOnExists", func() {
			BeforeEach(func() {
				inst = makeInstaller(CreationPolicy_ForceUpdateOnExists)
			})
			It("re-creates resources with immutable fields", func() {
				res = makeService(ns)
				ctx := context.TODO()

				// call creation function to create object
				err := inst.getCreationFunction(ctx, res)()
				Expect(err).NotTo(HaveOccurred())

				// record resource version after create
				oldRv := getResourceVersion()

				// modify object
				res.SetLabels(func() map[string]string {
					labels := res.GetLabels()
					if labels == nil {
						labels = map[string]string{}
					}
					labels["new"] = "label"
					return labels
				}())

				// create again to update
				err = inst.getCreationFunction(ctx, res)()
				Expect(err).NotTo(HaveOccurred())

				// record resource version after update
				newRv := getResourceVersion()

				Expect(oldRv).NotTo(Equal(newRv))
			})
		})
	})
})

func makeConfigmap(unique, randomLabel, ns string) *unstructured.Unstructured {
	grafanaCfg := &unstructured.Unstructured{}
	err := json.Unmarshal([]byte(fmt.Sprintf(`{"apiVersion":"v1","data":{"dashboardproviders.yaml":"apiVersion:1\nproviders:\n- disableDeletion: false\n  folder: istio\n  name: istio\n  options:\n    path:/var/lib/grafana/dashboards/istio\n  orgId: 1\n  type: file\n","datasources.yaml":"apiVersion:1\ndatasources:\n- access: proxy\n  editable: true\n  isDefault: true\n  jsonData:\n    timeInterval:5s\n  name: Prometheus\n  orgId: 1\n  type: prometheus\n  url: http://prometheus:9090\n"},"kind":"ConfigMap","metadata":{"labels":{"app":"istio-grafana","chart":"grafana-1.0.6","heritage":"Tiller","istio":"grafana","release":"istio","v1-install":"supergloo-system.istio"},"name":"istio-grafana","namespace":"%v"}}`, ns)), &grafanaCfg.Object)
	Expect(err).NotTo(HaveOccurred())

	labels := grafanaCfg.GetLabels()
	labels[unique] = randomLabel
	grafanaCfg.SetLabels(labels)
	err = setInstallationAnnotation(grafanaCfg)
	Expect(err).NotTo(HaveOccurred())
	return grafanaCfg
}

func makeService(ns string) *unstructured.Unstructured {
	basicServiceCfg := &unstructured.Unstructured{}
	err := json.Unmarshal([]byte(fmt.Sprintf(`{ "apiVersion": "v1",
    "kind": "Service",
    "metadata": {
        "name": "kubernetes",
        "namespace": "%s"
    },
    "spec": {
        "ports": [
            {
                "name": "https",
                "port": 443,
                "protocol": "TCP",
                "targetPort": 443
            }
        ]
    }
}`, ns)), &basicServiceCfg.Object)
	Expect(err).NotTo(HaveOccurred())

	labels := basicServiceCfg.GetLabels()
	basicServiceCfg.SetLabels(labels)
	err = setInstallationAnnotation(basicServiceCfg)
	Expect(err).NotTo(HaveOccurred())
	return basicServiceCfg
}
