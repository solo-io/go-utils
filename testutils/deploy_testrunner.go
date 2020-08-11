package testutils

import (
	"context"

	"k8s.io/client-go/rest"
)

// Deprecated: no one calls this
func DeployTestRunner(ctx context.Context, cfg *rest.Config, namespace string) error {
	return DeployFromYaml(ctx, cfg, namespace, TestRunnerYaml)
}

const TestRunnerYaml = `
apiVersion: v1
kind: Pod
metadata:
  labels:
    gloo: testrunner
  name: testrunner
spec:
  containers:
  - image: soloio/testrunner:testing-8671e8b9
    imagePullPolicy: IfNotPresent
    command:
      - sleep
      - "36000"
    name: testrunner
  restartPolicy: Always`

const NginxYaml = `
apiVersion: apps/v1 
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  selector:
    matchLabels:
      app: nginx
  replicas: 2 
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  ports:
  - port: 80
    name: http
  selector:
    app: nginx
`
