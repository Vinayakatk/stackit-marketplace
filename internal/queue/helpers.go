package queue

import (
	"fmt"
	"github.com/Vinayakatk/marketplace-prototype/internal/services/deployments/provisioner"
	"github.com/Vinayakatk/marketplace-prototype/pkg/database"
	"github.com/Vinayakatk/marketplace-prototype/pkg/models"
	"log"
	"strconv"
	"time"
)

func provisionApplication(installReq provisioner.InstallRequest) error {
	pv, err := provisioner.NewProvisioner(installReq)
	if err != nil {
		log.Println(err)
		return err
	}

	if err := pv.Provision(); err != nil {
		log.Println("❌ Provisioning failed:", err)
		return err
	}

	// Record Billing
	addBillingRecord(installReq)

	return nil
}

func addBillingRecord(installReq provisioner.InstallRequest) {
	applicationID := installReq.ApplicationID
	uintID, err := strconv.ParseUint(applicationID, 10, 32)
	if err != nil {
		log.Println("failed to convert application id to uint:", err)
	}

	var app models.Application
	if err := database.DB.Preload("Publisher").First(&app, uintID).Error; err != nil {
		log.Println("failed to find application:", err)
	}
	billing := models.BillingRecord{
		ID:            fmt.Sprintf("%s-bill", installReq.DeploymentID),
		ConsumerID:    installReq.ConsumerID,
		DeploymentID:  installReq.DeploymentID,
		ApplicationID: app.ID,
		HourlyRate:    app.HourlyRate,
		Amount:        0.0,
		StartTime:     time.Now(),
		CreatedAt:     time.Now(),
	}

	if err := database.DB.Create(&billing).Error; err != nil {
		log.Println("❌ Failed to create deployment record:", err)
	}
}
