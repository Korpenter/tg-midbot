package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Korpenter/tg-midbot/midbot-in/pkg/agent"
	"github.com/Korpenter/tg-midbot/midbot-in/pkg/repo/ydb"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

var Version string

func main() {
	log.Println("Release info: ", fmt.Sprintf("Version: %s ", Version))
	godotenv.Load()

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		panic(err)
	}

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
	r := ydb.New()
	c := sqs.NewFromConfig(cfg)

	a := agent.New(bot, c, r, http.DefaultClient)
	e := echo.New()
	e.POST("/", a.HandleUpdate)
	e.Start(":" + os.Getenv("PORT"))
}
