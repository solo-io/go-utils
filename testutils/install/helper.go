package install

import (
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/logger"
	"github.com/solo-io/go-utils/testutils/exec"
	"k8s.io/helm/pkg/repo"
	"os"
	"path/filepath"
	"runtime"
)

const (
	GATEWAY = "gateway"
	INGRESS = "ingress"
	KNATIVE = "knative"
)

// Default test configuration
var defaults = TestConfig{
	TestAssetDir:          "_test",
	BuildAssetDir:         "_output",
	HelmRepoIndexFileName: "index.yaml",
	GlooctlExecName:       "glooctl-" + runtime.GOOS + "-amd64",
}

// Function to provide/override test configuration. Default values will be passed in.
type TestConfigFunc func(defaults TestConfig) TestConfig

type TestConfig struct {
	// All relative paths will assume this as the base directory. This is usually the project base directory.
	RootDir string
	// The directory holding the test assets. Must be relative to RootDir.
	TestAssetDir string
	// The directory holding the build assets. Must be relative to RootDir.
	BuildAssetDir string
	// Helm chart name
	HelmChartName string
	// Name of the helm index file name
	HelmRepoIndexFileName string
	// Name of the glooctl executable
	GlooctlExecName string
	// If provided, the licence key to install the enterprise version of Gloo
	LicenseKey string

	// The version of the Helm chart
	version string
}

// This helper is meant to provide a standard way of deploying Gloo/GlooE to a k8s cluster during tests.
// It assumes that build and test assets are present in the `_output` and `_test` directories (these are configurable).
// Specifically, it expects the glooctl executable in the BuildAssetDir and a helm chart in TestAssetDir.
// It also assumes that a kubectl executable is on the PATH.
type SoloTestHelper struct {
	*TestConfig
}

func NewSoloTestHelper(configFunc TestConfigFunc) (*SoloTestHelper, error) {

	// Get and validate test config
	testConfig := defaults
	if configFunc != nil {
		testConfig = configFunc(defaults)
	}
	if err := validateConfig(testConfig); err != nil {
		return nil, errors.Wrapf(err, "test config validation failed")
	}

	// Get chart version
	version, err := getChartVersion(testConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "getting Helm chart version")
	}
	testConfig.version = version

	return &SoloTestHelper{&testConfig}, nil
}

// Return the version of the Helm chart
func (h *SoloTestHelper) ChartVersion() string {
	return h.version
}

func (h *SoloTestHelper) InstallGloo(deploymentType, namespace string) error {
	glooctlCommand := []string{
		filepath.Join(h.BuildAssetDir, h.GlooctlExecName),
		"install", deploymentType,
		"-n", namespace,
		"-f", filepath.Join(h.TestAssetDir, h.HelmChartName+"-"+h.version+".tgz"),
	}
	if h.LicenseKey != "" {
		glooctlCommand = append(glooctlCommand, "--license-key", h.LicenseKey)
	}
	return exec.RunCommand(h.RootDir, true, glooctlCommand...)
}

// Parses the Helm index file and returns the version of the chart.
func getChartVersion(config TestConfig) (string, error) {

	// Find helm index file in test asset directory
	helmIndexFile := filepath.Join(config.RootDir, config.TestAssetDir, config.HelmRepoIndexFileName)
	helmIndex, err := repo.LoadIndexFile(helmIndexFile)
	if err != nil {
		return "", errors.Wrapf(err, "parsing Helm index file")
	}
	logger.Debugf("found Helm index file at: %s", helmIndexFile)

	// Read and return version from helm index file
	if chartVersions, ok := helmIndex.Entries[config.HelmChartName]; !ok {
		return "", errors.Errorf("index file does not contain entry with key: %s", config.HelmChartName)
	} else if len(chartVersions) == 0 || len(chartVersions) > 1 {
		return "", errors.Errorf("expected a single entry with name [%s], found: %v", config.HelmChartName, len(chartVersions))
	} else {
		version := chartVersions[0].Version
		logger.Debugf("version of [%s] Helm chart is: %s", config.HelmChartName, version)
		return version, nil
	}
}

func validateConfig(config TestConfig) error {
	if err := validateDir(config.RootDir); err != nil {
		return err
	}
	if err := validateDir(filepath.Join(config.RootDir, config.TestAssetDir)); err != nil {
		return err
	}
	if err := validateDir(filepath.Join(config.RootDir, config.BuildAssetDir)); err != nil {
		return err
	}
	return nil
}

func validateDir(dir string) error {
	if stat, err := os.Stat(dir); err != nil {
		return errors.Wrapf(err, "finding directory: %s", dir)
	} else if !stat.IsDir() {
		return errors.Errorf("expected a directory. Got: %s", dir)
	}
	return nil
}
