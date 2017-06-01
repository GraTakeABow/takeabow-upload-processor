package app

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/streadway/amqp"
	"github.com/therealpenguin/takeabow-upload-processor/processor"
	"github.com/therealpenguin/takeabow-upload-processor/video"
	"log"
	"os"
)

const EnvBucket = "BOW_BUCKET"
const EnvProcessedPrefix = "BOW_PREFIX_PROCESSED"
const EnvSmallPrefix = "BOW_PREFIX_SMALL"
const EnvAMQPUrl = "BOW_AMQP"
const EnvMYSQLDsn = "BOW_MYSQL_DSN"
const EnvTmpDir = "BOW_TMP_DIR"

const ChannelUploads = "uploads"

const TemplateEmpty = "%s is empty"

// App holds a valid configuration and some dependencies for the upload processor
type App struct {
	Ch              *amqp.Channel
	S3              *s3.S3
	Sess            *session.Session
	DB              *sql.DB
	Bucket          string
	ProcessedPrefix string
	TmpDir          string
	processor       *processor.Processor
	SmallPrefix     string
}

// NewVideoRequest creates and validates the application's config
func New() (*App, error) {
	a := &App{
		Bucket:          os.Getenv(EnvBucket),
		ProcessedPrefix: os.Getenv(EnvProcessedPrefix),
		TmpDir:          os.Getenv(EnvTmpDir),
		SmallPrefix:     os.Getenv(EnvSmallPrefix),
	}

	if a.Bucket == "" {
		return nil, errors.New(fmt.Sprintf(TemplateEmpty, EnvBucket))
	}

	if a.ProcessedPrefix == "" {
		return nil, errors.New(fmt.Sprintf(TemplateEmpty, EnvProcessedPrefix))
	}

	if a.TmpDir == "" {
		return nil, errors.New(fmt.Sprintf(TemplateEmpty, EnvTmpDir))
	}

	if a.SmallPrefix == "" {
		return nil, errors.New(fmt.Sprintf(TemplateEmpty, EnvSmallPrefix))
	}

	// Make the temp directory if it doesn't exist
	_, err := os.Stat(a.TmpDir)
	if err != nil {
		if err != os.ErrNotExist {
			return nil, err
		}

		err = os.Mkdir(a.TmpDir, 0755)

		if err != nil {
			return nil, err
		}
	}

	creds := credentials.NewEnvCredentials()
	_, err = creds.Get()
	if err != nil {
		return nil, err
	}

	session, err := session.NewSession(aws.NewConfig().WithRegion("eu-west-1").WithCredentials(creds))
	if err != nil {
		return nil, err
	}
	a.Sess = session
	a.S3 = s3.New(session)
	a.processor = processor.New(a.Sess, a.TmpDir, a.Bucket, a.ProcessedPrefix, a.SmallPrefix)

	return a, nil
}

func (a *App) Run() error {
	msgs, err := a.Ch.Consume(
		ChannelUploads, // queue
		"",             // consumer
		false,          // auto-ack
		false,          // exclusive
		false,          // no-local
		false,          // no-wait
		nil,            // args
	)

	if err != nil {
		return err
	}

	log.Printf("Listening on channel %s", ChannelUploads)

	for d := range msgs {
		v, err := video.New(d.Body, a.S3, a.Sess, a.Bucket)
		if err != nil {
			a.logOnError(nil, err)
			d.Ack(false)
			continue
		}

		r := v.GetRequest()
		r.SetOriginalUrl(a.DB)

		err = a.processor.Process(v)
		if err != nil {
			a.logOnError(v, err)
			d.Ack(false)
			continue
		}

		err = r.SetStatus("transcoded", a.DB)
		if err != nil {
			a.logOnError(v, err)
		}

		err = r.SaveDuration(a.DB)
		if err != nil {
			a.logOnError(v, err)
		}

		d.Ack(false)
	}

	return errors.New("Listen queue exited")
}

func (a *App) logOnError(v video.Video, err error) {
	if v == nil {
		log.Printf("Error receiving video: %+v", err.Error())
		return
	}

	r := v.GetRequest()

	log.Printf("Error processing video %s: %+v", r.Id, err.Error())

	err = r.SetStatus("error", a.DB)
	if err != nil {
		log.Printf("Error saving status on video %s: %+v", r.Id, err.Error())
	}
}
