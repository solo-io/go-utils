package testutils

import (
	"os"

	"github.com/solo-io/go-utils/log"
)

func AreTestsDisabled() bool {
	if os.Getenv("RUN_KUBE2E_TESTS") != "1" {
		log.Warnf("This test requires a running kubernetes cluster and is disabled by default. " +
			"To enable, set RUN_KUBE2E_TESTS=1 in your env.")
		return true
	}
	return false
}
