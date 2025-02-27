package s3client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewClient(ctx context.Context, keyID, secret string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithCredentialsProvider(
			&credentials.StaticCredentialsProvider{
				Value: aws.Credentials{
					AccessKeyID:     keyID,
					SecretAccessKey: secret,
				},
			},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load default s3 config, err: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return client, nil
}
