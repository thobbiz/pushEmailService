package main

import (
	// "context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"push_service/api"
	"push_service/consumer"
	"push_service/models"
	"push_service/util"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go" // RabbitMQ client
)

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

	log.Println("[Main] Starting background consumer workers...")
	go consumer.StartConsumer(conChannel)

	router := gin.Default()
	router.POST("/notification", func(ctx *gin.Context) {
		api.NotificationHandler(&p, ctx)
	})

	log.Println("API Starting API server on :8080")
	router.Run(":8080")
}

func connectPostgreSQL() {
	// Construct the connection string (DSN)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		os.Getenv("DATABASE_HOST"),
		os.Getenv("DATABASE_PORT"),
		os.Getenv("DATABASE_USERNAME"),
		os.Getenv("DATABASE_PASSWORD"),
		os.Getenv("DATABASE_NAME"),
	)

	// Open the connection
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Printf("Failed to open PostgreSQL connection: %v\n", err)
		return
	}
	defer db.Close()

	// Ping the database to verify the connection
	err = db.Ping()
	if err != nil {
		log.Printf("Failed to ping PostgreSQL: %v\n", err)
		return
	}

	log.Println("✅ Successfully connected to PostgreSQL (Aiven)!")
}

// connectUpstashRedis connects to your Upstash Redis REST API
// func connectUpstashRedis() {
// 	client := upstash.NewClient(
// 		os.Getenv("UPSTASH_REDIS_REST_URL"),
// 		os.Getenv("UPSTASH_REDIS_REST_TOKEN"),
// 	)

// 	// Ping the server
// 	ctx := context.Background()
// 	res, err := client.Ping(ctx)
// 	if err != nil {
// 		log.Printf("Failed to ping Upstash Redis: %v\n", err)
// 		return
// 	}

// 	log.Printf("✅ Successfully connected to Upstash Redis! (Ping response: %s)\n", res)
// }

// // connectRabbitMQ connects to your CloudAMQP RabbitMQ instance
// func connectRabbitMQ() {
// 	// Get the connection URL from the environment
// 	amqpURL := os.Getenv("RABBITMQ_URL")
// 	if amqpURL == "" {
// 		log.Println("RABBITMQ_URL is not set")
// 		return
// 	}

// 	// Dial the server
// 	conn, err := amqp.Dial(amqpURL)
// 	if err != nil {
// 		log.Printf("Failed to connect to RabbitMQ: %v\n", err)
// 		return
// 	}
// 	defer conn.Close()

// 	log.Println("✅ Successfully connected to RabbitMQ (CloudAMQP)!")
// }
