package usageutils

import (
	check "github.com/solo-io/go-checkpoint"
	"math/rand"
	"time"
)

// A simple interface for interacting with the checkpoint server, for reporting and version checking
type UsageClient interface {
	Start(name, version string)
}

var _ UsageClient = NewUsageClient()

func NewUsageClient() *usageClient {
	return &usageClient{}
}

type usageClient struct {
}

func (c *usageClient) Start(name, version string) {
	now := time.Now()
	check.CallReport(name, version, now)
	check.CallCheck(name, version, now)

	jitter := time.Duration(rand.Intn(120)) * time.Minute
	ticker := time.NewTicker(23 * time.Hour + jitter)
	go func() {
		for t := range ticker.C {
			check.CallCheck(name, version, t)
		}
	}()
}
