package limits

import (
	"fmt"
	"os/exec"
)

func CheckLimits() error {
	cmd := exec.Command("kubectl", "get", "deployments", "-A", "-o", "jsonpath={range .items[*]}{.metadata.namespace}{\"\\t\"}{.metadata.name}{\"\\t\"}{.spec.template.spec.containers[*].resources}{\"\\n\"}{end}")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	fmt.Println("Deployments and their resource limits:")
	fmt.Println(string(output))
	return nil
}
