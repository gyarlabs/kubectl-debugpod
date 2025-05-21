package debugpod

import (
    "fmt"
    "os"
    "os/exec"
)

type DebugOptions struct {
    Namespace string
    Name      string
    Image     string
    Node      string
    Stay      bool
    Command   string
}

func RunDebugPod(opt DebugOptions) error {
    manifest := fmt.Sprintf(`
apiVersion: v1
kind: Pod
metadata:
  name: %s
  namespace: %s
spec:
  containers:
  - name: debug
    image: %s
    command: ["/bin/sh", "-c", "%s"]
    stdin: true
    tty: true
  restartPolicy: Never`, opt.Name, opt.Namespace, opt.Image, opt.Command)

    if opt.Node != "" {
        manifest += fmt.Sprintf("\n  nodeName: %q", opt.Node)
    }

    file, err := os.CreateTemp("", "debugpod-*.yaml")
    if err != nil {
        return err
    }
    defer os.Remove(file.Name())

    _, err = file.WriteString(manifest)
    if err != nil {
        return err
    }
    file.Close()

    if err := exec.Command("kubectl", "apply", "-f", file.Name()).Run(); err != nil {
        return err
    }

    waitCmd := exec.Command("kubectl", "wait", "--for=condition=Ready", "pod/"+opt.Name, "-n", opt.Namespace, "--timeout=30s")
    waitCmd.Stdout = os.Stdout
    waitCmd.Stderr = os.Stderr
    if err := waitCmd.Run(); err != nil {
        return err
    }

    attachCmd := exec.Command("kubectl", "-n", opt.Namespace, "attach", "-it", opt.Name)
    attachCmd.Stdin = os.Stdin
    attachCmd.Stdout = os.Stdout
    attachCmd.Stderr = os.Stderr
    if err := attachCmd.Run(); err != nil {
        return err
    }

    if !opt.Stay {
        _ = exec.Command("kubectl", "delete", "pod", opt.Name, "-n", opt.Namespace).Run()
    } else {
        fmt.Println("Pod is left running.")
    }

    return nil
}
