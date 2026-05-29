package sns

type Config interface {
	Region() string
	Address() string
	AccessKeyID() string
	SecretAccessKey() string
	TopicARN() string
}
