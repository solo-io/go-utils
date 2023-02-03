package grpcutils_test

import (
    "testing"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

func TestGrpc(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Grpc Suite")
}
