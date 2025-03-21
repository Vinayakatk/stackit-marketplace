package helm

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// DeployHelmChart deploys a Helm chart onto a Kind cluster
func DeployHelmChart(clusterName, repoURL, chartName, application string) error {
	// Check if the repo already exists
	listCmd := exec.Command("helm", "repo", "list")
	output, err := listCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to list Helm repos: %v\n%s", err, string(output))
	}

	if strings.Contains(string(output), application) {
		log.Printf("Helm repo %s already exists, skipping add", application)
	} else {
		// Add repo if it doesn't exist
		addCmd := exec.Command("helm", "repo", "add", application, repoURL)
		if addOutput, addErr := addCmd.CombinedOutput(); addErr != nil {
			return fmt.Errorf("failed to add repo: %v\n%s", addErr, string(addOutput))
		}
		log.Printf("Helm repo %s added successfully", application)
	}

	// Update the Helm repo
	updateRepoCmd := exec.Command("helm", "repo", "update")
	if output, err := updateRepoCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to update repo: %v\n%s", err, string(output))
	}

	// Install the Helm chart
	installCmd := exec.Command("helm", "install", chartName, application+"/"+chartName, "--kube-context", "kind-"+clusterName)
	output, err = installCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to deploy Helm chart: %v\n%s", err, string(output))
	}

	fmt.Printf("✅ Helm chart %s deployed successfully on cluster %s\n", chartName, clusterName)
	return nil
}
