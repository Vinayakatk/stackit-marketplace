package users

import (
	"encoding/json"
	"fmt"
	"github.com/Vinayakatk/marketplace-prototype/pkg/database"
	"github.com/Vinayakatk/marketplace-prototype/pkg/models"
	"net/http"
)

// CreateUser API
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user := models.User{Name: req.Name}
	if err := database.DB.Create(&user).Error; err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": fmt.Sprintf("User: %s created successfully", user.Name),
	})
}

// ListUsers API
func ListUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	database.DB.Find(&users)
	json.NewEncoder(w).Encode(users)
}
