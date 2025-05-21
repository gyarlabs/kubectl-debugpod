package cluster

import (
    "fmt"
    "os/exec"
)

func RunClusterCheck(args []string) error {
    cmdArgs := append([]string{"analyze", "--output", "text"}, args...)
    cmd := exec.Command("k8sgpt", cmdArgs...)
    cmd.Stdout = nil
    cmd.Stderr = nil

    fmt.Println("Running k8sgpt:", cmd.String())
    return cmd.Run()
}
