package cluster

import (
    "fmt"
    "strings"

    "github.com/gyarlabs/kubectl-debugpod/internal/debugpod"
)

func RunClusterCheck(userArgs []string) {
    fmt.Println("Creating debug pod to run k8sgpt...")

    namespace := "default"

    // Construct k8sgpt command
    args := []string{"analyze", "--output", "text"}
    args = append(args, userArgs...)
    cmd := "k8sgpt " + strings.Join(args, " ")

    // Run debug pod
    debugpod.RunDebugPod(debugpod.DebugOptions{
        Namespace:      namespace,
        Image:          "arsaphone/debugpod:v2",
        Command:        cmd,
        ServiceAccount: "debugpod-sa",
        ClusterCheck:   true,
    })
}