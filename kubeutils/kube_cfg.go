package kubeutils

import (
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// GetConfig gets the kubernetes client config
func GetConfig(masterURL, kubeconfigPath string) (*rest.Config, error) {
	return GetConfigWithContext(masterURL, kubeconfigPath, "")
}

func GetConfigWithContext(masterURL, kubeconfigPath, context string) (*rest.Config, error) {
	envVarName := clientcmd.RecommendedConfigPathEnvVar
	if kubeconfigPath == "" && masterURL == "" && os.Getenv(envVarName) == "" {
		// Neither kubeconfig nor master URL nor envVarName(KUBECONFIG) was specified. Using the inClusterConfig.
		kubeconfig, err := rest.InClusterConfig()
		if err == nil {
			return kubeconfig, nil
		}
		// error creating inClusterConfig, falling back to default config
	}

	if kubeconfigPath != "" {
		if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
			// the specified kubeconfig does not exist so fallback to default
			kubeconfigPath = ""
		}
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfigPath
	configOverrides := &clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: masterURL}}
	if context != "" {
		configOverrides.CurrentContext = context
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
}

// GetConfig gets the kubernetes client config
func GetKubeConfig(masterURL, kubeconfigPath string) (*clientcmdapi.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfigPath
	configOverrides := &clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: masterURL}}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ConfigAccess().GetStartingConfig()
}
