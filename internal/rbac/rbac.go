package rbac

import (
	"fmt"
	"os/exec"
	"strings"
)

func CreateRBACResources(namespace, name string) error {
	if err := applyYAML(serviceAccount(namespace, name)); err != nil {
		return err
	}
	if err := applyYAML(clusterRole()); err != nil {
		return err
	}
	if err := applyYAML(clusterRoleBinding(namespace, name)); err != nil {
		return err
	}
	return nil
}

func DeleteRBACResources(namespace, name string) {
	exec.Command("kubectl", "delete", "serviceaccount", fmt.Sprintf("%s-sa", name), "-n", namespace).Run()
	exec.Command("kubectl", "delete", "clusterrolebinding", fmt.Sprintf("%s-binding", name)).Run()
	// Optional: delete ClusterRole if not shared
	// exec.Command("kubectl", "delete", "clusterrole", "cluster-reader").Run()
}

func applyYAML(yaml string) error {
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(yaml)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func serviceAccount(namespace, name string) string {
	return fmt.Sprintf(`
apiVersion: v1
kind: ServiceAccount
metadata:
  name: %s-sa
  namespace: %s
`, name, namespace)
}

func clusterRole() string {
	return `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-reader
rules:
- apiGroups: [""]
  resources: ["pods", "nodes", "endpoints", "services", "configmaps", "persistentvolumeclaims"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments", "statefulsets", "replicasets"]
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
`
}

func clusterRoleBinding(namespace, name string) string {
	return fmt.Sprintf(`
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: %s-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-reader
subjects:
- kind: ServiceAccount
  name: %s-sa
  namespace: %s
`, name, name, namespace)
}
