package cluster

import (
	"fmt"
	"strings"

	"github.com/gyarlabs/kubectl-debugpod/internal/debugpod"
)

func RunClusterCheck(userArgs []string) {
	fmt.Println("Creating debug pod to run k8sgpt...")

	if err := CreateRBAC(); err != nil {
		fmt.Println("Failed to create RBAC:", err)
		return
	}
	defer DeleteRBAC()

	useAI := false
	for _, arg := range userArgs {
		if strings.Contains(arg, "--provider") || strings.Contains(arg, "--token") || strings.Contains(arg, "--no-ai") {
			useAI = true
			break
		}
	}

	args := []string{"analyze", "--output", "text"}
	if !useAI {
		args = append(args, "--no-ai")
	}
	args = append(args, userArgs...)

	cmd := "k8sgpt " + strings.Join(args, " ")

	debugpod.RunDebugPod(debugpod.DebugOptions{
		Namespace: "default",
		Image:     "arsaphone/debugpod:v2",
		Stay:      false,
		Command:   []string{"/bin/sh", "-c", cmd},
	})
}
