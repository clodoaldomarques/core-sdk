package sns

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/clodoaldomarques/core-sdk/pkg/aws"
	"github.com/clodoaldomarques/core-sdk/pkg/logger"
)

func NewSNSClient(ctx context.Context, c aws.Config) *sns.Client {
	cfg, err := aws.NewCustomConfig(ctx, c)
	if err != nil {
		logger.Fatal(ctx, "fail on loading settings", logger.Fields{
			"error":      err.Error(),
			"AwsRegion":  cfg.Region,
			"AwsAddress": cfg.BaseEndpoint,
		})
	}
	return sns.NewFromConfig(cfg)
}
