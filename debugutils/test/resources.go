package test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func GeneratePodList() *corev1.PodList {
	podList := &corev1.PodList{
		Items: []corev1.Pod{*GatewayPod(), *GatewayProxyPod(), *GlooPod(), *DiscoveryPod()},
	}
	return podList
}

var (
	GatewayPod = func() *corev1.Pod {
		return convertPod([]byte(GatewayPodYaml))
	}
	DiscoveryPod = func() *corev1.Pod {
		return convertPod([]byte(DiscoveryPodYaml))
	}
	GatewayProxyPod = func() *corev1.Pod {
		return convertPod([]byte(GatewayProxyPodYaml))
	}
	GlooPod = func() *corev1.Pod {
		return convertPod([]byte(GlooPodYaml))
	}
)

func convertPod(podYaml []byte) *corev1.Pod {
	var pod corev1.Pod
	Expect(yaml.Unmarshal(podYaml, &pod)).NotTo(HaveOccurred())
	return &pod
}

var DiscoveryPodYaml = `
apiVersion: v1
kind: Pod
metadata:
  annotations:
    prometheus.io/path: /metrics
    prometheus.io/port: "9091"
    prometheus.io/scrape: "true"
  labels:
    gloo: discovery
    pod-template-hash: 688d68bb6b
  name: discovery-688d68bb6b-4c4cp
  namespace: gloo-system
spec:
  containers:
  - env:
    - name: POD_NAMESPACE
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.namespace
    - name: START_STATS_SERVER
      value: "true"
    image: quay.io/solo-io/discovery:0.13.33
    imagePullPolicy: Always
    name: discovery
    resources: {}
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL
      procMount: Default
      readOnlyRootFilesystem: true
      runAsNonRoot: true
      runAsUser: 10101
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-9xmkt
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  nodeName: minikube
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  volumes:
  - name: default-token-9xmkt
    secret:
      defaultMode: 420
      secretName: default-token-9xmkt
  phase: Running
`

var GlooPodYaml = `
apiVersion: v1
kind: Pod
metadata:
  annotations:
    prometheus.io/path: /metrics
    prometheus.io/port: "9091"
    prometheus.io/scrape: "true"
  labels:
    gloo: gloo
    pod-template-hash: d97b6bd64
  name: gloo-d97b6bd64-6sgrm
  namespace: gloo-system
spec:
  containers:
  - env:
    - name: POD_NAMESPACE
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.namespace
    - name: START_STATS_SERVER
      value: "true"
    image: quay.io/solo-io/gloo:0.13.33
    imagePullPolicy: Always
    name: gloo
    ports:
    - containerPort: 9977
      name: grpc
      protocol: TCP
    resources:
      requests:
        cpu: 500m
        memory: 256Mi
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL
      procMount: Default
      readOnlyRootFilesystem: true
      runAsNonRoot: true
      runAsUser: 10101
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-9xmkt
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  nodeName: minikube
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  phase: Running

`

var GatewayPodYaml = `
apiVersion: v1
kind: Pod
metadata:
  annotations:
    prometheus.io/path: /metrics
    prometheus.io/port: "9091"
    prometheus.io/scrape: "true"
  labels:
    gloo: gateway
    pod-template-hash: 5d77dd7d4b
  name: gateway-5d77dd7d4b-c8gk2
  namespace: gloo-system
spec:
  containers:
  - env:
    - name: POD_NAMESPACE
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.namespace
    - name: START_STATS_SERVER
      value: "true"
    image: quay.io/solo-io/gateway:0.13.33
    imagePullPolicy: Always
    name: gateway
    resources: {}
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - ALL
      procMount: Default
      readOnlyRootFilesystem: true
      runAsNonRoot: true
      runAsUser: 10101
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-9xmkt
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  nodeName: minikube
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  volumes:
  - name: default-token-9xmkt
    secret:
      defaultMode: 420
      secretName: default-token-9xmkt
  phase: Running
`

var GatewayProxyPodYaml = `
apiVersion: v1
kind: Pod
metadata:
  annotations:
    prometheus.io/path: /metrics
    prometheus.io/port: "8081"
    prometheus.io/scrape: "true"
  labels:
    gloo: gateway-proxy
    pod-template-hash: 655765b499
  name: gateway-proxy-655765b499-wqxhm
  namespace: gloo-system
spec:
  containers:
  - args:
    - --disable-hot-restart
    env:
    - name: POD_NAMESPACE
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.namespace
    - name: POD_NAME
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.name
    image: quay.io/solo-io/gloo-envoy-wrapper:0.13.33
    imagePullPolicy: Always
    name: gateway-proxy
    ports:
    - containerPort: 8080
      name: http
      protocol: TCP
    - containerPort: 8443
      name: https
      protocol: TCP
    resources: {}
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        add:
        - NET_BIND_SERVICE
        drop:
        - ALL
      procMount: Default
      readOnlyRootFilesystem: true
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /etc/envoy
      name: envoy-config
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-9xmkt
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  nodeName: minikube
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  volumes:
  - configMap:
      defaultMode: 420
      name: gateway-proxy-envoy-config
    name: envoy-config
  - name: default-token-9xmkt
    secret:
      defaultMode: 420
      secretName: default-token-9xmkt
  phase: Running

`
