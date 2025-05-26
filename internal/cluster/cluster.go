package cluster

import (
	"fmt"
	"strings"

	"github.com/gyarlabs/kubectl-debugpod/internal/debugpod"
	"github.com/gyarlabs/kubectl-debugpod/internal/rbac"
)

func RunClusterCheck(userArgs []string) {
	fmt.Println("Creating debug pod to run k8sgpt...")

	// Create RBAC resources
	if err := rbac.CreateRBAC(); err != nil {
		fmt.Printf("Error creating RBAC resources: %v\n", err)
		return
	}
	defer func() {
		if err := rbac.DeleteRBAC(); err != nil {
			fmt.Printf("Error deleting RBAC resources: %v\n", err)
		}
	}()

	// Construct k8sgpt command
	args := []string{"analyze", "--output", "text"}
	args = append(args, userArgs...)
	cmd := "k8sgpt " + strings.Join(args, " ")

	// Run debug pod
	debugpod.RunDebugPod(debugpod.DebugOptions{
		Namespace:      "default",
		Image:          "arsaphone/debugpod:v2",
		Command:        []string{"/bin/sh", "-c", cmd},
		ServiceAccount: "debugpod-sa",
	})
}
