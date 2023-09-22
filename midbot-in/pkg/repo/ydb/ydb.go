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
		log.Print(err)
	}

	return &ydb{
		db,
	}
}

func (y *ydb) GetApplication(app *models.Application) (*models.Application, error) {
	var resultApp models.Application

	err := y.DB.Table("midbot/Applications").Get("ChatID", app.ChatID).Range("ApplicationID", dynamo.Equal, app.ApplicationID).One(&resultApp)
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
	existingApp, err := y.GetApplication(app)
	if err != nil && !errors.Is(err, agent.ErrAppNotFound) {
		return err
	}
	if existingApp != nil {
		return agent.ErrAlreadyExists
	}

	err = y.DB.Table("midbot/Applications").Put(app).Run()
	if err != nil {
		log.Printf("Error saving application: %v", err)
	}
	return err
}

func (y *ydb) RemoveApplication(app *models.Application) error {
	log.Printf("Attempting to remove application with ChatID: %v and ApplicationID: %v", app.ChatID, app.ApplicationID)

	_, err := y.GetApplication(app)
	if errors.Is(err, dynamo.ErrNotFound) {
		return agent.ErrAppNotFound
	}
	if err != nil {
		return err
	}

	err = y.DB.Table("midbot/Applications").Delete("ChatID", app.ChatID).Range("ApplicationID", app.ApplicationID).Run()
	if err != nil {
		log.Printf("Error removing application: %v", err)
	}
	return err
}

func (y *ydb) GetAllApplications() ([]models.Application, error) {
	var apps []models.Application
	err := y.DB.Table("midbot/Applications").Scan().All(&apps)
	if err != nil {
		log.Printf("Error fetching all applications: %v", err)
		return nil, err
	}
	return apps, nil
}

func (y *ydb) UpdateApplicationStatus(app *models.Application, status int) error {
	err := y.DB.Table("midbot/Applications").Update("ChatID", app.ChatID).
		Range("ApplicationID", app.ApplicationID).
		Set("Status", status).Run()
	if err != nil {
		log.Printf("Error updating application status: %v", err)
	}
	return err
}
