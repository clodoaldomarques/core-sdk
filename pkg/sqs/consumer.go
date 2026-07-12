package sqs

import (
	"context"
	"strconv"
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

// WithDLQ configura a Dead Letter Queue.
// maxReceives: número máximo de tentativas antes de mover para DLQ.
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
		dlqURL:          c.DeadLetterQueueURL(), // Pode vir da config ou via opção
		handler:         handler,
		maxMessages:     10,
		maxReceiveCount: c.MaxReceiveCount(), // Pode vir da config
		waitTime:        20,
		visibility:      30,
		workerCount:     5,
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
		AttributeNames:        []types.QueueAttributeName{"All"}, // Inclui ApproximateReceiveCount
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

	// Processa cada mensagem
	for _, msg := range result.Messages {
		c.processMessage(ctx, msg, workerID)
	}
}

// processMessage lida com o ciclo de vida de uma única mensagem: handler, DLQ, deleção.
func (c *Consumer) processMessage(ctx context.Context, msg types.Message, workerID int) {
	ourMsg := NewMessageFromAWS(msg)

	// Verifica se a mensagem já atingiu o limite de recebimentos (para DLQ)
	receiveCount := c.getReceiveCount(msg)
	if c.shouldMoveToDLQ(receiveCount) {
		c.moveToDLQ(ctx, msg, receiveCount, workerID)
		return // Não executa o handler, pois já está descartando/movendo
	}

	// Executa o handler do usuário
	err := c.handler(ctx, ourMsg)
	if err != nil {
		logger.Error(ctx, "handler failed for message", logger.Fields{
			"message_id":    ourMsg.MessageID,
			"error":         err.Error(),
			"worker_id":     workerID,
			"receive_count": receiveCount,
		})
		// Não deleta nem move para DLQ agora; a mensagem retornará após visibility timeout.
		// A próxima tentativa incrementará o contador.
		return
	}

	// Sucesso: deleta a mensagem
	err = c.deleteMessage(ctx, msg.ReceiptHandle)
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

// shouldMoveToDLQ decide se a mensagem deve ser movida para a DLQ.
func (c *Consumer) shouldMoveToDLQ(receiveCount int) bool {
	if c.dlqURL == "" || c.maxReceiveCount <= 0 {
		return false // DLQ não configurada
	}
	return receiveCount >= c.maxReceiveCount
}

// getReceiveCount extrai o ApproximateReceiveCount do atributo da mensagem.
func (c *Consumer) getReceiveCount(msg types.Message) int {
	if msg.Attributes == nil {
		return 0
	}
	countStr, ok := msg.Attributes["ApproximateReceiveCount"]
	if !ok {
		return 0
	}
	count, err := strconv.Atoi(countStr)
	if err != nil {
		logger.Warn(context.Background(), "invalid ApproximateReceiveCount", logger.Fields{
			"value": countStr,
			"error": err.Error(),
		})
		return 0
	}
	return count
}

// moveToDLQ envia a mensagem para a fila DLQ e a deleta da fila original.
func (c *Consumer) moveToDLQ(ctx context.Context, msg types.Message, receiveCount int, workerID int) {
	ourMsg := NewMessageFromAWS(msg)
	logger.Warn(ctx, "moving message to DLQ", logger.Fields{
		"message_id":    ourMsg.MessageID,
		"receive_count": receiveCount,
		"worker_id":     workerID,
		"dlq_url":       c.dlqURL,
	})

	// Envia para a DLQ (preserva corpo e atributos)
	_, err := c.svc.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:          &c.dlqURL,
		MessageBody:       msg.Body,
		MessageAttributes: msg.MessageAttributes,
		// Opcional: adicionar atributo para rastrear motivo
	})
	if err != nil {
		logger.Error(ctx, "failed to send message to DLQ", logger.Fields{
			"message_id": ourMsg.MessageID,
			"error":      err.Error(),
			"worker_id":  workerID,
		})
		// Não deletamos a mensagem original para não perdê-la; ela retornará
		// e será tentada novamente (mas pode causar loop infinito se o erro persistir).
		// Uma alternativa é abortar e deixar o operador resolver.
		return
	}

	// Deleta da fila original
	err = c.deleteMessage(ctx, msg.ReceiptHandle)
	if err != nil {
		logger.Error(ctx, "failed to delete original message after DLQ send", logger.Fields{
			"message_id": ourMsg.MessageID,
			"error":      err.Error(),
			"worker_id":  workerID,
		})
		// Aqui temos um problema: a mensagem foi enviada para a DLQ, mas não deletada da original.
		// Isso pode causar duplicação. Uma abordagem é tentar deletar novamente ou registrar para
		// monitoramento manual.
	} else {
		logger.Info(ctx, "message moved to DLQ and deleted from original queue", logger.Fields{
			"message_id": ourMsg.MessageID,
			"worker_id":  workerID,
		})
	}
}

// deleteMessage é um helper para deletar uma mensagem pelo receipt handle.
func (c *Consumer) deleteMessage(ctx context.Context, receiptHandle *string) error {
	_, err := c.svc.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &c.queueURL,
		ReceiptHandle: receiptHandle,
	})
	return err
}
