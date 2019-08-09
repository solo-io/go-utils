package kube

import (
	"fmt"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/go-utils/testutils/helper"
	"strconv"
	"strings"
)

const maxLogLines = 250

// The PrintKubeDump fail handler eases debugging failed regression/e2e tests by dumping all
// kube pod logs and statuses in the install namespace
func PrintKubeDump(testHelper *helper.SoloTestHelper) {
	var logs strings.Builder
	podsList, err := testutils.KubectlOut("get", "pods", "--no-headers", "-n", testHelper.InstallNamespace,
		"-o", "custom-columns=:metadata.name")
	if err != nil {
		fmt.Println("PrintKubeDump - error getting pods: ", err.Error())
		return
	}
	unfilteredPods := strings.Split(podsList, "\n")
	pods := make([]string, 0, len(unfilteredPods))
	for _, pod := range unfilteredPods {
		if pod != "" {
			pods = append(pods, pod)
		}
	}

	for _, pod := range pods {
		podStatus, err := testutils.KubectlOut("get", "pods", pod, "-n", testHelper.InstallNamespace,
			"-o", "go-template=\"{{range .status.containerStatuses}}{{.state}}{{end}}\"")
		if err != nil {
			logs.WriteString("unable to get state for pod: " + pod + "\n")
		} else {
			logs.WriteString("state for pod: " + pod + ": " + podStatus + "\n")
		}

		podLogs, err := testutils.KubectlOut("logs", pod, "-n", testHelper.InstallNamespace,
			"--all-containers", "--tail", strconv.Itoa(maxLogLines))
		if err != nil {
			logs.WriteString("error getting logs for " + pod + ": " + err.Error() + "\n")

			prevPodLogs, err := testutils.KubectlOut("logs", pod, "-n", testHelper.InstallNamespace,
				"--all-containers", "--tail", strconv.Itoa(maxLogLines), "-p")
			if err != nil {
				logs.WriteString("error getting previous logs for " + pod + ": " + err.Error() + "\n\n")
				continue
			}
			logs.WriteString("--- previous logs for " + pod + " ---\n")
			logs.WriteString(prevPodLogs)
			logs.WriteString("--- end previous logs for " + pod + " ---\n\n")
			continue
		}
		logs.WriteString("--- logs for " + pod + " ---\n")
		logs.WriteString(podLogs)
		logs.WriteString("--- end logs for " + pod + " ---\n\n")
	}
	if len(pods) > 0 {
		fmt.Printf("\n********** Kube Dump **********\n\n" + logs.String() +
			"\n******** End Kube Dump ********\n\n")
	} else {
		fmt.Println("PrintKubeDump - No pods found, thus no logs to print")
	}
}