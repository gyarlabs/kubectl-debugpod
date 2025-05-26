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
    command: {{.Command}}
`

func RunDebugPod(opts DebugOptions) {
	fmt.Println("Creating debug pod to run k8sgpt...")

	tmpl, err := template.New("debugpod").Parse(podTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing pod template: %v\n", err)
		return
	}

	var podYAML bytes.Buffer
	err = tmpl.Execute(&podYAML, opts)
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
		// Proceed to fetch logs even if waiting fails
	}

	// Stream logs
	logsCmd := exec.Command("kubectl", "logs", "-f", "debugpod", "-n", opts.Namespace)
	logsCmd.Stdout = os.Stdout
	logsCmd.Stderr = os.Stderr
	if err := logsCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error streaming pod logs: %v\n", err)
	}

	// Delete the pod after execution
	deleteCmd := exec.Command("kubectl", "delete", "pod", "debugpod", "-n", opts.Namespace)
	deleteCmd.Stdout = os.Stdout
	deleteCmd.Stderr = os.Stderr
	if err := deleteCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting debug pod: %v\n", err)
	}
}
