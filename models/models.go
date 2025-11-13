package models

import (
	"firebase.google.com/go/v4/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	Email                NotificationType = "email"
	Push                 NotificationType = "push"
	MaxRetries                            = 2
	RetryDelayMs                          = int64(5000)
	RetryExName                           = "retry-notifs_ex"
	RetryQueueName                        = "retry-notifs_queue"
	RetryQueueRoutingKey                  = "retried-messages"
	RoutingKey                            = "notifs"
	ExName                                = "push_notifs"
	DlqRoutingKey                         = "failed-messages"
	DlxName                               = "push_notifs_dlx"
	DlqName                               = "push_notifs_dlq"
	Token                                 = "e2SUbDFyiaLMoIjmSe6bDl:APA91bEYcdOP4yPHLdZdS9ZdHz0wvfZRDZVqXsV1nkLQzm5FmUfJ8yUOKyJYvF8ZTq5wgA4jc800KEUcbQjZRVlMDHVwC8cSX574yZyDqVt5iEVegavJ-YU"
)

type NotificationType string

type Publisher struct {
	Channel *amqp.Channel
}

type Consumer struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
	QueueName  string
	RetryQueue string

	PrefetchCount int
	WorkerCount   int
	RetryCount    int64

	Client *messaging.Client
}

type ConsumerMetrics struct {
	messagesProcessed int
	messagesSucceeded int
	messagesFailed    int
	messagesRetried   int
}

// UserData holds user-specific information for the notification.
type UserData struct {
	Name string         `json:"name"`
	Link string         `json:"link"`
	Meta map[string]any `json:"meta,omitempty"`
}

type NotifPushRequest struct {
	NotificationType NotificationType `json:"notification_type"`
	UserID           string           `json:"user_id"`
	TemplateID       string           `json:"template_code"`
	Variables        UserData         `json:"variables"`
	RequestId        string           `json:"request_id"`
}

type User struct {
	Name        string      `json:"name"`
	Email       string      `json:"email"`
	PushToken   string      `json:"push_token"`
	Preferences Preferences `json:"preferences"`
}

type Preferences struct {
	Email bool `json:"email"`
	Push  bool `json:"push"`
}
