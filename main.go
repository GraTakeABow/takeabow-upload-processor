package main // import "github.com/therealpenguin/takeabow-upload-processor"
import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/streadway/amqp"
	"github.com/therealpenguin/takeabow-upload-processor/app"
	"log"
	"os"
)

func main() {
	// Get config from environment
	a, err := app.New()
	failOnError(err, "Couldn't create app")

	// Establish RabbitMQ connection to the uploads channel
	conn, err := amqp.Dial(os.Getenv(app.EnvAMQPUrl))

	if err != nil {
		failOnError(err, "Failed to connect to RabbitMQ")
	}
	defer conn.Close()

	ch, err := conn.Channel()

	_, err = ch.QueueDeclare(
		app.ChannelUploads, // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	failOnError(err, "Failed to declare the uploads queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")
	a.Ch = ch

	// Establish connection to database
	dsn := os.Getenv(app.EnvMYSQLDsn)
	if dsn == "" {
		err = errors.New(fmt.Sprintf(app.TemplateEmpty, app.EnvMYSQLDsn))
		failOnError(err, "Couldn't connect to MySQL")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		failOnError(err, "Couldn't connect to MySQL")
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		failOnError(err, "Couldn't connect to MySQL")
	}

	a.DB = db

	err = a.Run()
	if err != nil {
		failOnError(err, "Error")
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
