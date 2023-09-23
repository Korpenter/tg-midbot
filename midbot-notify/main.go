package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/Korpenter/tg-midbot/midbot-notify/pkg/dto"
	"github.com/Korpenter/tg-midbot/midbot-notify/pkg/models"
	"github.com/Korpenter/tg-midbot/midbot-notify/pkg/repo/ydb"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

const BatchSize = 100

var userAgents = []string{
	"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
}

type Storage interface {
	RemoveApplication(app *models.Application) error
	UpdateApplicationStatus(app *models.Application, status int) error
	SaveCheckpoint(appID string) error
	GetCheckpoint() (string, error)
	DeleteCheckpoint() error
	GetAllApplicationsBatched(startId string, BatchSize int64) ([]models.Application, error)
}

var sqsClient *sqs.Client
var repo Storage
var midClient *http.Client

func Handler(ctx context.Context, event dto.Event) error {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           "https://message-queue.api.cloud.yandex.net",
			SigningRegion: "ru-central1",
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return err
	}
	repo, err = ydb.New()
	if err != nil {
		return err
	}
	sqsClient = sqs.NewFromConfig(cfg)
	midClient = http.DefaultClient

	lastCheckpoint, err := repo.GetCheckpoint()
	if err != nil {
		log.Printf("Error fetching checkpoint: %v", err)
		return err
	}

	for {
		apps, err := repo.GetAllApplicationsBatched(lastCheckpoint, BatchSize)
		if err != nil {
			log.Printf("Error fetching applications in batch from checkpoint %s: %v", lastCheckpoint, err)
			return err
		}
		if len(apps) == 0 {
			break
		}

		for _, app := range apps {
			status, err := requestStatus(app.ApplicationID)
			if err != nil {
				sendToSQS(app.ChatID, "Error fetching status")
				continue
			}

			var msg string
			if status.PassportStatus.ID != app.Status {
				if status.PassportStatus.ID == 4 {
					msg = fmt.Sprintf("Application %s is processed with status: %s.\nStopped tracking.", app.ApplicationID, status.PassportStatus.Name)
					repo.RemoveApplication(&app)
				} else {
					msg = fmt.Sprintf("Application %s is updated with status: %s", app.ApplicationID, status.PassportStatus.Name)
				}
				sendToSQS(app.ChatID, msg)
			}

			if err := repo.SaveCheckpoint(app.ApplicationID); err != nil {
				log.Printf("Error saving checkpoint: %v", err)
				return err
			}
		}

		if len(apps) < BatchSize {
			if err := repo.DeleteCheckpoint(); err != nil {
				log.Printf("Error deleting checkpoint: %v", err)
				return err
			}
			break
		}

		lastCheckpoint = apps[len(apps)-1].ApplicationID
	}

	return nil
}

func sendToSQS(chatID int64, msg string) error {
	payload := dto.SQSMessagePayload{
		ChatID:      chatID,
		MessageBody: msg,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	queueUrl := os.Getenv("QUEUE_URL")
	_, err = sqsClient.SendMessage(context.Background(), &sqs.SendMessageInput{
		QueueUrl:    &queueUrl,
		MessageBody: aws.String(string(jsonPayload)),
	})
	return err
}

func requestStatus(id string) (*models.Status, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://info.midpass.ru/api/request/%s", strings.TrimSpace(id)), nil)
	if err != nil {
		return nil, errors.New("Remote server error.")
	}

	randomIndex := rand.Intn(len(userAgents))
	req.Header.Set("User-Agent", userAgents[randomIndex])
	req.Header.Set("Accept", "application/json")

	resp, err := midClient.Do(req)
	if err != nil {
		return nil, errors.New("Error fetching application status.")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		default:
			return nil, fmt.Errorf("received unexpected status: %v", resp.Status)
		}
	}

	var application models.Status
	if err := json.NewDecoder(resp.Body).Decode(&application); err != nil {
		return nil, errors.New("Error parsing application status.")
	}

	return &application, nil
}
