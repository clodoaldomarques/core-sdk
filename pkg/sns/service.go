package sns

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/clodoaldomarques/core-sdk/pkg/aws"
	"github.com/clodoaldomarques/core-sdk/pkg/logger"
)

type Publisher struct {
	ctx      context.Context
	svc      *sns.Client
	topicARN string
}

func NewPublisher(ctx context.Context, c aws.Config, t string) *Publisher {
	return &Publisher{
		ctx:      ctx,
		svc:      NewSNSClient(ctx, c),
		topicARN: t,
	}
}

func (p Publisher) Emit(ctx context.Context, e Event) {
	input := &sns.PublishInput{
		Message:  e.ToMessage(),
		TopicArn: &p.topicARN,
	}

	result, err := p.svc.Publish(ctx, input)
	if err != nil {
		logger.Error(ctx, "error on publish", logger.Fields{
			"error": err.Error(),
		})
	}
	logger.Info(ctx, "event published with success", logger.Fields{
		"message_id": result.MessageId,
		"event":      e,
	})
}
