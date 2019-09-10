package test

import (
	"fmt"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/solo-io/go-utils/manifesttestutils"
)

// TODO joekelley don't ship this

var _ = Describe("Rbac Test", func() {
	Describe("Permissions", func() {
		Describe("AddExpectedPermission", func() {
			It("works", func() {
				subject := &ServiceAccountPermissions{}
				subject.AddExpectedPermission("tst", "tset", []string{"gloo.solo.io"}, []string{"proxies"}, []string{"get", "list"})
				subject.AddExpectedPermission("tst", "tset", []string{"gloo.solo.io"}, []string{"proxies"}, []string{"watch", "list"})
				x, _ := yaml.Marshal(subject)
				fmt.Println(string(x))
				Fail("nice")
			})
		})
	})
})
