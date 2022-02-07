package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

type S3Config struct {
	Sign     bool
	GPGKeyID string

	Region           string
	Bucket           string
	Endpoint         string
	DisableSSL       bool
	S3ForcePathStyle bool

	UseStaticCreds bool
	AccessKey      string
	SecretKey      string
}

func (s S3Config) GenerateS3Config() *aws.Config {
	c := &aws.Config{
		Region: aws.String(s.Region),
	}
	if s.Endpoint != "" {
		c.Endpoint = aws.String(s.Endpoint)
	}

	c.DisableSSL = &s.DisableSSL
	c.S3ForcePathStyle = &s.S3ForcePathStyle

	if s.UseStaticCreds {
		c.Credentials = credentials.NewStaticCredentials(
			s.AccessKey,
			s.SecretKey,
			"",
		)
	}
	return c
}
