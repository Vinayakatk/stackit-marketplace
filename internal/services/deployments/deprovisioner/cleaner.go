package deprovisioner

import (
	"fmt"
	"github.com/Vinayakatk/marketplace-prototype/pkg/database"
	"github.com/Vinayakatk/marketplace-prototype/pkg/models"
	"os/exec"
	"time"
)

type UninstallRequest struct {
	DeploymentID   string
	DeploymentType string
	ClusterName    string
	VMName         string
}

// ResourceCleaner defines an interface for cleaning up different deployment types.
type ResourceCleaner interface {
	Clean() error
}

// K8sCleaner implements ResourceCleaner for Kubernetes resources
type K8sCleaner struct {
	ClusterName string
}

func (k *K8sCleaner) Clean() error {
	if k.ClusterName == "" {
		fmt.Println("⚠️ No cluster name provided, skipping Kubernetes cleanup")
		return nil
	}
	fmt.Printf("🛑 Cleaning up Kubernetes cluster: %s\n", k.ClusterName)

	cmd := exec.Command("kind", "delete", "cluster", "--name", k.ClusterName)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("❌ Failed to delete Kubernetes cluster %s: %v\n", k.ClusterName, err)
		return err
	}
	fmt.Printf("✅ Kubernetes cluster %s cleaned up successfully\n", k.ClusterName)
	return nil
}

// VMCleaner implements ResourceCleaner for Virtual Machines
type VMCleaner struct {
	VMName string
}

func (v *VMCleaner) Clean() error {
	if v.VMName == "" {
		fmt.Println("⚠️ No VM name provided, skipping VM cleanup")
		return nil
	}
	fmt.Printf("🛑 Cleaning up VM instance: %s\n", v.VMName)

	time.Sleep(2 * time.Second) // Simulating VM deletion
	fmt.Printf("✅ VM instance %s cleaned up successfully\n", v.VMName)
	return nil
}

// CleanResource selects the correct cleaner based on DeploymentType
func CleanResource(req UninstallRequest) {
	var cleaner ResourceCleaner

	switch req.DeploymentType {
	case "k8s":
		cleaner = &K8sCleaner{ClusterName: req.ClusterName}
	case "vm":
		cleaner = &VMCleaner{VMName: req.VMName}
	default:
		fmt.Printf("⚠️ Unsupported deployment type: %s\n", req.DeploymentType)
		return
	}

	// Execute cleanup
	if err := cleaner.Clean(); err != nil {
		fmt.Printf("❌ Failed to clean resource: %v\n", err)
	}

	var deployment models.Deployment
	if err := database.DB.First(&deployment, req.DeploymentID).Error; err != nil {
		fmt.Printf("deployment not found: %v\n", err)
		return
	}
	// Delete deployment from DB
	if err := database.DB.Delete(&deployment).Error; err != nil {
		fmt.Printf("failed to delete deployment: %v\n", err)
		return
	}
}
