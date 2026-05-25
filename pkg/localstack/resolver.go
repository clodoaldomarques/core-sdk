package localstack

import "github.com/aws/aws-sdk-go-v2/aws"

type EndpointResolver struct {
}

func (e EndpointResolver) ResolveEndpoint(service, region string) (aws.Endpoint, error) {
	return aws.Endpoint{}, nil
}
