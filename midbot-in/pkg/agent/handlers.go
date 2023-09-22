package agent

import (
	"errors"
	"fmt"
	"log"

	"github.com/Korpenter/tg-midbot/midbot-in/pkg/helpers"
	"github.com/Korpenter/tg-midbot/midbot-in/pkg/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	MsgProvideAppNumber           = "Please provide an application number after /track."
	MsgInvalidID                  = "Invalid id."
	MsgAlreadyTracked             = "Application is already being tracked."
	BatchSize                     = 100
	MsgProvideAppNumberAfterCheck = "Please provide an application number after /check."
)

func (a *Agent) handleCheckCommand(message *tgbotapi.Message) error {
	id := message.CommandArguments()
	chatID := message.Chat.ID

	if id == "" {
		return a.sendToSQS(chatID, MsgProvideAppNumberAfterCheck)
	}
	if !helpers.IsValidID(id) {
		return a.sendToSQS(chatID, MsgInvalidID)
	}

	application, err := a.requestStatus(id)
	if err != nil {
		log.Printf("Error fetching status for application %s: %v", id, err)
		return a.sendToSQS(chatID, err.Error())
	}

	msg := fmt.Sprintf("Status for application %s:\n- Passport Status: %s\n- Internal Status: %s",
		id, application.PassportStatus.Name, application.InternalStatus.Name)

	return a.sendToSQS(chatID, msg)
}

func (a *Agent) handleTrackCommand(message *tgbotapi.Message) error {
	id := message.CommandArguments()
	chatID := message.Chat.ID

	if id == "" {
		return a.sendToSQS(chatID, MsgProvideAppNumber)
	}
	if !helpers.IsValidID(id) {
		return a.sendToSQS(chatID, MsgInvalidID)
	}

	status, err := a.requestStatus(id)
	if err != nil {
		return a.sendToSQS(chatID, err.Error())
	}

	newApp := &models.Application{
		ChatID:        chatID,
		ApplicationID: status.UID,
	}
	app, err := a.repo.GetApplication(newApp)
	if err != nil && !errors.Is(err, ErrAppNotFound) {
		return err
	}
	if app != nil {
		return a.sendToSQS(chatID, MsgAlreadyTracked)
	}

	err = a.repo.SaveApplication(newApp)
	if err != nil {
		log.Printf("Error tracking application %s: %v", id, err)
		return a.sendToSQS(chatID, err.Error())
	}

	return a.sendToSQS(chatID, fmt.Sprintf("Started tracking application %s.", id))
}

func (a *Agent) handleUntrackCommand(message *tgbotapi.Message) error {
	id := message.CommandArguments()
	chatID := message.Chat.ID

	if id == "" {
		return a.sendToSQS(chatID, MsgProvideAppNumber)
	}
	if !helpers.IsValidID(id) {
		return a.sendToSQS(chatID, MsgInvalidID)
	}

	app := &models.Application{
		ChatID:        chatID,
		ApplicationID: id,
	}
	log.Printf("Handling untrack command for application: %s", id)

	_, err := a.repo.GetApplication(app)
	if err != nil {
		if errors.Is(err, ErrAppNotFound) {
			return a.sendToSQS(chatID, fmt.Sprintf("Application %s is not being tracked.", id))
		}
		log.Printf("Error fetching application %s: %v", id, err)
		return err
	}

	err = a.repo.RemoveApplication(app)
	if err != nil {
		log.Printf("Error untracking application %s: %v", id, err)
		return a.sendToSQS(chatID, err.Error())
	}
	return a.sendToSQS(chatID, fmt.Sprintf("Stopped tracking application %s.", id))
}

func (a *Agent) handleNotifyCommand() error {
	lastCheckpoint, err := a.repo.GetCheckpoint()
	if err != nil {
		log.Printf("Error fetching checkpoint: %v", err)
		return err
	}

	for {
		apps, err := a.repo.GetAllApplicationsBatched(lastCheckpoint, BatchSize)
		if err != nil {
			log.Printf("Error fetching applications in batch from checkpoint %s: %v", lastCheckpoint, err)
			return err
		}
		if len(apps) == 0 {
			break
		}

		for _, app := range apps {
			status, err := a.requestStatus(app.ApplicationID)
			if err != nil {
				a.sendToSQS(app.ChatID, "Error fetching status")
				continue
			}

			var msg string
			if status.PassportStatus.ID != app.Status {
				if status.PassportStatus.ID == 5 { // should be 4 for processed (currently for testing purposes)
					msg = fmt.Sprintf("Application %s is processed with status: %s.\nStopped tracking.", app.ApplicationID, status.PassportStatus.Name)
					a.repo.RemoveApplication(&app)
				} else {
					msg = fmt.Sprintf("Application %s is updated with status: %s", app.ApplicationID, status.PassportStatus.Name)
				}
				a.sendToSQS(app.ChatID, msg)
			}

			if err := a.repo.SaveCheckpoint(app.ApplicationID); err != nil {
				log.Printf("Error saving checkpoint: %v", err)
				return err
			}
		}

		if len(apps) < BatchSize {
			if err := a.repo.DeleteCheckpoint(); err != nil {
				log.Printf("Error deleting checkpoint: %v", err)
				return err
			}
			break
		}

		lastCheckpoint = apps[len(apps)-1].ApplicationID
	}
	return nil
}
