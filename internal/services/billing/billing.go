package billing

import (
	"encoding/json"
	"github.com/Vinayakatk/marketplace-prototype/pkg/database"
	"github.com/Vinayakatk/marketplace-prototype/pkg/models"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// Get billing history for a specific user
func GetUserBilling(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	var records []models.BillingRecord
	if err := database.DB.Where("consumer_id = ?", userID).Find(&records).Error; err != nil {
		http.Error(w, "Failed to fetch billing records", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(records)
}

// Get Billing data by User (Consumer) and Deployment ID
func GetBillingByUserAndDeployment(w http.ResponseWriter, r *http.Request) {
	consumerID := chi.URLParam(r, "consumerID")
	deploymentID := chi.URLParam(r, "deploymentID")

	var records []models.BillingRecord
	if err := database.DB.Where("consumer_id = ? AND deployment_id = ?", consumerID, deploymentID).Find(&records).Error; err != nil {
		http.Error(w, "Failed to fetch billing records", http.StatusInternalServerError)
		return
	}

	if len(records) == 0 {
		http.Error(w, "No billing records found for the given user and deployment", http.StatusNotFound)
		return
	}

	// Return the records
	json.NewEncoder(w).Encode(records)
}
