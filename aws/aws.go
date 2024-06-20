package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"log"
	"os"
)

func LoadAWSConfig(ctx context.Context) (aws.Config, error) {
	// Attempt to load the default AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Unable to load default AWS config, falling back to environment variables: %v", err)

		// If default config fails, load from environment variables
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(os.Getenv("AWS_REGION")),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				os.Getenv("AWS_ACCESS_KEY_ID"),
				os.Getenv("AWS_SECRET_ACCESS_KEY"),
				"",
			)),
		)
		if err != nil {
			return aws.Config{}, err
		}
	}

	return cfg, nil
}

func AwsConfigWithRegion(region string) aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}
	return cfg
}
