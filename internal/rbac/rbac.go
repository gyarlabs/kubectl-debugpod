package rbac

import (
	"bytes"
	"fmt"
	"os/exec"
)

const rbacManifest = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: debugpod-sa
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: debugpod-role
rules:
- apiGroups: [""]
  resources: ["pods", "configmaps", "endpoints", "nodes", "persistentvolumeclaims"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets", "statefulsets"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["batch"]
  resources: ["cronjobs"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["networking.k8s.io"]
  resources: ["ingresses"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["admissionregistration.k8s.io"]
  resources: ["validatingwebhookconfigurations", "mutatingwebhookconfigurations"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: debugpod-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: debugpod-role
subjects:
- kind: ServiceAccount
  name: debugpod-sa
  namespace: default
`

func CreateRBAC() error {
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = bytes.NewBufferString(rbacManifest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create RBAC resources: %v\nOutput: %s", err, output)
	}
	return nil
}

func DeleteRBAC() error {
	cmd := exec.Command("kubectl", "delete", "-f", "-")
	cmd.Stdin = bytes.NewBufferString(rbacManifest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete RBAC resources: %v\nOutput: %s", err, output)
	}
	return nil
}
