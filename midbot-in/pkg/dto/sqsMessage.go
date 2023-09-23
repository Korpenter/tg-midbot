package dto

type SQSMessagePayload struct {
	ChatID      int64  `json:"chat_id"`
	MessageBody string `json:"message_body"`
}
