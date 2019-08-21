package helper

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/avast/retry-go"
	"github.com/solo-io/go-utils/testutils/kube"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/go-utils/testutils/exec"
	"k8s.io/helm/pkg/repo"
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
	DeployTestRunner:      true,
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
	// The namespace gloo (and the test runner) will be installed to. If empty, will use the helm chart version.
	InstallNamespace string
	// Name of the glooctl executable
	GlooctlExecName string
	// If provided, the licence key to install the enterprise version of Gloo
	LicenseKey string
	// Determines whether the test runner pod gets deployed
	DeployTestRunner bool

	// If true, glooctl will be run with a -v flag
	Verbose bool

	// The version of the Helm chart
	version string
}

// This helper is meant to provide a standard way of deploying Gloo/GlooE to a k8s cluster during tests.
// It assumes that build and test assets are present in the `_output` and `_test` directories (these are configurable).
// Specifically, it expects the glooctl executable in the BuildAssetDir and a helm chart in TestAssetDir.
// It also assumes that a kubectl executable is on the PATH.
type SoloTestHelper struct {
	*TestConfig
	TestRunner
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

	// Default the install namespace to the chart version.
	// Currently the test chart version built in CI contains the build id, so the namespace will be unique).
	if testConfig.InstallNamespace == "" {
		testConfig.InstallNamespace = version
	}

	// Optionally, initialize a test runner
	var testRunner *testRunner
	if testConfig.DeployTestRunner {
		testRunner, err = NewTestRunner(testConfig.InstallNamespace)
		if err != nil {
			return nil, errors.Wrapf(err, "initializing testrunner")
		}
	}

	return &SoloTestHelper{
		TestConfig: &testConfig,
		TestRunner: testRunner,
	}, nil
}

// Return the version of the Helm chart
func (h *SoloTestHelper) ChartVersion() string {
	return h.version
}

type InstallOption func(*InstallOptions)

type InstallOptions struct {
	GlooctlCommand []string
	Verbose        bool
}

func ExtraArgs(args ...string) func(*InstallOptions) {
	return func(io *InstallOptions) {
		io.GlooctlCommand = append(io.GlooctlCommand, args...)
	}
}

// Installs Gloo (and, optionally, the test runner)
func (h *SoloTestHelper) InstallGloo(deploymentType string, timeout time.Duration, options ...InstallOption) error {
	log.Printf("installing gloo in [%s] mode to namespace [%s]", deploymentType, h.InstallNamespace)
	glooctlCommand := []string{
		filepath.Join(h.BuildAssetDir, h.GlooctlExecName),
		"install", deploymentType,
		"-n", h.InstallNamespace,
		"-f", filepath.Join(h.TestAssetDir, h.HelmChartName+"-"+h.version+".tgz"),
	}
	if h.Verbose {
		glooctlCommand = append(glooctlCommand, "-v")
	}
	if h.LicenseKey != "" {
		glooctlCommand = append(glooctlCommand, "--license-key", h.LicenseKey)
	}
	io := &InstallOptions{
		GlooctlCommand: glooctlCommand,
		Verbose:        true,
	}
	for _, opt := range options {
		opt(io)
	}
	if err := exec.RunCommand(h.RootDir, io.Verbose, io.GlooctlCommand...); err != nil {
		return errors.Wrapf(err, "error while installing gloo")
	}

	if h.TestRunner != nil {
		if err := waitForDefaultServiceAccount(h.InstallNamespace); err != nil {
			return errors.Wrapf(err, "waiting for default service account")
		}
		if err := h.TestRunner.Deploy(timeout); err != nil {
			return errors.Wrapf(err, "deploying testrunner")
		}
	}
	return nil
}

func waitForDefaultServiceAccount(installNamespace string) error {
	kubeClient := kube.MustKubeClient()
	getDefaultServiceAccount := func() error {
		_, err := kubeClient.CoreV1().ServiceAccounts(installNamespace).Get("default", v1.GetOptions{})
		return err
	}
	return retry.Do(getDefaultServiceAccount)
}

// passes the --all flag to glooctl uninstall
func (h *SoloTestHelper) UninstallGlooAll() error {
	return h.uninstallGloo(true)
}

// does not pass the --all flag to glooctl uninstall
func (h *SoloTestHelper) UninstallGloo() error {
	return h.uninstallGloo(false)
}

func (h *SoloTestHelper) uninstallGloo(all bool) error {
	if h.TestRunner != nil {
		log.Debugf("terminating %s...", TestrunnerName)
		if err := h.TestRunner.Terminate(); err != nil {
			// Just log a warning, we don't want to fail
			log.Warnf("error terminating %s", TestrunnerName)
		}
	}

	log.Printf("uninstalling gloo...")
	cmdArgs := []string{
		filepath.Join(h.BuildAssetDir, h.GlooctlExecName), "uninstall", "-n", h.InstallNamespace, "--delete-namespace",
	}
	if all {
		cmdArgs = append(cmdArgs, "--all")
	}
	return exec.RunCommand(h.RootDir, true, cmdArgs...)
}

// Parses the Helm index file and returns the version of the chart.
func getChartVersion(config TestConfig) (string, error) {

	// Find helm index file in test asset directory
	helmIndexFile := filepath.Join(config.RootDir, config.TestAssetDir, config.HelmRepoIndexFileName)
	helmIndex, err := repo.LoadIndexFile(helmIndexFile)
	if err != nil {
		return "", errors.Wrapf(err, "parsing Helm index file")
	}
	log.Printf("found Helm index file at: %s", helmIndexFile)

	// Read and return version from helm index file
	if chartVersions, ok := helmIndex.Entries[config.HelmChartName]; !ok {
		return "", errors.Errorf("index file does not contain entry with key: %s", config.HelmChartName)
	} else if len(chartVersions) == 0 || len(chartVersions) > 1 {
		return "", errors.Errorf("expected a single entry with name [%s], found: %v", config.HelmChartName, len(chartVersions))
	} else {
		version := chartVersions[0].Version
		log.Printf("version of [%s] Helm chart is: %s", config.HelmChartName, version)
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
