package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

const testYaml = `---
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  annotations:
    git_commit: '{{.git_commit}}'
    kubernetes.io/ingress.allow-http: "false"
    kubernetes.io/ingress.global-static-ip-name: '{{.ingress_static_ip_name}}'
    networking.gke.io/managed-certificates: '{{.ssl_cert_name}}'
  labels:
    app: '{{.app}}'
  name: '{{.app}}'
  namespace: '{{.namespace}}'
spec:
  backend:
    serviceName: '{{.app}}'
    servicePort: svc-https
---
apiVersion: v1
data:
  POSTGRESQL_PW: 'SECRET_POSTGRESQL_PW'
  POSTGRESQL_USER: 'SECRET_POSTGRESQL_USER'
kind: Secret
metadata:
  annotations:
    git_commit: '{{.git_commit}}'
  labels:
    app: '{{.app}}'
  name: postgresql-credentials
  namespace: '{{.namespace}}'
type: Opaque
---
apiVersion: v1
data:
  svc_account.json: 'SECRET_GOOGLE_APPLICATION_CREDENTIALS'
kind: Secret
metadata:
  annotations:
    git_commit: '{{.git_commit}}'
  labels:
    app: '{{.app}}'
  name: gcp-service-account-keys
  namespace: '{{.namespace}}'
type: Opaque`

var (
	testSecretsData = map[string]string{
		"svc_account.json": "'SECRET_GOOGLE_APPLICATION_CREDENTIALS'",
		"POSTGRESQL_PW":    "'SECRET_POSTGRESQL_PW'",
		"POSTGRESQL_USER":  "'SECRET_POSTGRESQL_USER'",
	}
)

func TestRedactedWriter_Write(t *testing.T) {
	buf := new(bytes.Buffer)
	redactedWriter := NewSecretsWriter(buf)
	if _, err := redactedWriter.Write([]byte(testYaml)); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for k, v := range testSecretsData {
		if strings.Contains(out, v) {
			t.Fatalf("found secret in output: %s", v)
		}
		if !strings.Contains(out, fmt.Sprintf("%s: <REDACTED>", k)) {
			t.Fatalf("did not find redacted secret in output: %s", k)
		}
	}
}
