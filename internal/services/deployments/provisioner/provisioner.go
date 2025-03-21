package provisioner

import (
	"fmt"
)

// InstallRequest represents a message in the queue
type InstallRequest struct {
	DeploymentID  string
	ConsumerID    string
	ApplicationID string
	Application   string
	DeployType    string
	RepoURL       string
	ChartName     string
	Inputs        map[string]interface{}
}

type Provisioner interface {
	Provision() error
}

func NewProvisioner(installReq InstallRequest) (Provisioner, error) {
	switch installReq.DeployType {
	case "k8s":
		return &KubernetesProvisioner{InstallReq: installReq}, nil
	case "vm":
		return &VMProvisioner{InstallReq: installReq}, nil
	default:
		return nil, fmt.Errorf("invalid deployment type: %s", installReq.DeployType)
	}
}
