package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
// @Param        request  body      models.NotifMessageRequest  true  "Notification request payload"
// @Success      202      {object}  map[string]string        "status: queued, request_id: string"
// @Failure      400      {object}  map[string]string        "error: validation failed or invalid notification type"
// @Failure      500      {object}  map[string]string        "error: internal server error"
// @Router       /notification [post]
func NotificationHandler(p *models.Publisher, ctx *gin.Context) {
	var req models.NotifMessageRequest
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
	log.Printf("Received Message")

	ctx.JSON(http.StatusAccepted, gin.H{"status": "queued", "request_id": req.UserID})
}

func FetchUser(userId string) (*models.User, error) {
	API_NAME := "https://user-service-yci9.onrender.com/api"
	url := fmt.Sprintf("%s/v1/users/%s", API_NAME, userId)
	token := os.Getenv("LOGIN_USER_TOKEN")

	client := &http.Client{
		Timeout: 40 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not make new request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch data from User service [%s]: %w", API_NAME, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user service responded with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var user models.UserResponses
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user JSON: %w", err)
	}

	log.Printf("user is %v", fmt.Sprintf("%v", user))
	return &user.Data, nil
}

func FetchTemplate(template_name string) (*models.TemplateResponse, error) {
	API_NAME := "https://template-service-mza0.onrender.com"
	url := fmt.Sprintf("%s/templates/name/%s", API_NAME, template_name)

	client := &http.Client{
		Timeout: 40 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not make new request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch data from Template service [%s]: %w", API_NAME, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("template service responded with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var template models.TemplateResponse
	err = json.Unmarshal(body, &template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template JSON: %w", err)
	}

	log.Printf("template was retrieved successfully")
	return &template, nil
}
