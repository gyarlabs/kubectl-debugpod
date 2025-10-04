package rbac

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const serviceAccountManifest = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: debugpod-sa
  namespace: %s
`

const clusterRoleManifest = `
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

const clusterRoleBindingManifest = `
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
`

func CreateRBAC(namespace string) error {
	fmt.Println("Creating RBAC resources...")

	manifests := []string{
		fmt.Sprintf(serviceAccountManifest, namespace),
		clusterRoleManifest,
		fmt.Sprintf(clusterRoleBindingManifest, namespace),
	}

	for _, manifest := range manifests {
		if err := applyManifest(manifest); err != nil {
			DeleteRBAC(namespace)
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
		cmd := exec.Command("kubectl", append([]string{"delete", "--ignore-not-found"}, args...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to delete RBAC resource (%s): %v\n", strings.Join(args, " "), err)
		}
	}
}

func applyManifest(manifest string) error {
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(manifest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
