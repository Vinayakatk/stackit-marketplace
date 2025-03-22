package deployments

import (
	"encoding/json"
	"fmt"
	"github.com/Vinayakatk/marketplace-prototype/internal/queue"
	"github.com/Vinayakatk/marketplace-prototype/internal/services/deployments/deprovisioner"
	"github.com/Vinayakatk/marketplace-prototype/internal/services/deployments/provisioner"
	"github.com/Vinayakatk/marketplace-prototype/pkg/database"
	"github.com/Vinayakatk/marketplace-prototype/pkg/models"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

type deploymentResponse struct {
	ID          uint `json:"id"`
	Application struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"application"`
	DeploymentType string `json:"deployment_type"`
	ClusterName    string `json:"cluster_name,omitempty"`
	VMName         string `json:"vm_name,omitempty"`
	VMIP           string `json:"vm_ip,omitempty"`
	Status         string `json:"status"`
}

// DeployApplication API (only for consumers)
func DeployApplication(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ConsumerID    uint `json:"consumer_id"`
		ApplicationID uint `json:"application_id"`
		ProjectID     uint `json:"project_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate Project Exists
	var project models.Project
	if err := database.DB.First(&project, req.ProjectID).Error; err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Fetch application details
	var app models.Application
	if err := database.DB.First(&app, req.ApplicationID).Error; err != nil {
		http.Error(w, "Application not found", http.StatusNotFound)
		return
	}

	// Check if the consumer exists
	var publisher models.User
	if err := database.DB.First(&publisher, req.ConsumerID).Error; err != nil {
		http.Error(w, "Publisher not found", http.StatusBadRequest)
		return
	}

	// Initialize Deployment
	deployment := models.Deployment{
		ConsumerID:     req.ConsumerID,
		ApplicationID:  req.ApplicationID,
		ProjectID:      req.ProjectID,
		DeploymentType: app.Deployment.Type,
		Status:         "pending", // Initial status
	}

	// Store Deployment Record (Initial Status)
	if err := database.DB.Create(&deployment).Error; err != nil {
		http.Error(w, "Failed to save deployment record", http.StatusInternalServerError)
		return
	}

	// Push to Redis Queue for Asynchronous Processing
	err := queue.PushToInstallerQueue(provisioner.InstallRequest{
		DeploymentID:  fmt.Sprintf("%d", deployment.ID),
		ConsumerID:    fmt.Sprintf("%d", req.ConsumerID),
		ApplicationID: fmt.Sprintf("%d", req.ApplicationID),
		Application:   app.Name,
		DeployType:    app.Deployment.Type,
		RepoURL:       app.Deployment.RepoURL,
		ChartName:     app.Deployment.ChartName,
		Inputs:        app.Inputs,
	})
	if err != nil {
		http.Error(w, "Failed to queue deployment", http.StatusInternalServerError)
		return
	}

	// Return Deployment ID
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Deployment request queued",
		"deploymentID": deployment.ID,
	})
}

func GetDeployment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id") // Get deployment ID from URL

	var deployment models.Deployment
	if err := database.DB.
		Preload("Application", func(db *gorm.DB) *gorm.DB {
			// Preload only the fields of Application you want (exclude Publisher)
			return db.Select("id, name, description")
		}).
		Select("id, application_id, deployment_type, cluster_name, vm_name, vm_ip, status").
		First(&deployment, id).Error; err != nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	// Define response DTO to exclude Consumer & Project
	response := deploymentResponse{
		ID: deployment.ID,
		Application: struct {
			ID          uint   `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
		}{
			ID:          deployment.Application.ID,
			Name:        deployment.Application.Name,
			Description: deployment.Application.Description,
		},
		DeploymentType: deployment.DeploymentType,
		ClusterName:    deployment.ClusterName,
		VMName:         deployment.VMName,
		VMIP:           deployment.VMIP,
		Status:         deployment.Status,
	}

	json.NewEncoder(w).Encode(response)
}

func DeleteDeployment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var deployment models.Deployment
	if err := database.DB.First(&deployment, id).Error; err != nil {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}

	// Queue a delete message
	queue.PushToUninstallerQueue(deprovisioner.UninstallRequest{
		DeploymentID:   id,
		DeploymentType: deployment.DeploymentType,
		ClusterName:    deployment.ClusterName,
		VMName:         deployment.VMName,
	})

	// Fetch Billing Record
	var billing models.BillingRecord
	if err := database.DB.Where("deployment_id = ?", id).First(&billing).Error; err != nil {
		http.Error(w, "Billing record not found", http.StatusNotFound)
		return
	}

	// Calculate final amount
	endTime := time.Now()
	totalHours := endTime.Sub(billing.StartTime).Hours()
	totalCost := totalHours * billing.HourlyRate

	// Update Billing Record with final amount
	billing.EndTime = &endTime
	billing.Amount = totalCost
	if err := database.DB.Save(&billing).Error; err != nil {
		http.Error(w, "Failed to save billing record", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListUserDeployments API to list deployments of a user
func ListUserDeployments(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from URL parameter
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get the 'status' query parameter (optional)
	status := r.URL.Query().Get("status")

	// Define the allowed statuses for validation
	validStatuses := []string{"installing", "installed", "failed", "pending"}
	isValidStatus := false
	for _, s := range validStatuses {
		if status == s {
			isValidStatus = true
			break
		}
	}

	// If status is invalid, return a Bad Request response
	if status != "" && !isValidStatus {
		http.Error(w, "Invalid status. Valid statuses are: installing, installed, failed, pending", http.StatusBadRequest)
		return
	}

	// Create a slice to hold the deployments
	var deployments []models.Deployment

	// Query deployments where the ConsumerID matches the userID, preloading the Application model
	query := database.DB.Preload("Application", func(db *gorm.DB) *gorm.DB {
		// Preload only necessary fields of the Application model
		return db.Select("id, name, description")
	}).Where("consumer_id = ?", userID)

	// If a status is provided, filter by status
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Execute the query to fetch the deployments
	if err := query.Find(&deployments).Error; err != nil {
		http.Error(w, "Error fetching user deployments", http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := make([]deploymentResponse, 0)
	for _, deployment := range deployments {
		response = append(response, deploymentResponse{
			ID: deployment.ID,
			Application: struct {
				ID          uint   `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
			}{
				ID:          deployment.Application.ID,
				Name:        deployment.Application.Name,
				Description: deployment.Application.Description,
			},
			DeploymentType: deployment.DeploymentType,
			ClusterName:    deployment.ClusterName,
			VMName:         deployment.VMName,
			VMIP:           deployment.VMIP,
			Status:         deployment.Status,
		})
	}

	// Return the deployments as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
