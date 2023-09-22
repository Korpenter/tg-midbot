package agent

import (
	"fmt"
	"log"

	"github.com/Korpenter/tg-midbot/midbot-in/pkg/helpers"
	"github.com/Korpenter/tg-midbot/midbot-in/pkg/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (a *Agent) handleCheckCommand(message *tgbotapi.Message) error {
	id := message.CommandArguments()
	if id == "" {
		return a.sendToSQS(message.Chat.ID, fmt.Sprintf("Please provide an application number after /track."))
	}
	if !helpers.IsValidID(id) {
		return a.sendToSQS(message.Chat.ID, "Invalid id.")
	}
	application, err := a.requestStatus(id)
	if err != nil {
		return a.sendToSQS(message.Chat.ID, err.Error())
	}

	msg := fmt.Sprintf("Status for application %s:\n- Passport Status: %s\n- Internal Status: %s",
		id, application.PassportStatus.Name, application.InternalStatus.Name)

	return a.sendToSQS(message.Chat.ID, msg)
}

func (a *Agent) handleTrackCommand(message *tgbotapi.Message) error {
	id := message.CommandArguments()
	if id == "" {
		return a.sendToSQS(message.Chat.ID, "Please provide an application number after /track.")
	}
	if !helpers.IsValidID(id) {
		return a.sendToSQS(message.Chat.ID, "Invalid id.")
	}
	status, err := a.requestStatus(id)
	if err != nil {
		return a.sendToSQS(message.Chat.ID, err.Error())
	}
	// if status.PassportStatus.ID == 4 {
	// 	msg := "Application is ready for pickup. You can track it manually now."
	// 	err = a.sendToSQS(message.Chat.ID, msg)
	// 	return err
	// }
	app := &models.Application{
		ChatID:        message.Chat.ID,
		ApplicationID: status.UID,
	}
	log.Printf("Handling track command for application: %v", id)
	err = a.repo.SaveApplication(app)
	switch err {
	case ErrAlreadyExists:
		return a.sendToSQS(message.Chat.ID, fmt.Sprintf("Application %s is already being tracked.", id))
	case nil:
		return a.sendToSQS(message.Chat.ID, fmt.Sprintf("Started tracking application %s.", id))
	default:
		log.Printf("Error tracking application %s: %v", id, err)
		return a.sendToSQS(message.Chat.ID, err.Error())
	}
}

func (a *Agent) handleUntrackCommand(message *tgbotapi.Message) error {
	id := message.CommandArguments()
	if id == "" {
		return a.sendToSQS(message.Chat.ID, fmt.Sprintf("Please provide an application number after /track."))
	}
	if !helpers.IsValidID(id) {
		return a.sendToSQS(message.Chat.ID, fmt.Sprintf("Invalid id."))
	}
	app := &models.Application{
		ChatID:        message.Chat.ID,
		ApplicationID: id,
	}
	log.Printf("Handling untrack command for application: %v", id)
	err := a.repo.RemoveApplication(app)
	switch err {
	case ErrAppNotFound:
		return a.sendToSQS(message.Chat.ID, fmt.Sprintf("Application %s is not being tracked.", id))
	case nil:
		return a.sendToSQS(message.Chat.ID, fmt.Sprintf("Stopped tracking application %s.", id))
	default:
		log.Printf("Error untracking application %s: %v", id, err)
		return a.sendToSQS(message.Chat.ID, err.Error())
	}
}

func (a *Agent) handleNotifyCommand() error {
	apps, err := a.repo.GetAllApplications()
	if err != nil {
		return err
	}
	var msg string
	for _, app := range apps {
		status, err := a.requestStatus(app.ApplicationID)
		if err != nil {
			continue
		}
		if status.PassportStatus.ID != app.Status {
			if status.PassportStatus.ID == 5 { // should be 4
				msg = fmt.Sprintf("Application %s is processed with status: %s.\n Stopped tracking.", app.ApplicationID, status.PassportStatus.Name)
				a.repo.RemoveApplication(&app)
			} else {
				msg = fmt.Sprintf("Application %s is updated with status: %s", app.ApplicationID, status.PassportStatus.Name)
				a.repo.UpdateApplicationStatus(&app, status.PassportStatus.ID)
			}
			a.sendToSQS(app.ChatID, msg)
		}
	}
	return nil
}
