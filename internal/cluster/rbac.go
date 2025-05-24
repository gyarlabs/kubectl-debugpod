package cluster

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	serviceAccount     = "debugpod-sa"
	clusterRole        = "debugpod-role"
	clusterRoleBinding = "debugpod-rolebinding"
	namespace          = "default"
)

func CreateRBAC() error {
	fmt.Println("Creating RBAC resources for k8sgpt...")

	// Create sa
	if err := kubectlApply([]string{"create", "serviceaccount", serviceAccount, "-n", namespace}); err != nil {
		return err
	}

	// Create cr
	roleYAML := fmt.Sprintf(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: %s
rules:
- apiGroups: ["", "apps", "batch", "networking.k8s.io", "admissionregistration.k8s.io"]
  resources: ["pods", "nodes", "configmaps", "services", "endpoints", "deployments", "replicasets", "cronjobs", "statefulsets", "ingresses", "validatingwebhookconfigurations", "mutatingwebhookconfigurations", "persistentvolumeclaims"]
  verbs: ["get", "list"]
`, clusterRole)

	if err := kubectlApplyYAML(roleYAML); err != nil {
		return err
	}

	// Create crb
	bindingYAML := fmt.Sprintf(`apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: %s
subjects:
- kind: ServiceAccount
  name: %s
  namespace: %s
roleRef:
  kind: ClusterRole
  name: %s
  apiGroup: rbac.authorization.k8s.io
`, clusterRoleBinding, serviceAccount, namespace, clusterRole)

	return kubectlApplyYAML(bindingYAML)
}

func DeleteRBAC() {
	fmt.Println("Cleaning up RBAC resources...")

	_ = kubectlApply([]string{"delete", "serviceaccount", serviceAccount, "-n", namespace})
	_ = kubectlApply([]string{"delete", "clusterrole", clusterRole})
	_ = kubectlApply([]string{"delete", "clusterrolebinding", clusterRoleBinding})
}

func kubectlApply(args []string) error {
	cmd := exec.Command("kubectl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error: %s\n%s\n", err, string(output))
	}
	return err
}

func kubectlApplyYAML(yaml string) error {
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(yaml)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error: %s\n%s\n", err, string(output))
	}
	return err
}
