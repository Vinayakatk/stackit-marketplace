package catalog

import (
	"encoding/json"
	"fmt"
	"github.com/Vinayakatk/marketplace-prototype/pkg/database"
	"github.com/Vinayakatk/marketplace-prototype/pkg/models"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type deploymentSpec struct {
	Type      string `json:"type"`
	RepoURL   string `json:"repoURL"`
	ChartName string `json:"chartName"`
	Image     string `json:"image"`
	CPU       string `json:"cpu,omitempty"`
	Memory    string `json:"memory,omitempty"`
}

// AddApplication API (only for publishers)
func AddApplication(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		PublisherID uint                   `json:"publisher_id"`
		HourlyRate  float64                `json:"hourly_rate"`
		Deployment  deploymentSpec         `json:",inline"`
		Inputs      map[string]interface{} `json:"inputs"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate Deployment Type
	if req.Deployment.Type != "k8s" && req.Deployment.Type != "vm" {
		http.Error(w, "Invalid deployment type", http.StatusBadRequest)
		return
	}

	// Check if the publisher exists
	var publisher models.User
	if err := database.DB.First(&publisher, req.PublisherID).Error; err != nil {
		http.Error(w, "Publisher not found", http.StatusBadRequest)
		return
	}

	// Create the application
	app := models.Application{
		Name:        req.Name,
		Description: req.Description,
		PublisherID: req.PublisherID,
		HourlyRate:  req.HourlyRate,
		Deployment:  models.DeploymentSpec(req.Deployment),
		Inputs:      req.Inputs, // Set the dynamic inputs
	}

	if err := database.DB.Create(&app).Error; err != nil {
		http.Error(w, "Failed to add application", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": fmt.Sprintf("Application: %s created successfully", app.Name),
	})
}

func ListApplications(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	publisherName := r.URL.Query().Get("publisher")       // Filter by publisher name
	deploymentType := r.URL.Query().Get("deploymentType") // Filter by deployment type (k8s/vm)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Query builder
	query := database.DB.Preload("Publisher").Model(&models.Application{})

	// Apply filters
	if publisherName != "" {
		query = query.Joins("JOIN users ON users.id = applications.publisher_id").
			Where("users.name = ?", publisherName)
	}
	if deploymentType != "" {
		query = query.Where("type = ?", deploymentType)
	}

	var total int64
	query.Count(&total)

	var applications []models.Application
	result := query.Limit(limit).Offset(offset).Find(&applications)
	if result.Error != nil {
		http.Error(w, "Failed to fetch applications", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"data":        applications,
		"page":        page,
		"limit":       limit,
		"total_items": total,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetApplication(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id") // Get ID from URL

	var app models.Application
	if err := database.DB.Preload("Publisher").First(&app, id).Error; err != nil {
		http.Error(w, "Application not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(app)
}

func UpdateApplication(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var app models.Application
	if err := database.DB.First(&app, id).Error; err != nil {
		http.Error(w, "Application not found", http.StatusNotFound)
		return
	}

	var req struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Deployment  models.DeploymentSpec  `json:"deployment"`
		Inputs      map[string]interface{} `json:"inputs"` // Input fields as JSON
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Update application details
	app.Name = req.Name
	app.Description = req.Description
	app.Deployment = req.Deployment
	app.Inputs = req.Inputs // Update the inputs

	if err := database.DB.Save(&app).Error; err != nil {
		http.Error(w, "Failed to update application", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(app)
}

func DeleteApplication(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Check if the application has active deployments
	var count int64
	database.DB.Model(&models.Deployment{}).Where("application_id = ?", id).Count(&count)
	if count > 0 {
		http.Error(w, "Cannot delete: Active deployments exist", http.StatusConflict)
		return
	}

	// Delete application
	if err := database.DB.Delete(&models.Application{}, id).Error; err != nil {
		http.Error(w, "Failed to delete application", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
