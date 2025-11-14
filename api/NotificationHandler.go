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
	"push_service/util"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	InternalServerError = "Internal server error"
	NotFound            = "Country not found"
	ValidationFailed    = "Validation failed"
)

// NotificationHandler godoc
// @Summary      Queues a push notification
// @Description  Accepts a push notification payload and sends it to the RabbitMQ queue
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        request  body      models.NotifPushRequest  true  "Notification request payload"
// @Success      202      {object}  map[string]string        "status: queued, request_id: string"
// @Failure      400      {object}  map[string]string        "error: validation failed or invalid notification type"
// @Failure      500      {object}  map[string]string        "error: internal server error"
// @Router       /notification [post]
func NotificationHandler(p *models.Publisher, ctx *gin.Context) {
	var req models.NotifPushRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Println(util.ErrorResponse(err))
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(errors.New(ValidationFailed)))
		return
	}

	body, err := json.Marshal(req)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(errors.New("Failed to process request")))
		return
	}

	if req.NotificationType != "push" {
		log.Printf("Invalid notification type: %s. This handler only accepts push notifications.", req.NotificationType)
		ctx.JSON(http.StatusBadRequest, util.ErrorResponse(errors.New("Invalid notification_type, expected push")))
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
		ctx.JSON(http.StatusInternalServerError, util.ErrorResponse(errors.New("Failed to queue notification")))
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
