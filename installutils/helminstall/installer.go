package helminstall

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/kubeutils"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"sigs.k8s.io/yaml"
)

//go:generate mockgen -destination mocks/mock_helm_installer.go -source ./installer.go

var (
	ReleaseAlreadyInstalledErr = func(name, namespace string) error {
		return eris.Errorf("The helm release you are trying to install (%s) appears"+
			" to already exist in %s", name, namespace)
	}
)

type Installer interface {
	Install(installerConfig *InstallerConfig) error
}

type InstallerConfig struct {
	DryRun           bool
	CreateNamespace  bool
	Verbose          bool
	InstallNamespace string
	ReleaseName      string
	// the uri to the helm chart, can either be a local file or a valid http/https link
	ReleaseUri  string
	ValuesFiles []string
	ExtraValues map[string]interface{}

	PreInstallMessage  string
	PostInstallMessage string
}

type installer struct {
	helmClient   HelmClient
	kubeNsClient NamespaceCLient
	out          io.Writer
}

func MustInstaller() Installer {
	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		log.Fatal(err)
	}
	client := kubernetes.NewForConfigOrDie(cfg)
	return NewInstaller(DefaultHelmClient(), client.CoreV1().Namespaces(), os.Stdout)
}

// visible for testing
func NewInstaller(helmClient HelmClient, kubeNsClient NamespaceCLient, outputWriter io.Writer) Installer {
	return &installer{
		helmClient:   helmClient,
		kubeNsClient: kubeNsClient,
		out:          outputWriter,
	}
}

func (i *installer) Install(installerConfig *InstallerConfig) error {
	namespace := installerConfig.InstallNamespace
	releaseName := installerConfig.ReleaseName
	if !installerConfig.DryRun {
		if releaseExists, err := i.helmClient.ReleaseExists(namespace, releaseName); err != nil {
			return err
		} else if releaseExists {
			return ReleaseAlreadyInstalledErr(releaseName, namespace)
		}
		if installerConfig.CreateNamespace {
			// Create the namespace if it doesn't exist. Helm3 no longer does this.
			i.createNamespace(namespace)
		}
	}

	if !installerConfig.DryRun && installerConfig.PreInstallMessage != "" {
		fmt.Fprintf(i.out, installerConfig.PreInstallMessage)
	} else {
		i.defaultPreInstallMessage(installerConfig)
	}

	helmInstall, helmEnv, err := i.helmClient.NewInstall(namespace, releaseName, installerConfig.DryRun)
	if err != nil {
		return err
	}

	if installerConfig.Verbose {
		fmt.Printf("Looking for chart at %s\n", installerConfig.ReleaseUri)
	}

	chartObj, err := i.helmClient.DownloadChart(installerConfig.ReleaseUri)
	if err != nil {
		return err
	}

	// Merge values provided via the '--values' flag
	valueOpts := &values.Options{
		ValueFiles: installerConfig.ValuesFiles,
	}
	cliValues, err := valueOpts.MergeValues(getter.All(helmEnv))
	if err != nil {
		return err
	}

	// Merge the CLI flag values into the extra values, giving the latter higher precedence.
	// (The first argument to CoalesceTables has higher priority)
	completeValues := chartutil.CoalesceTables(installerConfig.ExtraValues, cliValues)
	if installerConfig.Verbose {
		b, err := json.Marshal(completeValues)
		if err != nil {
			fmt.Fprintf(i.out, "error: %v\n", err)
		}
		y, err := yaml.JSONToYAML(b)
		if err != nil {
			fmt.Fprintf(i.out, "error: %v\n", err)
		}
		fmt.Fprintf(i.out, "Installing the %s chart with the following value overrides:\n%s\n", chartObj.Metadata.Name, string(y))
	}

	rel, err := helmInstall.Run(chartObj, completeValues)
	if err != nil {
		return err
	}
	if !installerConfig.DryRun && installerConfig.PostInstallMessage != "" {
		fmt.Fprintf(i.out, installerConfig.PostInstallMessage)
	} else {
		i.defaultPostInstallMessage(installerConfig)
	}

	if installerConfig.Verbose {
		fmt.Printf("Successfully ran helm install with release %s\n", releaseName)
	}

	if installerConfig.DryRun {
		fmt.Fprintf(i.out, rel.Manifest)
	}

	return nil
}

func (i *installer) createNamespace(namespace string) {
	_, err := i.kubeNsClient.Get(namespace, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		fmt.Fprintf(i.out, "Creating namespace %s... ", namespace)
		if _, err := i.kubeNsClient.Create(&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}); err != nil {
			fmt.Fprintf(i.out, "\nUnable to create namespace %s. Continuing...\n", namespace)
		} else {
			fmt.Fprintf(i.out, "Done.\n")
		}
	} else {
		fmt.Fprintf(i.out, "\nUnable to check if namespace %s exists. Continuing...\n", namespace)
	}

}

func (i *installer) defaultPreInstallMessage(config *InstallerConfig) {
	if config.DryRun {
		return
	}
	fmt.Fprintf(i.out, "Starting helm installation\n")
}

func (i *installer) defaultPostInstallMessage(config *InstallerConfig) {
	if config.DryRun {
		return
	}
	fmt.Fprintf(i.out, "Successful installation!\n")
}

type NamespaceCLient interface {
	Create(ns *corev1.Namespace) (*corev1.Namespace, error)
	Delete(name string, options *metav1.DeleteOptions) error
	Get(name string, options metav1.GetOptions) (*corev1.Namespace, error)
	List(opts metav1.ListOptions) (*corev1.NamespaceList, error)
}

type namespaceClient struct {
	client v1.NamespaceInterface
}

func (n *namespaceClient) Create(ns *corev1.Namespace) (*corev1.Namespace, error) {
	return n.client.Create(ns)
}

func (n *namespaceClient) Delete(name string, options *metav1.DeleteOptions) error {
	return n.client.Delete(name, options)
}

func (n *namespaceClient) Get(name string, options metav1.GetOptions) (*corev1.Namespace, error) {
	return n.client.Get(name, options)
}

func (n *namespaceClient) List(opts metav1.ListOptions) (*corev1.NamespaceList, error) {
	return n.List(opts)
}

func NewNamespaceClient(client v1.NamespaceInterface) *namespaceClient {
	return &namespaceClient{client: client}
}
