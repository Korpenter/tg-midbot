package models

type Application struct {
	ChatID        int64  `dynamo:"ChatID,hash"`
	ApplicationID string `dynamo:"ApplicationID,range"`
	Status        int    `dynamo:"Status,"`
}
