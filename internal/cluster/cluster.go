package cluster

import (
	"fmt"
	"strings"

	"github.com/gyarlabs/kubectl-debugpod/internal/debugpod"
	"github.com/gyarlabs/kubectl-debugpod/internal/rbac"
)

func RunClusterCheck(userArgs []string) {
	const (
		namespace = "default"
		name      = "debugpod"
		image     = "arsaphone/debugpod:v2"
	)

	fmt.Println("Creating debug pod to run k8sgpt...")

	// Prepare k8sgpt command
	args := []string{"analyze", "--output", "text"}
	args = append(args, userArgs...)
	cmd := "k8sgpt " + strings.Join(args, " ")

	// Create necessary RBAC resources
	if err := rbac.CreateRBACResources(namespace, name); err != nil {
		fmt.Printf("Failed to create RBAC resources: %v\n", err)
		return
	}
	defer rbac.DeleteRBACResources(namespace, name)

	// Run debug pod with the created service account
	debugpod.RunDebugPod(debugpod.DebugOptions{
		Namespace:      namespace,
		Image:          image,
		Stay:           false,
		ServiceAccount: fmt.Sprintf("%s-sa", name),
		Command:        []string{"/bin/sh", "-c", cmd},
	})
}
