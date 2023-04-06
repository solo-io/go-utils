package port

import (
	"sync/atomic"

	. "github.com/onsi/ginkgo/v2"
)

var MaxTests = 1000

type TestPort struct {
	port *uint32
}

// Helps you get a free port with ginkgo tests.
func NewTestPort(initial uint32) TestPort {
	return TestPort{
		port: &initial,
	}
}

func (t TestPort) NextPort() uint32 {
	return atomic.AddUint32(t.port, 1) + uint32(GinkgoParallelProcess()*MaxTests)
}
