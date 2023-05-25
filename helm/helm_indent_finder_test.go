package helm_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/helm"
)

var _ = Describe("helm indent finder", func() {
	It("should be able to detect 4 spaces, when it should be 2", func() {
		data := `
apiVersion: v1
kind: Service
metadata:
   labels:
    app: gloo
    gloo: rate-limit
  name: rate-limit
  namespace: default
spec:
`
		badWindows := helm.FindHelmChartWhiteSpaces(data)
		Expect(len(badWindows)).To(Equal(2))
	})
	It("should not detect fine yaml", func() {
		data := `
apiVersion: v1
kind: Service
metadata:
  labels:
    app: gloo
    gloo: rate-limit
spec:
`
		badWindows := helm.FindHelmChartWhiteSpaces(data)
		Expect(len(badWindows)).To(Equal(0))
	})
	It("should not detect empty lines", func() {
		data := `

`
		badWindows := helm.FindHelmChartWhiteSpaces(data)
		Expect(len(badWindows)).To(Equal(0))
	})
	It("should detect empty lines with spaces", func() {
		data := `
 
`
		badWindows := helm.FindHelmChartWhiteSpaces(data)
		Expect(len(badWindows)).To(Equal(1))
	})
	It("should detect white space at the end of an line", func() {
		data := `
apiVersion: v1
kind: Service
metadata:
  labels:       
    app: gloo
    gloo: rate-limit
spec:
`
		badWindows := helm.FindHelmChartWhiteSpaces(data)
		Expect(len(badWindows)).To(Equal(0))
	})
	It("should detect a single white space in the yaml", func() {
		// the line after the comment has 1 space in it
		data := `
roleRef:
  kind: ClusterRole
  name: gloo-graphqlapi-mutator-default
  apiGroup: rbac.authorization.k8s.io
---
# Source: gloo-ee/charts/gloo/templates/1-gloo-deployment.yaml
 


apiVersion: apps/v1
kind: Deployment
metadata:
`
		badWindows := helm.FindHelmChartWhiteSpaces(data)
		Expect(len(badWindows)).To(Equal(1))
	})
	It("should detect a single white spaces in the yaml", func() {
		data := `
      
      
---

---
# Source: gloo-ee/charts/gloo/templates/2-gloo-service.yaml
 

apiVersion: v1
kind: Service
metadata:
  labels:
`
		badWindows := helm.FindHelmChartWhiteSpaces(data)
		Expect(len(badWindows)).To(Equal(2))
	})
	It("should be ok with array in array", func() {
		data := `
  resources:
  - federatedauthconfigs
  - federatedauthconfigs/status
  verbs:
  - '*'
- apiGroups:
    - fed.ratelimit.solo.io
  resources:
    - federatedratelimitconfigs
    - federatedratelimitconfigs/status
  verbs:
    - '*'
`
		badWindows := helm.FindHelmChartWhiteSpaces(data)
		Expect(len(badWindows)).To(Equal(0))
	})
	It("should be find that the array is off by one", func() {
		// the line fed.ratelimit.solo.io is off by one
		data := `
  resources:
  - federatedauthconfigs
  - federatedauthconfigs/status
  verbs:
  - '*'
- apiGroups:
     - fed.ratelimit.solo.io
  resources:
    - federatedratelimitconfigs
    - federatedratelimitconfigs/status
  verbs:
    - '*'
`
		badWindows := helm.FindHelmChartWhiteSpaces(data)
		Expect(len(badWindows)).To(Equal(1))
	})
})
