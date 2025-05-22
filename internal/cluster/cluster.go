package cluster

import (
	"fmt"
	"os"
	"os/exec"
)

func RunClusterCheck(args []string) {
	baseArgs := []string{"analyze", "--output", "text"}

	// Append additional user-supplied args
	baseArgs = append(baseArgs, args...)

	fmt.Println("Running k8sgpt:", "k8sgpt", baseArgs)

	cmd := exec.Command("k8sgpt", baseArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running k8sgpt: %v\n", err)
	}
}
