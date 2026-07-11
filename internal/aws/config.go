package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func NewCustomConfig(ctx context.Context, c Config) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(c.Region()),
		config.WithBaseEndpoint(c.Address()),
		config.WithCredentialsProvider(NewCustomCredentials(c.AccessKeyID(), c.SecretAccessKey())),
	)

	if err != nil {
		return aws.Config{}, fmt.Errorf("fail on loading AWS configurations: %w", err)
	}

	return cfg, nil
}

func NewCustomCredentials(accessKeyID, secretAccessKey string) aws.CredentialsProvider {
	return aws.NewCredentialsCache(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		creds := aws.Credentials{
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
			Source:          "Environment",
		}
		if creds.AccessKeyID == "" || creds.SecretAccessKey == "" {
			return aws.Credentials{}, fmt.Errorf("credenciais AWS ausentes nas variáveis de ambiente")
		}
		return creds, nil
	}))
}

func ParseAWSString(s string) *string {
	return aws.String(s)
}
