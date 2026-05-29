package sns

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Type string

type Event struct {
	EventID   uuid.UUID `json:"event_id"`
	EventType string    `json:"event_type`
	EventData any       `json:"event_data"`
	EventDate time.Time `json:"event_date"`
}

func (e Event) ToMessage() *string {
	evt, _ := json.Marshal(e)
	msg := string(evt)
	return &msg
}
