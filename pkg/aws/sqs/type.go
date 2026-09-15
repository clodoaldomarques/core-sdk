package sqs

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
)

// Message representa uma mensagem da fila SQS.
type Message struct {
	MessageID         string                                 `json:"message_id"`
	ReceiptHandle     string                                 `json:"receipt_handle"`
	Body              string                                 `json:"body"`
	Attributes        map[string]string                      `json:"attributes"`         // atributos do sistema (ex: SentTimestamp)
	MessageAttributes map[string]types.MessageAttributeValue `json:"message_attributes"` // atributos personalizados
}

// NewMessageFromAWS converte uma mensagem da SDK para o nosso tipo.
func NewMessageFromAWS(msg types.Message) *Message {
	return &Message{
		MessageID:         *msg.MessageId,
		ReceiptHandle:     *msg.ReceiptHandle,
		Body:              *msg.Body,
		Attributes:        msg.Attributes,
		MessageAttributes: msg.MessageAttributes,
	}
}

// Event (opcional) – se quiser manter o mesmo formato de evento do SNS,
// pode ter um método para converter o Body em Event.
// Exemplo:
type Event struct {
	EventID   uuid.UUID `json:"event_id"`
	EventType string    `json:"event_type"`
	EventData any       `json:"event_data"`
	EventDate time.Time `json:"event_date"`
}

// ToEvent desserializa o Body da mensagem para um Event.
func (m *Message) ToEvent() (*Event, error) {
	var evt Event
	if err := json.Unmarshal([]byte(m.Body), &evt); err != nil {
		return nil, err
	}
	return &evt, nil
}
