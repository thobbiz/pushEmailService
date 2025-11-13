package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"push_service/models"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	InternalServerError = "Internal server error"
	NotFound            = "Country not found"
	ValidationFailed    = "Validation failed"
)

func NotificationHandler(p *models.Publisher, ctx *gin.Context) {
	var req models.NotifPushRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Println(errorResponse(err))
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New(ValidationFailed)))
		return
	}

	body, err := json.Marshal(req)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New("Failed to process request")))
		return
	}

	if req.NotificationType != "push" {
		log.Printf("Invalid notification type: %s. This handler only accepts push notifications.", req.NotificationType)
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("Invalid notification_type, expected push")))
		return
	}

	err = p.Channel.PublishWithContext(
		context.Background(),
		models.ExName, // exchange
		"notifs",      // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         []byte(body),
		})
	if err != nil {
		log.Printf("Failed to publish a message: %v", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(errors.New("Failed to queue notification")))
		return
	}
	log.Printf(" [x] Sent %s\n", body)

	ctx.JSON(http.StatusAccepted, gin.H{"status": "queued", "request_id": req.RequestId})
}

func FetchUser(userId string) (*models.User, error) {
	API_NAME := "https://user-service-giq0.onrender.com"
	url := fmt.Sprintf("%s/v1/users/%s", API_NAME, userId)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil || (resp.StatusCode != http.StatusOK) {
		return nil, fmt.Errorf("could not fetch data from User service [%s]: %w", API_NAME, err)
	}
	log.Printf("User service responded: %d", resp.StatusCode)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var user models.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user JSON: %w", err)
	}

	log.Printf("user is %v", fmt.Sprintf("%v", user))
	return &user, nil
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
