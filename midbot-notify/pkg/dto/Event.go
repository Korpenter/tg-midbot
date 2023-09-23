package dto

type Event struct {
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
	QueueID string `json:"queue_id"`
	Payload string `json:"payload"`
}
