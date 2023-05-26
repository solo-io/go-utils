package helmutils_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/helmutils"
)

var _ = Describe("helmutils indent finder", func() {
	var opts helmutils.HelmDetectOptions
	Context("should pass on the following", func() {
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
			badWindows := helmutils.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(0))
		})
		It("should pass with a comment line", func() {
			data := `
apiVersion: v1
kind: Service
metadata:
  labels:
              # app: gloo
    gloo: rate-limit
spec:
`
			badWindows := helmutils.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(0))
		})
		It("should not detect empty lines", func() {
			data := `

`
			badWindows := helmutils.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(0))
		})
		It("should not detect empty lines with spaces", func() {
			data := `
        
`
			badWindows := helmutils.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(0))
		})
	})

	Describe("dealing with white spaces in helmutils", func() {
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
			badWindows := helmutils.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(2))
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
			badWindows := helmutils.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(0))
		})
		It("should detect empty lines with spaces", func() {
			opts := helmutils.HelmDetectOptions{DetectWhiteSpacesInEmptyLines: true}
			data := `
 
`
			badWindows := helmutils.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(1))
		})

		It("should detect a single white space in the yaml", func() {
			opts := helmutils.HelmDetectOptions{DetectWhiteSpacesInEmptyLines: true}
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
			badWindows := helmutils.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(1))
		})
		It("should detect a single white spaces in the yaml", func() {
			opts := helmutils.HelmDetectOptions{DetectWhiteSpacesInEmptyLines: true}
			data := `
      
      
---

---
# Source: gloo-ee/charts/gloo/templates/2-gloo-service.yaml
 

apiVersion: v1
kind: Service
metadata:
  labels:
`
			badWindows := helmutils.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(2))
		})
	})

	Describe("dealing with arrays in helmutils", func() {
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
			badWindows := helmutils.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(0))
		})
		It("should find that the array is off by one", func() {
			// the line fed.ratelimit.solo.io is off by one space
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
			badWindows := helmutils.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(1))
		})
		It("should find that the array is off by 4 spaces", func() {
			// note the - "" is off by 4 spaces, this is acceptable
			// this is whithin the array formatting specs. Same line, 2 spaces, or 4 spaces.
			data := `
resource:
  release: gloo-ee-test
    chart: prometheus-13.3.3
    heritage: Helm
  name: glooe-prometheus-server
rules:
  - apiGroups:
      - ""
    resources:
      - nodes
      - nodes/proxy
      - nodes/metrics
      - services
`
			badWindows := helmutils.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(0))
		})
		It("should accept arrays", func() {
			data := `
resources:
  something:
    there:
      should:
        be:
          envFrom:
            - configMapRef:
                name: gloo-ee-test-observability-config
            - secretRef:
                name: gloo-ee-test-observability-secrets
      securityContext:
        runAsNonRoot: true
        runAsUser: 10101
`
			badWindows := helmutils.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(0))
		})
	})

	Describe("special breaks", func() {
		DescribeTable("should accept special breaks", func(specialBreak string) {
			data := fmt.Sprintf(`
resources:
  envoy.json: %s
    {
      "annotations": {
        "list": [
          {
            "builtIn": 1,
            "datasource": "-- Grafana --",
            "enable": true,
            "hide": true,
            "iconColor": "rgba(0, 211, 255, 1)",
            "name": "Annotations & Alerts",
            "type": "dashboard"
          }
        ]
      },
`, specialBreak)
			badWindows := helmutils.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(0))
		}, Entry("should accept |-", "|-"),
			Entry("should accept |", "|"),
			Entry("should accept |+", "|+"),
			Entry("should accept >", ">"),
			Entry("should accept >-", ">+"),
			Entry("should accept >+", ">-"))
	})
})
