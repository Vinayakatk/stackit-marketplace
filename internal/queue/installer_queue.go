package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Vinayakatk/marketplace-prototype/internal/services/deployments/provisioner"
	"github.com/Vinayakatk/marketplace-prototype/pkg/database"
	"github.com/Vinayakatk/marketplace-prototype/pkg/models"
	redis "github.com/redis/go-redis/v9"
	"log"
	"time"
)

// Redis client
var redisClient = redis.NewClient(&redis.Options{
	Addr: "localhost:6370",
})

const InstallerQueue = "install_queue"

// PushToInstallerQueue pushes a message to Redis
func PushToInstallerQueue(req provisioner.InstallRequest) error {
	ctx := context.Background()

	// Convert Inputs to JSON
	inputsJSON, err := json.Marshal(req.Inputs)
	if err != nil {
		log.Println("❌ Failed to marshal inputs:", err)
		return err
	}

	_, err = redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: InstallerQueue,
		Values: map[string]interface{}{
			"deployment_id":  req.DeploymentID,
			"consumer_id":    req.ConsumerID,
			"application_id": req.ApplicationID,
			"application":    req.Application,
			"deploy_type":    req.DeployType,
			"repo_url":       req.RepoURL,
			"chart_name":     req.ChartName,
			"inputs":         inputsJSON,
		},
	}).Result()

	if err != nil {
		log.Println("❌ Failed to push to queue:", err)
	}
	return err
}

// StartInstallerConsumer processes deployment messages
func StartInstallerConsumer() {
	ctx := context.Background()
	log.Println("🚀 Redis Create Queue Consumer Started...")

	for {
		// Read new messages from the queue
		messages, err := redisClient.XRead(ctx, &redis.XReadArgs{
			Streams: []string{InstallerQueue, "0"},
			Count:   1,
			Block:   0, // Blocks indefinitely
		}).Result()

		if err != nil {
			log.Println("❌ Error reading from queue:", err)
			time.Sleep(2 * time.Second) // Retry after delay
			continue
		}

		for _, stream := range messages {
			for _, message := range stream.Messages {
				var inputs map[string]interface{}
				if err := json.Unmarshal([]byte(message.Values["inputs"].(string)), &inputs); err != nil {
					log.Println("❌ Failed to unmarshal inputs:", err)
					continue
				}
				installReq := provisioner.InstallRequest{
					DeploymentID:  message.Values["deployment_id"].(string),
					ConsumerID:    message.Values["consumer_id"].(string),
					ApplicationID: message.Values["application_id"].(string),
					Application:   message.Values["application"].(string),
					DeployType:    message.Values["deploy_type"].(string),
					RepoURL:       message.Values["repo_url"].(string),
					ChartName:     message.Values["chart_name"].(string),
					Inputs:        inputs,
				}

				fmt.Printf("📦 Processing Deployment %s for User %s: %s\n", installReq.DeploymentID, installReq.ConsumerID, installReq.Application)

				// Fetch Deployment Record
				var deployment models.Deployment
				if err := database.DB.First(&deployment, installReq.DeploymentID).Error; err != nil {
					log.Println("❌ Deployment not found:", err)
					continue
				}

				// Update status to "installing"
				database.DB.Model(&deployment).Update("status", "installing")

				if err := provisionApplication(installReq); err != nil {
					continue
				}

				fmt.Printf("✅ Deployment %s completed for %s\n", installReq.DeploymentID, installReq.Application)

				// Remove Processed Message from Queue
				_, err = redisClient.XDel(ctx, InstallerQueue, message.ID).Result()
				if err != nil {
					log.Println("❌ Failed to acknowledge message:", err)
				}
			}
		}
	}
}
