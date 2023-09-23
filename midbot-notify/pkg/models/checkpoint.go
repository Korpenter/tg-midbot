package models

type Checkpoint struct {
	Identifier    string `dynamo:"Identifier,hash"`
	ApplicationID string `dynamo:"ApplicationID"`
}
