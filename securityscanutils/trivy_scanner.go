package securityscanutils

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/go-utils/osutils/executils"
)

// Status code returned by Trivy if a vulnerability is found
const VulnerabilityFoundStatusCode = 52

// Runs trivy scan command
// Returns (trivy scan ran successfully, vulnerabilities found, error running trivy scan)
func RunTrivyScan(image, version, templateFile, output string) (bool, bool, error) {
	// Ensure Trivy is installed and on PATH
	_, err := exec.LookPath("trivy")
	if err != nil {
		return false, false, eris.Wrap(err, "trivy is not on PATH, make sure that the trivy v0.18 is installed and on PATH")
	}
	trivyScanArgs := []string{"image",
		// Trivy will return a specific status code (which we have specified) if a vulnerability is found
		"--exit-code", strconv.Itoa(VulnerabilityFoundStatusCode),
		"--severity", "HIGH,CRITICAL",
		"--format", "template",
		"--template", "@" + templateFile,
		"--output", output,
		image}
	// Execute the trivy scan, with retries and sleep's between each retry
	// This can occur due to connectivity issues or epehemeral issues with
	// the registery. For example sometimes quay has issues providing a given layer
	// This leads to a total wait time of up to 110 seconds outside of the base
	// operation. This timing is in the same ballpark as what k8s finds sensible
	out, statusCode, err := executeTrivyScanWithRetries(
		trivyScanArgs, 5,
		func(attempt int) { time.Sleep(time.Duration((attempt^2)*2) * time.Second) },
	)

	// Check if a vulnerability has been found
	vulnFound := statusCode == VulnerabilityFoundStatusCode
	// err will be non-nil if there is a non-zero status code
	// so if the status code is the special "vulnerability found" status code,
	// we don't want to report it as a regular error
	if !vulnFound && err != nil {
		// delete empty trivy output file that may have been created
		_ = os.Remove(output)
		// swallow error if image is not found error, so that we can continue scanning releases
		// even if some releases failed and we didn't publish images for those releases
		// this error used to happen if a release was a pre-release and therefore images
		// weren't pushed to the container registry.
		// we have since filtered out non-release images from being scanned so this warning
		// shouldn't occur, but leaving here in case there was another edge case we missed
		if IsImageNotFoundErr(string(out)) {
			log.Warnf("image %s not found for version %s", image, version)
			return false, false, nil
		}
		return false, false, eris.Wrapf(err, "error running trivy scan on image %s, version %s, Logs: \n%s", image, version, string(out))
	}
	return true, vulnFound, nil
}

func executeTrivyScanWithRetries(trivyScanArgs []string, retryCount int,
	backoffStrategy func(int)) ([]byte, int, error) {
	if retryCount == 0 {
		retryCount = 5
	}
	if backoffStrategy == nil {
		backoffStrategy = func(attempt int) {
			time.Sleep(time.Second)
		}
	}

	var (
		out        []byte
		statusCode int
		err        error
	)

	for attempt := 0; attempt < retryCount; attempt++ {
		trivyScanCmd := exec.Command("trivy", trivyScanArgs...)
		out, statusCode, err = executils.CombinedOutputWithStatus(trivyScanCmd)

		// If we receive the expected status code, the scan completed, don't retry
		if statusCode == VulnerabilityFoundStatusCode {
			return out, statusCode, nil
		}

		// If there is no error, don't retry
		if err == nil {
			return out, statusCode, err
		}

		// If there is no image, don't retry
		if IsImageNotFoundErr(string(out)) {
			return out, statusCode, err
		}

		backoffStrategy(attempt)
	}
	return out, statusCode, err
}

func IsImageNotFoundErr(logs string) bool {
	return strings.Contains(logs, "No such image: ")
}
