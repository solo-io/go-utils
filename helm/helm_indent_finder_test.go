package helm_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/helm"
)

var _ = Describe("helm indent finder", func() {
	var opts helm.HelmDetectOptions
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
			badWindows := helm.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(0))
		})
		It("should not detect empty lines", func() {
			data := `

`
			badWindows := helm.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(0))
		})
	})

	Describe("dealing with white spaces in helm", func() {
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
			badWindows := helm.FindHelmChartWhiteSpaces(data, opts)
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
			badWindows := helm.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(0))
		})
		It("should detect empty lines with spaces", func() {
			opts := helm.HelmDetectOptions{DetectWhiteSpacesInEmptyLines: true}
			data := `
 
`
			badWindows := helm.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(1))
		})

		It("should detect a single white space in the yaml", func() {
			opts := helm.HelmDetectOptions{DetectWhiteSpacesInEmptyLines: true}
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
			badWindows := helm.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(1))
		})
		It("should detect a single white spaces in the yaml", func() {
			opts := helm.HelmDetectOptions{DetectWhiteSpacesInEmptyLines: true}
			data := `
      
      
---

---
# Source: gloo-ee/charts/gloo/templates/2-gloo-service.yaml
 

apiVersion: v1
kind: Service
metadata:
  labels:
`
			badWindows := helm.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(2))
		})
	})

	Describe("dealing with arrays in helm", func() {
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
			badWindows := helm.FindHelmChartWhiteSpaces(data, opts)
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
			badWindows := helm.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(1))
		})
		It("should find that the array is off by one", func() {
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
			badWindows := helm.FindHelmChartWhiteSpaces(data, opts)
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
			badWindows := helm.FindHelmChartWhiteSpaces(data, opts)
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
			badWindows := helm.FindHelmChartWhiteSpaces(data, opts)
			Expect(len(badWindows)).To(Equal(0))
		}, Entry("should accept |-", "|-"),
			Entry("should accept |", "|"),
			Entry("should accept |+", "|+"),
			Entry("should accept >", ">"),
			Entry("should accept >-", ">+"),
			Entry("should accept >+", ">-"))
	})
	It("should run the chat", func() {
		data, err := os.ReadFile("chart.yaml")
		Expect(err).NotTo(HaveOccurred())
		badWindows := helm.FindHelmChartWhiteSpaces(string(data), opts)
		for _, w := range badWindows {
			for _, l := range w {
				fmt.Println(l)
			}
			fmt.Println("------------------------------------------------------------------------")
			fmt.Println("------------------------------------------------------------------------")
			fmt.Println("------------------------------------------------------------------------")
		}
		fmt.Println(fmt.Sprintf("%d number of bad windows", len(badWindows)))
	})
})
