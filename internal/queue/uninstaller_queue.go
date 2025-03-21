package queue

import (
	"context"
	"fmt"
	"github.com/Vinayakatk/marketplace-prototype/internal/services/deployments/deprovisioner"
	redis "github.com/redis/go-redis/v9"
	"log"
	"time"
)

// Stream Name
const UninstallerQueue = "delete_queue"

// PushToUninstallerQueue pushes a delete message to Redis
func PushToUninstallerQueue(req deprovisioner.UninstallRequest) error {
	ctx := context.Background()

	_, err := redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: UninstallerQueue,
		Values: map[string]interface{}{
			"deployment_id":   req.DeploymentID,
			"deployment_type": req.DeploymentType,
			"cluster_name":    req.ClusterName,
			"vm_name":         req.VMName,
		},
	}).Result()

	if err != nil {
		log.Println("❌ Failed to push delete request to queue:", err)
	}
	return err
}

// StartUninstallerConsumer processes delete messages
func StartUninstallerConsumer() {
	ctx := context.Background()
	log.Println("🚀 Redis Delete Queue Consumer Started...")

	for {
		// Read messages from the delete queue
		messages, err := redisClient.XRead(ctx, &redis.XReadArgs{
			Streams: []string{UninstallerQueue, "0"},
			Count:   1,
			Block:   0, // Blocks indefinitely
		}).Result()

		if err != nil {
			log.Println("❌ Error reading from delete queue:", err)
			time.Sleep(2 * time.Second) // Retry delay
			continue
		}

		for _, stream := range messages {
			for _, message := range stream.Messages {
				deleteReq := deprovisioner.UninstallRequest{
					DeploymentID:   message.Values["deployment_id"].(string),
					DeploymentType: message.Values["deployment_type"].(string),
					ClusterName:    message.Values["cluster_name"].(string),
					VMName:         message.Values["vm_name"].(string),
				}

				fmt.Printf("🗑️  Processing Delete Request for Deployment %s\n", deleteReq.DeploymentID)

				// Perform deletion
				deprovisioner.CleanResource(deleteReq)

				// Acknowledge message deletion
				_, err := redisClient.XDel(ctx, UninstallerQueue, message.ID).Result()
				if err != nil {
					log.Println("❌ Failed to acknowledge delete message:", err)
				}
			}
		}
	}
}
