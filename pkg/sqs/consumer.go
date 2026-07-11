package sqs

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/clodoaldomarques/core-sdk/pkg/logger"
)

// Consumer gerencia o consumo de mensagens da fila.
type Consumer struct {
	ctx      context.Context
	svc      *sqs.Client
	queueURL string
	handler  MessageHandler
	dlqURL   string

	// Opções de consumo
	maxReceiveCount int
	maxMessages     int32
	waitTime        int32
	visibility      int32
	workerCount     int
}

// ConsumerOption define funções de configuração do Consumer.
type ConsumerOption func(*Consumer)

// WithMaxMessages define o número máximo de mensagens por ReceiveMessage.
func WithMaxMessages(n int32) ConsumerOption {
	return func(c *Consumer) { c.maxMessages = n }
}

// WithWaitTime define o tempo de long polling (0-20s).
func WithWaitTime(seconds int32) ConsumerOption {
	return func(c *Consumer) { c.waitTime = seconds }
}

// WithVisibilityTimeout define o tempo de invisibilidade (em segundos).
func WithVisibilityTimeout(seconds int32) ConsumerOption {
	return func(c *Consumer) { c.visibility = seconds }
}

// WithWorkers define o número de workers concorrentes (goroutines).
func WithWorkers(n int) ConsumerOption {
	return func(c *Consumer) { c.workerCount = n }
}

func WithDLQ(url string, maxReceives int) ConsumerOption {
	return func(c *Consumer) {
		c.dlqURL = url
		c.maxReceiveCount = maxReceives
	}
}

// NewConsumer cria um novo Consumer.
func NewConsumer(ctx context.Context, c Config, handler MessageHandler, opts ...ConsumerOption) *Consumer {
	consumer := &Consumer{
		ctx:             ctx,
		svc:             NewSQSClient(ctx, c),
		queueURL:        c.QueueURL(),
		dlqURL:          c.DeadLetterQueueURL(),
		handler:         handler,
		maxMessages:     10, // padrão
		maxReceiveCount: c.MaxReceiveCount(),
		waitTime:        20, // long polling ativado por padrão
		visibility:      30, // 30 segundos
		workerCount:     5,  // 5 workers
	}
	for _, opt := range opts {
		opt(consumer)
	}
	return consumer
}

// Start inicia o consumo contínuo (bloqueia até o contexto ser cancelado).
func (c *Consumer) Start() error {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(c.ctx)
	defer cancel()

	for i := 0; i < c.workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			logger.Info(ctx, "SQS consumer worker started", logger.Fields{"worker_id": workerID})
			for {
				select {
				case <-ctx.Done():
					logger.Info(ctx, "SQS consumer worker stopping", logger.Fields{"worker_id": workerID})
					return
				default:
					c.pollAndProcess(ctx, workerID)
				}
			}
		}(i)
	}

	wg.Wait()
	return nil
}

// pollAndProcess faz uma única chamada ReceiveMessage e processa as mensagens.
func (c *Consumer) pollAndProcess(ctx context.Context, workerID int) {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:              &c.queueURL,
		MaxNumberOfMessages:   c.maxMessages,
		WaitTimeSeconds:       c.waitTime,
		VisibilityTimeout:     c.visibility,
		MessageAttributeNames: []string{"All"},
		AttributeNames:        []types.QueueAttributeName{"All"},
	}

	result, err := c.svc.ReceiveMessage(ctx, input)
	if err != nil {
		logger.Error(ctx, "error receiving messages", logger.Fields{
			"error":     err.Error(),
			"worker_id": workerID,
		})
		return
	}

	if len(result.Messages) == 0 {
		return
	}

	for _, msg := range result.Messages {
		// Processa a mensagem
		ourMsg := NewMessageFromAWS(msg)
		err := c.handler(ctx, ourMsg)
		if err != nil {
			// Se o handler retornar erro, NÃO deletamos a mensagem.
			// Ela voltará a ficar visível após o visibility timeout.
			logger.Error(ctx, "handler failed for message", logger.Fields{
				"message_id": ourMsg.MessageID,
				"error":      err.Error(),
				"worker_id":  workerID,
			})
			// Opcional: enviar para DLQ ou incrementar contador de recebimentos.
			continue
		}

		// Sucesso: deleta a mensagem
		_, err = c.svc.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      &c.queueURL,
			ReceiptHandle: msg.ReceiptHandle,
		})
		if err != nil {
			logger.Error(ctx, "failed to delete message", logger.Fields{
				"message_id": ourMsg.MessageID,
				"error":      err.Error(),
				"worker_id":  workerID,
			})
		} else {
			logger.Info(ctx, "message processed and deleted", logger.Fields{
				"message_id": ourMsg.MessageID,
				"worker_id":  workerID,
			})
		}
	}
}

// DeleteMessage é um método auxiliar para deletar uma mensagem individualmente (caso queira usar fora do fluxo).
func (c *Consumer) DeleteMessage(ctx context.Context, receiptHandle string) error {
	_, err := c.svc.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &c.queueURL,
		ReceiptHandle: &receiptHandle,
	})
	return err
}
