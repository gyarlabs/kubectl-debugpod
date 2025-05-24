package debugpod

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type DebugOptions struct {
	Namespace      string
	NodeName       string
	Image          string
	Stay           bool
	UseBash        bool
	Command        []string
	ServiceAccount string // <-- add this
}

var kubectlPath = "kubectl"

func RunDebugPod(opts DebugOptions) {
	podName := "debugpod"
	args := []string{"run", podName, "-n", opts.Namespace, "--image", opts.Image, "--restart=Never"}

	if opts.NodeName != "" {
		args = append(args, "--overrides", fmt.Sprintf(`{"spec":{"nodeName":"%s"}}`, opts.NodeName))
	}

	if opts.ServiceAccount != "" {
		args = append(args, "--serviceaccount", opts.ServiceAccount)
	}

	if !opts.Stay {
		args = append(args, "--rm", "-it")
	}

	// If a custom command is passed, use it
	if len(opts.Command) > 0 {
		args = append(args, "--")
		args = append(args, opts.Command...)
	} else {
		// Default to shell or sleep based on --stay
		shell := "/bin/sh"
		if opts.UseBash {
			shell = "/bin/bash"
		}

		if opts.Stay {
			args = append(args, "--", "sleep", "3600")
		} else {
			args = append(args, "--", shell)
		}
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

	// If staying, wait and attach after creation
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
		shell := "/bin/sh"
		if opts.UseBash {
			shell = "/bin/bash"
		}
		execCmd := exec.Command(kubectlPath, "exec", "-n", opts.Namespace, "-it", podName, "--", shell)
		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		if err := execCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "error attaching to pod: %v\n", err)
		}
	}
}
