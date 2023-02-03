package runners

import "github.com/onsi/ginkgo/v2"

func AllocateParallelPort(basePort int) int {
    return basePort + (ginkgo.GinkgoParallelNode()-1)*20
}
