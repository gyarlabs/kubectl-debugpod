package debugpod

import (
	"fmt"
	"os"
	"os/exec"
)

func CreateRBAC(namespace string) error {
	fmt.Println("Creating RBAC resources...")

	serviceAccount := fmt.Sprintf(`
apiVersion: v1
kind: ServiceAccount
metadata:
  name: debugpod-sa
  namespace: %s
`, namespace)

	clusterRole := `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: debugpod-role
rules:
- apiGroups: [""]
  resources: ["pods", "nodes", "namespaces", "services", "configmaps", "secrets"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets", "statefulsets"]
  verbs: ["get", "list", "watch"]
`

	clusterRoleBinding := fmt.Sprintf(`
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: debugpod-binding
subjects:
- kind: ServiceAccount
  name: debugpod-sa
  namespace: %s
roleRef:
  kind: ClusterRole
  name: debugpod-role
  apiGroup: rbac.authorization.k8s.io
`, namespace)

	resources := []string{serviceAccount, clusterRole, clusterRoleBinding}

	for _, resource := range resources {
		cmd := exec.Command("kubectl", "apply", "-f", "-")
		cmd.Stdin = stringToReader(resource)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create RBAC resource: %w", err)
		}
	}
	return nil
}

func DeleteRBAC(namespace string) {
	fmt.Println("Deleting RBAC resources...")

	resources := [][]string{
		{"serviceaccount", "debugpod-sa", "-n", namespace},
		{"clusterrole", "debugpod-role"},
		{"clusterrolebinding", "debugpod-binding"},
	}

	for _, args := range resources {
		cmd := exec.Command("kubectl", append([]string{"delete"}, args...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run() // ignore errors on cleanup
	}
}

func stringToReader(s string) *os.File {
	tmpfile, err := os.CreateTemp("", "rbac-*.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp file: %v\n", err)
		os.Exit(1)
	}
	_, _ = tmpfile.Write([]byte(s))
	_ = tmpfile.Close()

	f, err := os.Open(tmpfile.Name())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to reopen temp file: %v\n", err)
		os.Exit(1)
	}
	return f
}
