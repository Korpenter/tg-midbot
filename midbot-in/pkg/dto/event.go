package dto

type QueueEvent struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	EventMetadata EventMetadata `json:"event_metadata"`
	Details       Details       `json:"details"`
}

type EventMetadata struct {
	EventID        string   `json:"event_id"`
	EventType      string   `json:"event_type"`
	CreatedAt      string   `json:"created_at"`
	CloudID        string   `json:"cloud_id"`
	FolderID       string   `json:"folder_id"`
	TracingContext struct{} `json:"tracing_context"`
}

type Details struct {
	QueueID string         `json:"queue_id"`
	Message MessageDetails `json:"message"`
	Payload string         `json:"payload"`
}

type MessageDetails struct {
	MessageID              string               `json:"message_id"`
	MD5OfBody              string               `json:"md5_of_body"`
	Body                   string               `json:"body"`
	Attributes             map[string]string    `json:"attributes"`
	MessageAttributes      map[string]Attribute `json:"message_attributes"`
	MD5OfMessageAttributes string               `json:"md5_of_message_attributes"`
}

type Attribute struct {
	DataType    string `json:"dataType"`
	StringValue string `json:"stringValue"`
}
