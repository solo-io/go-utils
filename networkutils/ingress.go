package networkutils

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rotisserie/eris"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	LocalClusterName = "minikube"
)

type ServiceRef struct {
	Name      string
	Namespace string
}

func GetIngressHostAndPort(ctx context.Context, restCfg *rest.Config, ref *ServiceRef, proxyPort string) (string, uint32, error) {
	url, err := GetIngressHost(ctx, restCfg, ref, proxyPort)
	if err != nil {
		return "", 0, err
	}
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	parts := strings.Split(url, ":")
	if len(parts) == 2 {
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", 0, eris.Wrapf(err, "could not convert port to int")
		}
		return parts[0], uint32(port), nil
	}
	return "", 0, eris.Errorf("Unexpected url %s", url)
}

func GetIngressHost(ctx context.Context, restCfg *rest.Config, ref *ServiceRef, proxyPort string) (string, error) {
	kube, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return "", err
	}
	namespace, name := ref.Namespace, ref.Name
	svc, err := kube.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", eris.Wrapf(err, "could not detect '%v' service in %v namespace", name, namespace)
	}
	var svcPort *v1.ServicePort
	switch len(svc.Spec.Ports) {
	case 0:
		return "", eris.Errorf("service %v is missing ports", name)
	case 1:
		svcPort = &svc.Spec.Ports[0]
	default:
		for _, p := range svc.Spec.Ports {
			if p.Name == proxyPort {
				svcPort = &p
				break
			}
		}
		if svcPort == nil {
			return "", eris.Errorf("named port %v not found on service %v", proxyPort, name)
		}
	}

	var host, port string
	if len(svc.Status.LoadBalancer.Ingress) == 0 {
		// assume nodeport on kubernetes
		// TODO: support more types of NodePort services
		host, err = getNodeIp(ctx, svc, kube)
		if err != nil {
			return "", eris.Wrapf(err, "")
		}
		port = fmt.Sprintf("%v", svcPort.NodePort)
	} else {
		host = svc.Status.LoadBalancer.Ingress[0].Hostname
		if host == "" {
			host = svc.Status.LoadBalancer.Ingress[0].IP
		}
		port = fmt.Sprintf("%v", svcPort.Port)
	}
	return host + ":" + port, nil
}

func getNodeIp(ctx context.Context, svc *v1.Service, kube kubernetes.Interface) (string, error) {
	// pick a node where one of our pods is running
	pods, err := kube.CoreV1().Pods(svc.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(svc.Spec.Selector).String(),
	})
	if err != nil {
		return "", err
	}
	var nodeName string
	for _, pod := range pods.Items {
		if pod.Spec.NodeName != "" {
			nodeName = pod.Spec.NodeName
			break
		}
	}
	if nodeName == "" {
		return "", eris.Errorf("no node found for %v's pods. ensure at least one pod has been deployed "+
			"for the %v service", svc.Name, svc.Name)
	}
	// special case for minikube
	// we run `minikube ip` which avoids an issue where
	// we get a NAT network IP when the minikube provider is virtualbox
	if nodeName == "minikube" {
		return minikubeIp(LocalClusterName)
	}

	node, err := kube.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	for _, addr := range node.Status.Addresses {
		return addr.Address, nil
	}

	return "", eris.Errorf("no active addresses found for node %v", node.Name)
}

func minikubeIp(clusterName string) (string, error) {
	minikubeCmd := exec.Command("minikube", "ip", "-p", clusterName)

	hostname := &bytes.Buffer{}

	minikubeCmd.Stdout = hostname
	minikubeCmd.Stderr = os.Stderr
	err := minikubeCmd.Run()

	return strings.TrimSuffix(hostname.String(), "\n"), err
}
