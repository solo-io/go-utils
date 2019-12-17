package test

import (
	"context"
	"log"
	"os"

	"github.com/solo-io/go-utils/configutils"
	"github.com/solo-io/go-utils/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	kube2 "github.com/solo-io/go-utils/testutils/kube"
	kubeerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("ConfigTest", func() {

	var (
		configMapNamespace string
		configClient       configutils.ConfigClient
	)

	const (
		configMapName = "test-config"
		configKey     = "config.yaml"
	)

	getDefaultConfig := func() *GetApplicationDetailsRequest {
		return &GetApplicationDetailsRequest{
			ApplicationName: "foo",
			RegistryName:    "bar",
		}
	}

	Context("Kube Integration with config map client", func() {
		if os.Getenv("RUN_KUBE_TESTS") != "1" {
			log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
			return
		}

		var configMapClient configutils.ConfigMapClient

		BeforeEach(func() {
			rand := testutils.RandString(8)
			configMapNamespace = "test-" + rand
			configMapClient = configutils.NewConfigMapClient(kube2.MustKubeClient())
			configClient = configutils.NewConfigClient(configMapClient, configMapNamespace, configMapName, configKey, getDefaultConfig())
			kube2.MustCreateNs(configMapNamespace)
		})

		AfterEach(func() {
			kube2.MustDeleteNs(configMapNamespace)
		})

		It("works by default", func() {
			actual := GetApplicationDetailsRequest{}
			err := configClient.GetConfig(context.TODO(), &actual)
			Expect(err).NotTo(HaveOccurred())
			expected := getDefaultConfig()
			Expect(actual).To(BeEquivalentTo(*expected))
			expected.ApplicationName = "updated"
			err = configClient.SetConfig(context.TODO(), expected)
			Expect(err).NotTo(HaveOccurred())
			err = configClient.GetConfig(context.TODO(), &actual)
			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(BeEquivalentTo(*expected))
		})
	})

	Context("Unit tests for errors", func() {

		var mockConfigMapClient configutils.MockConfigMapClient
		testErr := errors.Errorf("test")

		BeforeEach(func() {
			configMapNamespace = "test"
			mockConfigMapClient = configutils.MockConfigMapClient{}
			configClient = configutils.NewConfigClient(&mockConfigMapClient, configMapNamespace, configMapName, configKey, getDefaultConfig())
		})

		It("errors on get when config map client errors", func() {
			mockConfigMapClient.GetError = testErr
			err := configClient.GetConfig(context.TODO(), &GetApplicationDetailsRequest{})
			Expect(err.Error()).To(BeEquivalentTo(configutils.ErrorLoadingExistingConfig(testErr).Error()))
		})

		It("errors on set when config map client errors", func() {
			mockConfigMapClient.SetError = testErr
			err := configClient.SetConfig(context.TODO(), getDefaultConfig())
			Expect(err.Error()).To(BeEquivalentTo(configutils.ErrorUpdatingConfig(testErr).Error()))
		})

		It("errors on get when config map contains invalid data", func() {
			mockConfigMapClient.Data = map[string]string{configKey: "dummy"}
			err := configClient.GetConfig(context.TODO(), &GetApplicationDetailsRequest{})
			unmarshalErr := errors.Errorf("json: cannot unmarshal string into Go value of type map[string]json.RawMessage")
			Expect(err.Error()).To(BeEquivalentTo(configutils.ErrorUnmarshallingConfig(unmarshalErr).Error()))
		})

		It("errors on get when default config can't be set", func() {
			mockConfigMapClient.GetError = kubeerr.NewNotFound(schema.GroupResource{}, "name")
			mockConfigMapClient.SetError = testErr
			err := configClient.GetConfig(context.TODO(), &GetApplicationDetailsRequest{})
			Expect(err.Error()).To(BeEquivalentTo(configutils.ErrorSettingDefaultConfig(testErr).Error()))
		})
	})
})
