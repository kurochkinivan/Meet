package s3client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewClient(ctx context.Context) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, func(lo *config.LoadOptions) error {
		lo.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load default s3 config, err: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return client, nil
}
