package helm

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

func DeployHelmChart(clusterName, repoURL, chartName, application, applicationID string) error {
	// Construct a unique repo name
	repoName := fmt.Sprintf("%s-%s", application, applicationID)

	// Check if the repo already exists
	listCmd := exec.Command("helm", "repo", "list", "--output", "json")
	output, err := listCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to list Helm repos: %v\n%s", err, string(output))
	}

	// Parse JSON output to check if the repo exists
	var repos []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(output, &repos); err != nil {
		return fmt.Errorf("failed to parse Helm repo list: %v\n%s", err, string(output))
	}

	// Check if the repo is already present
	repoExists := false
	for _, repo := range repos {
		if repo.Name == repoName {
			repoExists = true
			break
		}
	}

	if repoExists {
		log.Printf("✅ Helm repo %s already exists, skipping add", repoName)
	} else {
		// Add repo if it doesn't exist
		addCmd := exec.Command("helm", "repo", "add", repoName, repoURL)
		if addOutput, addErr := addCmd.CombinedOutput(); addErr != nil {
			return fmt.Errorf("failed to add repo: %v\n%s", addErr, string(addOutput))
		}
		log.Printf("✅ Helm repo %s added successfully", repoName)
	}

	// Update the Helm repo
	updateRepoCmd := exec.Command("helm", "repo", "update")
	if output, err := updateRepoCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to update repo: %v\n%s", err, string(output))
	}

	// Install the Helm chart using the unique repo name
	installCmd := exec.Command("helm", "install", chartName, repoName+"/"+chartName, "--kube-context", "kind-"+clusterName)
	output, err = installCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to deploy Helm chart: %v\n%s", err, string(output))
	}

	fmt.Printf("✅ Helm chart %s deployed successfully on cluster %s\n", chartName, clusterName)
	return nil
}
