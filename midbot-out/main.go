package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/Korpenter/tg-midbot/midbot-out/pkg/dto"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var bot *tgbotapi.BotAPI

func init() {
	var err error
	bot, err = tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		panic(err)
	}
}

func Handler(ctx context.Context, event dto.Event) error {
	message := event.Messages[0]
	var payload dto.SQSMessagePayload
	err := json.Unmarshal([]byte(message.Details.Message.Body), &payload)
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(payload.ChatID, payload.MessageBody)
	_, err = bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}
