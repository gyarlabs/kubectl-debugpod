package limits

import (
	"fmt"
	"os/exec"
	"strings"
)

func CheckLimits(namespace string) error {
	args := []string{"get", "deployments", "-o", "jsonpath={range .items[*]}{.metadata.namespace}{\"\\t\"}{.metadata.name}{\"\\t\"}{.spec.template.spec.containers[*].resources}{\"\\n\"}{end}"}
	if namespace != "" && namespace != "all" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "-A")
	}

	cmd := exec.Command("kubectl", args...)
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		fmt.Println("No deployments found.")
		return nil
	}

	fmt.Println("Deployments and their resource limits:")
	for _, line := range lines {
		fmt.Println(line)
	}
	return nil
}
