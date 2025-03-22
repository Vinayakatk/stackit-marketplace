package kubernetes

import (
	"fmt"
	"os/exec"
	"strings"
)

// CreateKindCluster creates a local Kubernetes cluster using Kind
func CreateKindCluster(clusterName string) error {
	// Check if the cluster already exists
	checkCmd := exec.Command("kind", "get", "clusters")
	output, err := checkCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to list Kind clusters: %v\n%s", err, string(output))
	}

	// Convert output to a list and check if clusterName exists
	existingClusters := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, cluster := range existingClusters {
		if cluster == clusterName {
			fmt.Printf("✅ Kind cluster %s already exists, skipping creation\n", clusterName)
			return nil
		}
	}

	// Create the cluster if it doesn't exist
	cmd := exec.Command("kind", "create", "cluster", "--name", clusterName)
	createOutput, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create Kind cluster: %v\n%s", err, string(createOutput))
	}

	fmt.Printf("✅ Kind cluster %s created successfully\n", clusterName)
	return nil
}

func DeleteKindCluster(clusterName string) error {
	cmd := exec.Command("kind", "delete", "cluster", "--name", clusterName)
	return cmd.Run()
}
