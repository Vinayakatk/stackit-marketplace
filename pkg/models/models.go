package models

import "time"

type User struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique"`
}

// Project represents a group of deployments under a user
type Project struct {
	ID     uint   `gorm:"primaryKey"`
	Name   string `gorm:"unique"`
	UserID uint   // Owner of the project

	User        User         `gorm:"foreignKey:UserID"`
	Deployments []Deployment `gorm:"foreignKey:ProjectID"`
}

type Application struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"unique"`
	Description string
	PublisherID uint
	HourlyRate  float64        // 💰 Cost per hour
	Deployment  DeploymentSpec `gorm:"embedded"` // Embedded struct for deployment details
	Publisher   User           `gorm:"foreignKey:PublisherID"`

	Inputs map[string]interface{} `gorm:"type:jsonb"` // Store input fields as JSON
}

// DeploymentSpec stores deployment-related data
type DeploymentSpec struct {
	Type      string `gorm:"type:varchar(10)"` // "k8s" or "vm"
	RepoURL   string // Only for Kubernetes-based apps
	ChartName string // Only for Kubernetes-based apps
	Image     string // VM image for VM-based apps
	CPU       string // VM CPU configuration (e.g., "2 vCPUs")
	Memory    string // VM Memory configuration (e.g., "4GB RAM")
}

type Deployment struct {
	ID             uint `gorm:"primaryKey"`
	ConsumerID     uint
	ApplicationID  uint
	ProjectID      uint   // The project under which this deployment is managed
	DeploymentType string `gorm:"type:varchar(10)"` // "k8s" or "vm"

	// Kubernetes-specific
	ClusterName string `gorm:"default:null"` // KIND cluster name (if K8s-based)

	// VM-specific
	VMName string `gorm:"default:null"` // VM instance name (if VM-based)
	VMIP   string `gorm:"default:null"` // IP of the created VM

	// Deployment status
	Status string `gorm:"type:varchar(20);default:'pending'"` // Possible values: "pending", "installing", "installed", "failed"

	Consumer    User        `gorm:"foreignKey:ConsumerID"`
	Application Application `gorm:"foreignKey:ApplicationID"`
	Project     Project     `gorm:"foreignKey:ProjectID"`
}

type BillingRecord struct {
	ID            string     `gorm:"primaryKey"`
	ConsumerID    string     `gorm:"index"`
	DeploymentID  string     `gorm:"index"`
	ApplicationID uint       `gorm:"index"`
	HourlyRate    float64    // 💰 Cost per hour
	Amount        float64    // 🔄 Total amount (updated hourly)
	StartTime     time.Time  // 📅 Start timestamp
	EndTime       *time.Time `gorm:"default:null"` // 📅 End timestamp (null if running)
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
