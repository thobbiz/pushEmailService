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

	ConsumerMetrics ConsumerMetrics

	PrefetchCount int
	WorkerCount   int
	RetryCount    int64

	Client *messaging.Client
}

type ConsumerMetrics struct {
	MessagesProcessed int `json:"messages_processed"`
	MessagesSucceeded int `json:"messages_succeeded"`
	MessagesFailed    int `json:"messages_failed"`
	MessagesRetried   int `json:"messages_retried"`
}

// UserData holds user-specific information for the notification.
type UserData struct {
	Name string         `json:"name" example:"John Doe"`
	Link string         `json:"link" example:"https://example.com/verify"`
	Meta map[string]any `json:"meta,omitempty" swaggertype:"object" example:"key1:value1,key2:value2"`
}

// NotifPushRequest represents the request body for sending notifications
type NotifPushRequest struct {
	NotificationType NotificationType `json:"notification_type" example:"push" enums:"push,email"`
	UserID           string           `json:"user_id" example:"user123"`
	TemplateID       string           `json:"template_code" example:"welcome_template"`
	Variables        UserData         `json:"variables"`
	RequestId        string           `json:"request_id" example:"req_abc123"`
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
