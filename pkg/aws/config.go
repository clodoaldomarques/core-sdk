package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

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

func NewCustomConfig(ctx context.Context, region, address, accessKeyID, secretAccessKey string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithBaseEndpoint(address),
		config.WithCredentialsProvider(NewCustomCredentials(accessKeyID, secretAccessKey)),
	)

	if err != nil {
		return aws.Config{}, fmt.Errorf("falha ao carregar configuração AWS: %w", err)
	}

	return cfg, nil
}
