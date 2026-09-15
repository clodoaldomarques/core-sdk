package sqs

import "context"

// Config agrupa as configurações necessárias para o SQS.
type Config interface {
	Region() string
	Address() string // endpoint personalizado (ex: LocalStack)
	AccessKeyID() string
	SecretAccessKey() string
	QueueURL() string
	DeadLetterQueueURL() string
	MaxReceiveCount() int
}

// MessageHandler define a função que processa uma mensagem recebida.
// Retorna um erro se o processamento falhar; caso contrário, a mensagem será deletada.
type MessageHandler func(ctx context.Context, msg *Message) error
