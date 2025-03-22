package projects

import (
	"encoding/json"
	"fmt"
	"github.com/Vinayakatk/marketplace-prototype/pkg/database"
	"github.com/Vinayakatk/marketplace-prototype/pkg/models"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func CreateProject(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string `json:"name"`
		UserID uint   `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	project := models.Project{Name: req.Name, UserID: req.UserID}
	if err := database.DB.Create(&project).Error; err != nil {
		http.Error(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": fmt.Sprintf("Project: %s created successfully", project.Name),
	})
}

func ListProjects(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")
	userID, err := strconv.ParseUint(userId, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	var projects []models.Project
	if err := database.DB.Preload("User").Preload("Deployments").Where("user_id = ?", userID).Find(&projects).Error; err != nil {
		http.Error(w, "Failed to fetch projects", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(projects)
}

func GetDeploymentsOfAProject(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	var project models.Project
	if err := database.DB.Preload("Deployments").First(&project, projectID).Error; err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(project.Deployments)
}

func DeleteProject(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	// Check if project exists
	var project models.Project
	if err := database.DB.First(&project, projectID).Error; err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Check if the project has deployments
	var deploymentCount int64
	if err := database.DB.Model(&models.Deployment{}).Where("project_id = ?", projectID).Count(&deploymentCount).Error; err != nil {
		http.Error(w, "Failed to check deployments", http.StatusInternalServerError)
		return
	}

	if deploymentCount > 0 {
		http.Error(w, "Cannot delete project with active deployments", http.StatusForbidden)
		return
	}

	// Delete project if no deployments exist
	if err := database.DB.Delete(&project).Error; err != nil {
		http.Error(w, "Failed to delete project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
