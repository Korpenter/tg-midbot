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
