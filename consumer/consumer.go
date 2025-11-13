package consumer

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"push_service/models"
	sendNotification "push_service/sendNotification"
	"push_service/util"

	firebase "firebase.google.com/go/v4"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/api/option"
)

func StartConsumer(ch *amqp.Channel) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Note: .env file not found, reading from system environment")
	}

	c := models.Consumer{
		QueueName:     "",
		RetryQueue:    "retry_notif",
		PrefetchCount: 1,
		WorkerCount:   5,
	}

	SetUp(&c, ch)

	for r := 0; r < c.WorkerCount; r++ {
		if c.Channel != nil {
			go NewWorker(&c, r)
		} else {
			log.Fatal("couldn't launch workers")
		}
	}

	log.Printf(" [*] Started %d workers. Waiting for messages. To exit press CTRL+C", c.WorkerCount)
	forever := make(chan struct{})
	<-forever
}

func SetUp(c *models.Consumer, ch *amqp.Channel) {
	c.Channel = ch
	// Declare the notification dlx
	err := ch.ExchangeDeclare(
		models.DlxName, // name
		"direct",       // type (direct, fanout, topic, etc.)
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	util.FailOnError(err, "Failed to declare notifs dlx")

	err = ch.ExchangeDeclare(
		models.RetryExName, // name
		"direct",           // type (direct, fanout, topic, etc.)
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	)
	util.FailOnError(err, "Failed to declare retry notifs exchange")

	// Configure channel
	err = ch.Qos(
		c.PrefetchCount, // prefetch count
		0,               // prefetch size
		false,           // global
	)
	util.FailOnError(err, "Failed to set QoS")

	// Declare arguments for queue
	args := amqp.Table{
		"x-dead-letter-exchange":    models.DlxName,
		"x-dead-letter-routing-key": models.DlqRoutingKey,
	}

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,  // arguments
	)
	util.FailOnError(err, "Failed to declare main queue")
	c.QueueName = q.Name

	err = ch.QueueBind(
		q.Name,            // queue name
		models.RoutingKey, // routing key
		models.ExName,     // exchange
		false,
		nil)
	util.FailOnError(err, "Failed to bind main queue with main exchange")

	dlqName := "my-app-dlq"

	_, err = ch.QueueDeclare(
		dlqName, // name
		true,    // durable
		false,   // auto-deleted
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	util.FailOnError(err, "Failed to declare notifs dlq")

	err = ch.QueueBind(
		dlqName,              // queue name
		models.DlqRoutingKey, // routing key
		models.DlxName,       // exchange name
		false,                // no-wait
		nil,                  // arguments
	)
	util.FailOnError(err, "Failed to bind dlq to dlx")

	args = amqp.Table{
		"x-dead-letter-exchange":    models.ExName,
		"x-message-ttl":             models.RetryDelayMs,
		"x-dead-letter-routing-key": models.RoutingKey,
	}

	_, err = ch.QueueDeclare(
		models.RetryQueueName, // name
		true,                  // durable
		false,                 // auto-deleted
		false,                 // exclusive
		false,                 // no-wait
		args,                  // arguments
	)
	util.FailOnError(err, "Failed to declare retry notifs queue")

	err = ch.QueueBind(
		models.RetryQueueName,       // queue name
		models.RetryQueueRoutingKey, // routing key
		models.RetryExName,          // exchange name
		false,                       // no-wait
		nil,                         // arguments
	)
	util.FailOnError(err, "Failed to bind retry_notifs queue to retry_notifs exchange")

	setUpFirebaseClient(c)
}

func NewWorker(c *models.Consumer, id int) {
	msgs, err := c.Channel.Consume(
		c.QueueName, // queue
		"",          // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		log.Printf("Worker %d Failed to register as a consumer: %s", id, err)
		return
	}

	go func() {
		for d := range msgs {

			var headerRetryCount int64 = 1
			if val, ok := d.Headers["x-retry-count"]; ok {
				if count, ok := val.(int64); ok {
					headerRetryCount = count
				}
			}
			log.Printf("Worker %d Received a message (Attempt %d)", id, headerRetryCount)

			var notif models.NotifPushRequest
			err := json.Unmarshal(d.Body, &notif)
			if err != nil {
				log.Printf("Worker %d FAILED to unmarshal JSON: %v. Sending to DLX.", id, err)

				d.Nack(false, false)
				continue
			}

			err = sendNotification.SendNotification(context.Background(), c, notif)
			if err != nil {
				log.Printf("Worker failed: %v", err)
				if headerRetryCount < models.MaxRetries {
					log.Println("Started retrying")
					d.Ack(false)

					err = c.Channel.Publish(
						models.RetryExName,          // exchange
						models.RetryQueueRoutingKey, // routing key (use original)
						false,                       // mandatory
						false,                       // immediate
						amqp.Publishing{
							ContentType: d.ContentType,
							Body:        d.Body,
							Headers: amqp.Table{
								"x-retry-count": headerRetryCount + 1,
							},
						},
					)
					log.Println("Finished publishing retry")
					if err != nil {
						log.Printf("Error publishing to retry exchange: %s", err)
					}
				} else {
					log.Printf("Max retries (%d) exceeded. Sending to DLX.", models.MaxRetries)
					d.Nack(false, false)
					log.Printf("Sent to DLX successfully")
				}
			} else {
				if ackErr := d.Ack(false); ackErr != nil {
					log.Printf(" [Worker %d] Failed to ack: %v", id, ackErr)
				} else {
					log.Printf(" [Worker %d] Message acknowledged", id)
				}
			}
		}
	}()
}

func setUpFirebaseClient(c *models.Consumer) {
	ctx := context.Background()

	serviceAccountJSON := os.Getenv("GOOGLE_SERVICE_ACCOUNT")
	opt := option.WithCredentialsJSON([]byte(serviceAccountJSON))

	config := firebase.Config{
		ProjectID: "pushservice-8f271",
	}

	app, err := firebase.NewApp(ctx, &config, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
	}
	c.Client = client
}
