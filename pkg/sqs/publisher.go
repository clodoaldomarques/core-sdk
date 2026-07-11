package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/clodoaldomarques/core-sdk/pkg/logger"
)

type Publisher struct {
	ctx      context.Context
	svc      *sqs.Client
	queueURL string
}

// NewPublisher cria um novo Publisher para a fila configurada.
func NewPublisher(ctx context.Context, c Config) *Publisher {
	return &Publisher{
		ctx:      ctx,
		svc:      NewSQSClient(ctx, c),
		queueURL: c.QueueURL(),
	}
}

// Send envia uma única mensagem (string ou Event).
// Para enviar um Event, use event.ToMessage() (como no SNS) ou serialize manualmente.
func (p *Publisher) Send(ctx context.Context, body string, attrs map[string]types.MessageAttributeValue) (string, error) {
	input := &sqs.SendMessageInput{
		QueueUrl:          &p.queueURL,
		MessageBody:       aws.String(body),
		MessageAttributes: attrs,
	}
	result, err := p.svc.SendMessage(ctx, input)
	if err != nil {
		return "", err
	}
	logger.Info(ctx, "message sent to SQS", logger.Fields{
		"message_id": *result.MessageId,
		"queue_url":  p.queueURL,
	})
	return *result.MessageId, nil
}

// SendBatch envia um lote de mensagens (máximo 10 por chamada).
func (p *Publisher) SendBatch(ctx context.Context, entries []types.SendMessageBatchRequestEntry) ([]types.SendMessageBatchResultEntry, error) {
	if len(entries) == 0 {
		return nil, nil
	}
	if len(entries) > 10 {
		// A SDK aceita no máximo 10; divida em lotes ou retorne erro.
		// Aqui truncamos para 10 (ou você pode implementar loop).
		entries = entries[:10]
	}
	input := &sqs.SendMessageBatchInput{
		QueueUrl: &p.queueURL,
		Entries:  entries,
	}
	result, err := p.svc.SendMessageBatch(ctx, input)
	if err != nil {
		return nil, err
	}
	logger.Info(ctx, "batch messages sent", logger.Fields{
		"successful": len(result.Successful),
		"failed":     len(result.Failed),
	})
	return result.Successful, nil
}
