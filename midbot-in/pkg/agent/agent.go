package agent

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

	"github.com/Korpenter/tg-midbot/midbot-in/pkg/dto"
	"github.com/Korpenter/tg-midbot/midbot-in/pkg/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/labstack/echo/v4"
)

var (
	ErrAppNotFound   = errors.New("Application not found.")
	ErrForbidden     = errors.New("Forbidden access. Please try again later.")
	ErrAlreadyExists = errors.New("Application is already being tracked.")
)

var userAgents = []string{
	"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
}

type Storage interface {
	GetApplication(*models.Application) (*models.Application, error)
	SaveApplication(*models.Application) error
	RemoveApplication(*models.Application) error
	UpdateApplicationStatus(app *models.Application, status int) error
	SaveCheckpoint(appID string) error
	GetCheckpoint() (string, error)
	DeleteCheckpoint() error

	GetAllApplicationsBatched(startId string, BatchSize int64) ([]models.Application, error)
}

type Agent struct {
	Bot       *tgbotapi.BotAPI
	sqsClient *sqs.Client
	repo      Storage
	midClient *http.Client
}

func New(bot *tgbotapi.BotAPI, sqsClient *sqs.Client, repo Storage, midClient *http.Client) *Agent {
	return &Agent{
		Bot:       bot,
		sqsClient: sqsClient,
		repo:      repo,
		midClient: midClient,
	}
}

func (a *Agent) HandleUpdate(c echo.Context) error {
	var sqsEvent dto.QueueEvent
	var err error
	if err := c.Bind(&sqsEvent); err != nil {
		log.Printf("Cannot bind message queue event: %v\n", err)
		return c.JSON(500, nil)
	}
	var update tgbotapi.Update
	body := sqsEvent.Messages[0].Details.Message.Body
	if len(body) > 0 {
		err = json.Unmarshal([]byte(body), &update)
		if err != nil {
			log.Printf("Failed to unmarshal body into update: %v\n", err)
			return c.JSON(500, nil)
		}
		if update.Message != nil && update.Message.IsCommand() {
			log.Println(update.Message, update.Message.Command())
			switch update.Message.Command() {
			case "check":
				return a.handleCheckCommand(update.Message)
			case "track":
				return a.handleTrackCommand(update.Message)
			case "untrack":
				return a.handleUntrackCommand(update.Message)
			default:
				return a.sendToSQS(update.Message.Chat.ID, fmt.Sprintf("Unknown command."))
			}
		}
	} else {
		payload := sqsEvent.Messages[0].Details.Payload
		switch {
		case payload == "/notify":
			return a.handleNotifyCommand()
		default:
			return nil
		}
	}
	return nil
}

func (a *Agent) sendToSQS(chatID int64, msg string) error {
	payload := dto.SQSMessagePayload{
		ChatID:      chatID,
		MessageBody: msg,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	queueUrl := os.Getenv("QUEUE_URL")
	_, err = a.sqsClient.SendMessage(context.Background(), &sqs.SendMessageInput{
		QueueUrl:    &queueUrl,
		MessageBody: aws.String(string(jsonPayload)),
	})
	return err
}

func (a *Agent) requestStatus(id string) (*models.Status, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://info.midpass.ru/api/request/%s", strings.TrimSpace(id)), nil)
	if err != nil {
		return nil, errors.New("Encountered an error, try again later.")
	}

	randomIndex := rand.Intn(len(userAgents))
	req.Header.Set("User-Agent", userAgents[randomIndex])
	req.Header.Set("Accept", "application/json")

	resp, err := a.midClient.Do(req)
	if err != nil {
		return nil, errors.New("Error fetching application status. Please try again later.")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case 400:
			return nil, errors.New("Application not found.")
		case 403:
			return nil, errors.New("Forbidden access. Please try again later.")
		default:
			return nil, fmt.Errorf("received unexpected status: %v", resp.Status)
		}
	}

	var application models.Status
	if err := json.NewDecoder(resp.Body).Decode(&application); err != nil {
		return nil, errors.New("Error parsing application status. Please try again later.")
	}

	return &application, nil
}
