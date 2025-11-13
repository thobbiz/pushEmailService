package sendNotification

import (
	"context"
	"fmt"
	"push_service/api"
	"push_service/models"

	"firebase.google.com/go/v4/messaging"
)

func SendNotification(ctx context.Context, c *models.Consumer, notifPushRequest models.NotifPushRequest) error {

	user, err := api.FetchUser(notifPushRequest.UserID)
	if err != nil {
		return err
	}

	dataPayload := map[string]string{
		"name":          user.Name,
		"link":          notifPushRequest.Variables.Link,
		"template_code": notifPushRequest.TemplateID,
	}

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: fmt.Sprintf("Hello, %s!", notifPushRequest.Variables.Name),
			Body:  fmt.Sprintf("%s", notifPushRequest.Variables.Link),
		},
		Data:  dataPayload,
		Token: user.PushToken,
	}

	if err = sendMessage(ctx, c, message); err != nil {
		return err
	}

	return nil
}

func sendMessage(ctx context.Context, c *models.Consumer, message *messaging.Message) error {
	response, err := c.Client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("error sending message: %v\n", err)
	}

	fmt.Printf("Successfully sent message: %s\n", response)
	return nil
}
