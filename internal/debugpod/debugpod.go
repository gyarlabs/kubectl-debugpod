package debugpod

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	rbac "github.com/gyarlabs/kubectl-debugpod/internal/rbac"
)

type DebugOptions struct {
	Namespace      string
	Image          string
	Command        string
	ServiceAccount string
	NodeName       string
	Stay           bool
	UseBash        bool
	ClusterCheck   bool
}

const (
	debugPodName       = "debugpod"
	defaultWaitTimeout = 2 * time.Minute
)

const podTemplate = `
apiVersion: v1
kind: Pod
metadata:
  name: ` + debugPodName + `
  namespace: {{.Namespace}}
spec:
  restartPolicy: Never
{{- if .ServiceAccount }}
  serviceAccountName: {{.ServiceAccount}}
{{- end }}
{{- if .NodeName }}
  nodeName: {{.NodeName}}
{{- end }}
  containers:
  - name: debug
    image: {{.Image}}
    command: ["/bin/sh", "-c"]
    args: ["{{.PodCommand}}"]
`

func RunDebugPod(opts DebugOptions) {
	fmt.Println("Creating debug pod...")

	podCommand := strings.TrimSpace(opts.Command)
	switch {
	case podCommand != "":
		// user override
	case opts.ClusterCheck:
		podCommand = "k8sgpt analyze --output text"
	default:
		podCommand = "sleep 3600"
	}

	escapedCommand := strings.ReplaceAll(podCommand, "\"", "\\\"")

	tmpl, err := template.New("debugpod").Parse(podTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing pod template: %v\n", err)
		return
	}

	if opts.ClusterCheck {
		if err := rbac.CreateRBAC(opts.Namespace); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating RBAC resources: %v\n", err)
			return
		}
		defer rbac.DeleteRBAC(opts.Namespace)
	}

	data := struct {
		Namespace      string
		Image          string
		ServiceAccount string
		PodCommand     string
		NodeName       string
	}{
		Namespace:      opts.Namespace,
		Image:          opts.Image,
		ServiceAccount: opts.ServiceAccount,
		PodCommand:     escapedCommand,
		NodeName:       opts.NodeName,
	}

	var podYAML bytes.Buffer
	if err := tmpl.Execute(&podYAML, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing pod template: %v\n", err)
		return
	}

	applyCmd := exec.Command("kubectl", "apply", "-f", "-")
	applyCmd.Stdin = &podYAML
	applyCmd.Stdout = os.Stdout
	applyCmd.Stderr = os.Stderr
	if err := applyCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error applying pod manifest: %v\n", err)
		return
	}

	if waitErr := waitForPodReady(opts.Namespace); waitErr != nil {
		phase, phaseErr := getPodPhase(opts.Namespace)
		if phaseErr == nil && phase == "Succeeded" {
			fmt.Fprintln(os.Stderr, "Debug pod completed before reporting Ready; continuing with log collection.")
		} else {
			fmt.Fprintf(os.Stderr, "Debug pod did not become ready: %v\n", waitErr)
			if phaseErr != nil {
				fmt.Fprintf(os.Stderr, "Failed to determine pod phase: %v\n", phaseErr)
			} else {
				fmt.Fprintf(os.Stderr, "Current pod phase: %s\n", phase)
			}
			_ = showPodDiagnostics(opts.Namespace)
			if !opts.Stay {
				_ = deleteDebugPod(opts.Namespace)
			}
			return
		}
	}

	if opts.UseBash || opts.Stay {
		shell := "/bin/sh"
		if opts.UseBash {
			shell = "/bin/bash"
		}

		execArgs := []string{"exec", "-n", opts.Namespace, "-i"}
		if isTerminal() {
			execArgs = append(execArgs, "-t")
		}
		execArgs = append(execArgs, debugPodName, "--", shell)

		execCmd := exec.Command("kubectl", execArgs...)
		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr

		if err := execCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running shell: %v\n", err)
		}

		if opts.Stay {
			fmt.Printf("Leaving pod %s/%s running because --stay was set.\n", opts.Namespace, debugPodName)
			return
		}

		_ = deleteDebugPod(opts.Namespace)
		return
	}

	if opts.ClusterCheck {
		streamLogs(opts.Namespace, true)
		_ = deleteDebugPod(opts.Namespace)
		return
	}

	streamLogs(opts.Namespace, true)
	_ = deleteDebugPod(opts.Namespace)
}

func waitForPodReady(namespace string) error {
	waitCmd := exec.Command(
		"kubectl",
		"wait",
		"--for=condition=Ready",
		fmt.Sprintf("pod/%s", debugPodName),
		"-n",
		namespace,
		fmt.Sprintf("--timeout=%s", defaultWaitTimeout.String()),
	)
	waitCmd.Stdout = os.Stdout
	waitCmd.Stderr = os.Stderr
	return waitCmd.Run()
}

func getPodPhase(namespace string) (string, error) {
	cmd := exec.Command(
		"kubectl",
		"get",
		"pod",
		debugPodName,
		"-n",
		namespace,
		"-o",
		"jsonpath={.status.phase}",
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

func showPodDiagnostics(namespace string) error {
	describeCmd := exec.Command("kubectl", "describe", "pod", debugPodName, "-n", namespace)
	describeCmd.Stdout = os.Stdout
	describeCmd.Stderr = os.Stderr
	if err := describeCmd.Run(); err != nil {
		return err
	}

	logsCmd := exec.Command("kubectl", "logs", "-n", namespace, fmt.Sprintf("pod/%s", debugPodName), "--tail=20")
	logsCmd.Stdout = os.Stdout
	logsCmd.Stderr = os.Stderr
	return logsCmd.Run()
}

func streamLogs(namespace string, follow bool) {
	args := []string{"logs", "-n", namespace, fmt.Sprintf("pod/%s", debugPodName)}
	if follow {
		args = append(args, "-f")
	}

	logsCmd := exec.Command("kubectl", args...)
	logsCmd.Stdout = os.Stdout
	logsCmd.Stderr = os.Stderr
	if err := logsCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error streaming pod logs: %v\n", err)
	}
}

func deleteDebugPod(namespace string) error {
	deleteCmd := exec.Command("kubectl", "delete", "pod", debugPodName, "-n", namespace, "--ignore-not-found")
	deleteCmd.Stdout = os.Stdout
	deleteCmd.Stderr = os.Stderr
	return deleteCmd.Run()
}

func isTerminal() bool {
	if fileInfo, err := os.Stdout.Stat(); err == nil && (fileInfo.Mode()&os.ModeCharDevice) != 0 {
		return true
	}
	return false
}
