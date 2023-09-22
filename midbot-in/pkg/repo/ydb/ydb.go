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
	checkTable *dynamo.Table
	appTable   *dynamo.Table
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
	err = db.CreateTable("midbot/Checkpoints", &models.Checkpoint{}).Run()
	if err != nil {
		log.Println(err)
	}
	checkTable := db.Table("midbot/Checkpoints")

	return &ydb{
		DB:         db,
		checkTable: &checkTable,
		appTable:   &appTable,
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

func (y *ydb) GetAllApplicationsBatched(startId string, batchSize int64) ([]models.Application, error) {
	var apps []models.Application

	scanOp := y.appTable.Scan().Limit(batchSize)

	if startId != "" {
		scanOp = scanOp.Filter("ApplicationID >= ?", startId)
	}

	err := scanOp.All(&apps)
	if err != nil {
		log.Printf("Error fetching applications batch: %v", err)
		return nil, err
	}

	return apps, nil
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

func (y *ydb) SaveCheckpoint(appID string) error {
	checkpoint := models.Checkpoint{
		Identifier:    "CHECKPOINT",
		ApplicationID: appID,
	}

	return y.checkTable.Put(checkpoint).Run()
}

func (y *ydb) GetCheckpoint() (string, error) {
	var cp models.Checkpoint
	err := y.checkTable.Get("Identifier", "CHECKPOINT").One(&cp)
	if err != nil {
		if errors.Is(err, dynamo.ErrNotFound) {
			return "", nil
		}
		return "", err
	}

	return cp.ApplicationID, nil
}

func (y *ydb) DeleteCheckpoint() error {
	return y.checkTable.Delete("Identifier", "CHECKPOINT").Run()
}
