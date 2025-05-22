package debugpod

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type DebugOptions struct {
	Namespace string
	NodeName  string
	Image     string
	Stay      bool
	UseBash   bool
}

var kubectlPath = "kubectl" // You can override this if needed

func RunDebugPod(opts DebugOptions) {
	podName := "debugpod"
	args := []string{"run", podName, "-n", opts.Namespace, "--image", opts.Image, "--restart=Never"}

	if opts.NodeName != "" {
		args = append(args, "--overrides", fmt.Sprintf(`{"spec":{"nodeName":"%s"}}`, opts.NodeName))
	}

	if !opts.Stay {
		args = append(args, "--rm", "-it")
	}

	shell := "/bin/sh"
	if opts.UseBash {
		shell = "/bin/bash"
	}

	// Attach shell directly unless --stay, then use sleep to hold the pod
	if opts.Stay {
		args = append(args, "--", "sleep", "3600")
	} else {
		args = append(args, "--", shell)
	}

	fmt.Println("Running debug pod:", kubectlPath, strings.Join(args, " "))

	cmd := exec.Command(kubectlPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running debug pod: %v\n", err)
		return
	}

	// Wait and attach only for --stay
	if opts.Stay {
		fmt.Println("Waiting for pod to be ready...")

		waitCmd := exec.Command(kubectlPath, "wait", "--for=condition=Ready", "pod/"+podName, "-n", opts.Namespace, "--timeout=30s")
		waitCmd.Stdout = os.Stdout
		waitCmd.Stderr = os.Stderr
		if err := waitCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "pod not ready: %v\n", err)
			return
		}

		fmt.Println("Attaching to pod...")
		execCmd := exec.Command(kubectlPath, "exec", "-n", opts.Namespace, "-it", podName, "--", shell)
		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		if err := execCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "error attaching to pod: %v\n", err)
		}
	}
}
