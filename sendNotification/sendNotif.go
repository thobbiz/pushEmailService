package sendNotification

import (
	"context"
	"fmt"
	"log"
	"push_service/api"
	"push_service/models"

	"firebase.google.com/go/v4/messaging"
)

func SendNotification(ctx context.Context, c *models.Consumer, notifMessageRequest models.NotifMessageRequest) error {
	token, title, body, err := resolveNotificationContent(notifMessageRequest)
	if err != nil {
		return fmt.Errorf("could not resolve notification content: %w", err)
	}

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data:  convertMetaToStringMap(notifMessageRequest.Variables),
		Token: token,
	}

	if err := sendMessage(ctx, c, message); err != nil {
		return err
	}

	return nil
}

func sendMessage(ctx context.Context, c *models.Consumer, message *messaging.Message) error {
	_, err := c.Client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("Error sending message: %v\n", err)
	}

	fmt.Printf("Successfully sent message")
	return nil
}

func convertMetaToStringMap(meta map[string]any) map[string]string {
	if meta == nil {
		return nil
	}

	result := make(map[string]string)
	for key, value := range meta {
		result[key] = fmt.Sprintf("%v", value)
	}
	return result
}

func resolveNotificationContent(req models.NotifMessageRequest) (token, title, body string, err error) {
	if req.PushToken == nil || *req.PushToken == "" {
		user, err := api.FetchUser(req.UserID)
		if err != nil {
			log.Println("Couldn't fetch user ")
			return "", "", "", fmt.Errorf("Couldn't fetch user: %v", err)
		}
		token = user.PushToken
	} else {
		token = *req.PushToken
	}

	var template *models.TemplateResponse

	if req.Body == nil || req.Title == nil || *req.Body == "" || *req.Title == "" {
		template, err = api.FetchTemplate("welcome_push")
		if err != nil {
			log.Println("Couldn't fetch template ")
			return "", "", "", fmt.Errorf("Couldn't fetch template: %v", err)
		}
	}

	if req.Body == nil {
		body = template.Body
	} else {
		body = *req.Body
	}

	if req.Title == nil {
		title = template.Name
	} else {
		title = *req.Title
	}

	return token, title, body, nil
}
