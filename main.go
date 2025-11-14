package main

import (
	// "context"

	"log"
	"os"
	"push_service/api"
	"push_service/consumer"
	_ "push_service/docs"
	"push_service/models"
	"push_service/util"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go" // RabbitMQ client

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Push Email MicroService
// @version         1.0
// @description     A microservice that pushes notification to the designated web user
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  odelolatojumi@gmail.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Note: .env file not found, reading from system environment")
	}

	amqpURL := os.Getenv("RABBITMQ_URL")
	if amqpURL == "" {
		log.Fatal("RABBITMQ_URL is not set")
	}

	conn, err := amqp.Dial(amqpURL)
	util.FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	pubChannel, err := conn.Channel()
	util.FailOnError(err, "Failed to open publisher channel")
	defer pubChannel.Close()

	conChannel, err := conn.Channel()
	util.FailOnError(err, "Failed to open consumer channel")
	defer conChannel.Close()

	err = conChannel.ExchangeDeclare(
		models.ExName,
		"direct",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	util.FailOnError(err, "Failed to declare main exchange")

	p := models.Publisher{
		Channel: pubChannel,
	}

	c := models.Consumer{
		QueueName:     "",
		RetryQueue:    "retry_notif",
		PrefetchCount: 1,
		WorkerCount:   5,
	}

	log.Println("[Main] Starting background consumer workers...")
	go consumer.StartConsumer(conChannel, &c)

	router := gin.Default()

	// Swagger route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/health", func(ctx *gin.Context) {
		api.HealthHandler(&c, ctx)
	})

	router.POST("/notification", func(ctx *gin.Context) {
		api.NotificationHandler(&p, ctx)
	})

	log.Println("API Starting API server on :8080")
	router.Run(":8080")
}
