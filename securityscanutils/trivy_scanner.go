package securityscanutils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/rotisserie/eris"
)

// Status code returned by Trivy if a vulnerability is found
const VulnerabilityFoundStatusCode = 52

type CmdExecutor func(cmd *exec.Cmd) ([]byte, int, error)

type TrivyScanner struct {
	executeCommand      CmdExecutor
	scanBackoffStrategy func(int)
	scanMaxRetries      int
}

func NewTrivyScanner(executeCommand CmdExecutor) *TrivyScanner {
	return &TrivyScanner{
		executeCommand:      executeCommand,
		scanBackoffStrategy: func(attempt int) { time.Sleep(time.Duration((attempt^2)*2) * time.Second) },
		scanMaxRetries:      5,
	}
}

func (t *TrivyScanner) ScanImage(ctx context.Context, image, templateFile, output string) (bool, bool, error) {
	trivyScanArgs := []string{"image",
		// Trivy will return a specific status code (which we have specified) if a vulnerability is found
		"--exit-code", strconv.Itoa(VulnerabilityFoundStatusCode),
		"--severity", "HIGH,CRITICAL",
		"--format", "template",
		"--template", "@" + templateFile,
		"--output", output,
		image}

	fmt.Println(os.Getwd())
	panic(nil)
	// Execute the trivy scan, with retries and sleep's between each retry
	// This can occur due to connectivity issues or epehemeral issues with
	// the registry. For example sometimes quay has issues providing a given layer
	// This leads to a total wait time of up to 110 seconds outside of the base
	// operation. This timing is in the same ballpark as what k8s finds sensible
	scanCompleted, vulnerabilityFound, err := t.executeScanWithRetries(ctx, trivyScanArgs)

	if !scanCompleted {
		// delete the empty trivy output file that may have been created
		_ = os.Remove(output)
	}

	return scanCompleted, vulnerabilityFound, err
}

// executeScanWithRetries executes a trivy command (with retries and backoff)
// and returns a tuple of (scanCompleted, vulnerabilitiesFound, error)
func (t *TrivyScanner) executeScanWithRetries(ctx context.Context, scanArgs []string) (bool, bool, error) {
	logger := contextutils.LoggerFrom(ctx)
	var (
		out        []byte
		statusCode int
		err        error
	)
	attemptStart := time.Now()
	for attempt := 0; attempt < t.scanMaxRetries; attempt++ {
		trivyScanCmd := exec.Command("trivy", scanArgs...)
		imageUri := scanArgs[11]
		out, statusCode, err = t.executeCommand(trivyScanCmd)

		// If we receive the expected status code, the scan completed, don't retry
		if statusCode == VulnerabilityFoundStatusCode {
			logger.Debugf("Trivy found vulnerabilies in after %s on %s", time.Since(attemptStart).String(), imageUri)
			return true, true, nil
		}

		// If there is no error, the scan completed and no vulnerability was found, don't retry
		if err == nil {
			logger.Debugf("Trivy returned %d after %s on %s", statusCode, time.Since(attemptStart).String(), imageUri)
			return true, false, err
		}

		// If there is no image, don't retry
		if IsImageNotFoundErr(string(out)) {
			logger.Warnf("Trivy scan with args [%v] produced image not found error", scanArgs)

			// swallow error if image is not found error, so that we can continue scanning releases
			// even if some releases failed and we didn't publish images for those releases
			// this error used to happen if a release was a pre-release and therefore images
			// weren't pushed to the container registry.
			// we have since filtered out non-release images from being scanned so this warning
			// shouldn't occur, but leaving here in case there was another edge case we missed
			return false, false, nil
		}

		//This backoff strategy is intended to handle network issues(i.e. an http 5xx error)
		t.scanBackoffStrategy(attempt)
	}
	// We only reach here if we exhausted our retries
	return false, false, eris.Errorf("Trivy scan with args [%v] did not complete after %d attempts", scanArgs, t.scanMaxRetries)
}

func IsImageNotFoundErr(logs string) bool {
	return strings.Contains(logs, "No such image: ")
}
