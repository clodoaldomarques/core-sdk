package aws

type Config interface {
	Region() string
	Address() string
	AccessKeyID() string
	SecretAccessKey() string
}
