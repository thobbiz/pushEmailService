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

// NotificationType defines the type of notification.
// @Enum
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

// ConsumerMetrics holds the aggregated metrics for the consumer service.
type ConsumerMetrics struct {
	MessagesProcessed int `json:"messages_processed"` // Total messages consumed from the queue
	MessagesSucceeded int `json:"messages_succeeded"` // Total messages successfully processed
	MessagesFailed    int `json:"messages_failed"`    // Total messages that failed processing
	MessagesRetried   int `json:"messages_retried"`   // Total messages that were retried
}

// UserData holds user-specific information for the notification.
// This data is used to populate templates.
type UserData struct {
	Name string         `json:"name"`                                // User's full name
	Link string         `json:"link"`                                // A relevant URL for the notification (e.g., verification link)
	Meta map[string]any `json:"meta,omitempty" swaggertype:"object"` // Extra metadata for template substitution
}

// NotifPushRequest represents the request body for sending notifications
type NotifPushRequest struct {
	NotificationType NotificationType `json:"notification_type" example:"push" enums:"push,email"` // The type of notification to send
	UserID           string           `json:"user_id" example:"user1ab2c"`                         // The unique identifier for the user
	TemplateID       string           `json:"template_code" example:"welcome_template"`            // The code/ID of the template to use
	Variables        UserData         `json:"variables"`                                           // The dynamic data to inject into the template
	RequestId        string           `json:"request_id" example:"req_abc123"`                     // A unique ID for tracking this request
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
