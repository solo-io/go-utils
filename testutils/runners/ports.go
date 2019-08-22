package runners

import "github.com/onsi/ginkgo"

func AllocateParallelPort(basePort int) int {
	return basePort + (ginkgo.GinkgoParallelNode()-1)*20
}
