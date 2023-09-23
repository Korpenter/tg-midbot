package ydb

import (
	"errors"
	"log"
	"os"

	"github.com/Korpenter/tg-midbot/midbot-in/pkg/agent"
	"github.com/Korpenter/tg-midbot/midbot-in/pkg/models"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

type ydb struct {
	*dynamo.DB
	appTable *dynamo.Table
}

func New() *ydb {
	sess := session.Must(session.NewSession())
	db := dynamo.New(sess,
		&aws.Config{
			Region:   aws.String(os.Getenv("AWS_DEFAULT_REGION")),
			Endpoint: aws.String(os.Getenv("YDB_ENDPOINT")),
		},
	)

	err := db.CreateTable("midbot/Applications", &models.Application{}).Run()
	if err != nil {
		log.Println(err)
	}
	appTable := db.Table("midbot/Applications")
	return &ydb{
		DB:       db,
		appTable: &appTable,
	}
}

func (y *ydb) GetApplication(app *models.Application) (*models.Application, error) {
	var resultApp models.Application

	err := y.appTable.Get("ChatID", app.ChatID).Range("ApplicationID", dynamo.Equal, app.ApplicationID).One(&resultApp)
	if errors.Is(err, dynamo.ErrNotFound) {
		return nil, agent.ErrAppNotFound
	}
	if err != nil {
		log.Printf("Error fetching application: %v", err)
		return nil, err
	}

	return &resultApp, nil
}

func (y *ydb) SaveApplication(app *models.Application) error {
	err := y.appTable.Put(app).Run()
	if err != nil {
		log.Printf("Error saving application: %v", err)
	}
	return err
}

func (y *ydb) RemoveApplication(app *models.Application) error {
	err := y.appTable.Delete("ChatID", app.ChatID).Range("ApplicationID", app.ApplicationID).Run()
	if err != nil {
		log.Printf("Error removing application: %v", err)
	}
	return err
}

func (y *ydb) UpdateApplicationStatus(app *models.Application, status int) error {
	err := y.appTable.Update("ChatID", app.ChatID).
		Range("ApplicationID", app.ApplicationID).
		Set("Status", status).Run()
	if err != nil {
		log.Printf("Error updating application status: %v", err)
	}
	return err
}
