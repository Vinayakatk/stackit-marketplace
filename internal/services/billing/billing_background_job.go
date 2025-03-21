package billing

import (
	"fmt"
	"github.com/Vinayakatk/marketplace-prototype/pkg/database"
	"github.com/Vinayakatk/marketplace-prototype/pkg/models"
	"log"
	"time"
)

func StartBillingUpdater() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		updateBillingRecords()
	}
}

func updateBillingRecords() {
	log.Println("🔄 Updating billing records...")

	// Fetch active billing records (where EndTime is NULL)
	var records []models.BillingRecord
	if err := database.DB.Where("end_time IS NULL").Find(&records).Error; err != nil {
		log.Println("❌ Failed to fetch billing records:", err)
		return
	}

	for _, record := range records {
		elapsedDuration := time.Since(record.StartTime)
		elapsedHours := elapsedDuration.Hours()
		elapsedMinutes := elapsedDuration.Minutes()

		newAmount := (elapsedMinutes / 60) * record.HourlyRate // Calculate cost based on total minutes

		// Update Billing Record
		if err := database.DB.Model(&record).Updates(models.BillingRecord{
			Amount:    newAmount,
			UpdatedAt: time.Now(),
		}).Error; err != nil {
			log.Println("❌ Failed to update billing:", err)
		} else {
			fmt.Printf("💰 Billing updated: %s → $%.2f (%.0f hours, %.0f mins)\n",
				record.DeploymentID, newAmount, elapsedHours, elapsedMinutes)
		}
	}
}
