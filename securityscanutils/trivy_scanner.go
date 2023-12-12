package securityscanutils

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/solo-io/go-utils/contextutils"

	"github.com/rotisserie/eris"
)

// Status code returned by Trivy if a vulnerability is found
const VulnerabilityFoundStatusCode = 52

var imageNotFoundError = errors.New("❗IMAGE MISSING UNEXPECTEDLY❗")

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
			logger.Debugf("Trivy found vulnerabilies after %s in %s", time.Since(attemptStart).String(), imageUri)
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

			// Indicate the scan has not yet completed and no vulnerability was found but there was an imageNotFoundError.
			// The upstream handler should check specifically for this error to ensure that the remaining images for
			// the specified version are scanned.
			return false, false, imageNotFoundError
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
