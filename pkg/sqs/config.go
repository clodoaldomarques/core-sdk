package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/clodoaldomarques/core-sdk/pkg/aws"
	"github.com/clodoaldomarques/core-sdk/pkg/logger"
)

func NewSQSClient(ctx context.Context, c aws.Config) *sqs.Client {
	cfg, err := aws.NewCustomConfig(ctx, c)
	if err != nil {
		logger.Fatal(ctx, "fail on loading settings", logger.Fields{
			"error":      err.Error(),
			"AwsRegion":  cfg.Region,
			"AwsAddress": cfg.BaseEndpoint,
		})
	}
	return sqs.NewFromConfig(cfg)
}
