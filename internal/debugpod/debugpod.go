package debugpod

import (
	"fmt"
	"os/exec"
)

type DebugOptions struct {
	Namespace string
	NodeName  string
	Image     string
	Stay      bool
	UseBash   bool
}

func RunDebugPod(opts DebugOptions) error {
	args := []string{"run", "debugpod", "-n", opts.Namespace, "--image", opts.Image, "--rm", "-i", "--tty"}
	if opts.NodeName != "" {
		args = append(args, "--overrides", fmt.Sprintf(`{"spec": {"nodeName": "%s"}}`, opts.NodeName))
	}
	if opts.Stay {
		args = append(args[:len(args)-2], "--restart=Never")
	}
	shell := "/bin/sh"
	if opts.UseBash {
		shell = "/bin/bash"
	}
	args = append(args, "--", shell)

	cmd := exec.Command("kubectl", args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	fmt.Println("Running debug pod:", cmd.String())
	return cmd.Run()
}
