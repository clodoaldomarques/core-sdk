package localstack

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type CredentialsProvider struct {
	AccessKeyID     string
	SecretAccessKey string
}

func (p CredentialsProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return aws.Credentials{
		AccessKeyID:     p.AccessKeyID,
		SecretAccessKey: p.SecretAccessKey,
		SessionToken:    "",
		Source:          "",
		CanExpire:       false,
		Expires:         time.Now().Add(time.Hour * 24),
		AccountID:       "000000000000",
	}, nil
}
