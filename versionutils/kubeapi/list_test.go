package kubeapi_test

import (
	"math/rand"
	"sort"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/versionutils/kubeapi"
)

var _ = Describe("VersionList", func() {
	Describe("sort", func() {
		It("works", func() {
			orderedVersions := []string{"v1alpha1", "v1beta1", "v1beta2", "v1", "v2beta1", "v2beta2", "v4", "v5alpha2", "v5beta1"}

			subject := make(kubeapi.VersionList, 0, len(orderedVersions))
			for _, v := range orderedVersions {
				parsedVersion, err := kubeapi.ParseVersion(v)
				Expect(err).NotTo(HaveOccurred())
				subject = append(subject, parsedVersion)
			}
			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(subject.Len(), subject.Swap)

			sort.Slice(subject, subject.Less)

			for i, apiVersion := range subject {
				Expect(apiVersion.String()).To(Equal(orderedVersions[i]))
			}
		})
	})
})
