package kubernetes

import (
	"fmt"
	"os/exec"
)

// CreateKindCluster creates a local Kubernetes cluster using Kind
func CreateKindCluster(clusterName string) error {
	cmd := exec.Command("kind", "create", "cluster", "--name", clusterName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create Kind cluster: %v\n%s", err, string(output))
	}
	fmt.Printf("✅ Kind cluster %s created successfully\n", clusterName)
	return nil
}
func DeleteKindCluster(clusterName string) error {
	cmd := exec.Command("kind", "delete", "cluster", "--name", clusterName)
	return cmd.Run()
}
