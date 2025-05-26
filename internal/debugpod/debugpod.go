package debugpod

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"text/template"
)

type DebugOptions struct {
	Namespace      string
	Image          string
	Command        []string
	ServiceAccount string
	NodeName       string
	Stay           bool // --bash flag
	UseBash        bool
	ClusterCheck   bool
}

const podTemplate = `
apiVersion: v1
kind: Pod
metadata:
  name: debugpod
  namespace: {{.Namespace}}
spec:
  restartPolicy: Never
  serviceAccountName: {{.ServiceAccount}}
  containers:
  - name: debug
    image: {{.Image}}
    command: ["/bin/sh", "-c"]
    args: ["{{.PodCommand}}"]
`

func RunDebugPod(opts DebugOptions) {
	fmt.Println("Creating debug pod...")

	// Default command
	podCommand := "sleep 3600"
	if opts.ClusterCheck {
		podCommand = "k8sgpt analyze --output text"
	}

	tmpl, err := template.New("debugpod").Parse(podTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing pod template: %v\n", err)
		return
	}

	data := struct {
		Namespace      string
		Image          string
		ServiceAccount string
		PodCommand     string
	}{
		Namespace:      opts.Namespace,
		Image:          opts.Image,
		ServiceAccount: opts.ServiceAccount,
		PodCommand:     podCommand,
	}

	var podYAML bytes.Buffer
	err = tmpl.Execute(&podYAML, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing pod template: %v\n", err)
		return
	}

	// Apply the pod manifest
	applyCmd := exec.Command("kubectl", "apply", "-f", "-")
	applyCmd.Stdin = &podYAML
	applyCmd.Stdout = os.Stdout
	applyCmd.Stderr = os.Stderr
	if err := applyCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error applying pod manifest: %v\n", err)
		return
	}

	// Wait for the pod to be ready
	waitCmd := exec.Command("kubectl", "wait", "--for=condition=Ready", "pod/debugpod", "-n", opts.Namespace, "--timeout=30s")
	waitCmd.Stdout = os.Stdout
	waitCmd.Stderr = os.Stderr
	if err := waitCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error waiting for pod readiness: %v\n", err)
		_ = deleteDebugPod(opts.Namespace)
		return
	}

	// Handle interactive bash mode
	if opts.Stay {
		shell := "/bin/sh"
		if opts.UseBash {
			shell = "/bin/bash"
		}

		execCmd := exec.Command("kubectl", "exec", "-n", opts.Namespace, "-it", "debugpod", "--", shell)
		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr

		if err := execCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running shell: %v\n", err)
		}

		_ = deleteDebugPod(opts.Namespace)
		return
	}

	// Cluster-check: stream logs and delete after
	if opts.ClusterCheck {
		logsCmd := exec.Command("kubectl", "logs", "-f", "debugpod", "-n", opts.Namespace)
		logsCmd.Stdout = os.Stdout
		logsCmd.Stderr = os.Stderr
		if err := logsCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error streaming pod logs: %v\n", err)
		}
		_ = deleteDebugPod(opts.Namespace)
		return
	}

	// Default case: just stream logs
	logsCmd := exec.Command("kubectl", "logs", "-f", "debugpod", "-n", opts.Namespace)
	logsCmd.Stdout = os.Stdout
	logsCmd.Stderr = os.Stderr
	if err := logsCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error streaming pod logs: %v\n", err)
	}
	_ = deleteDebugPod(opts.Namespace)
}

func deleteDebugPod(namespace string) error {
	deleteCmd := exec.Command("kubectl", "delete", "pod", "debugpod", "-n", namespace)
	deleteCmd.Stdout = os.Stdout
	deleteCmd.Stderr = os.Stderr
	return deleteCmd.Run()
}
