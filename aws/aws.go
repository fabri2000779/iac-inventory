package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/sts"
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

func AssumeRole(cfg aws.Config, roleArn string) (aws.Config, error) {
	stsSvc := sts.NewFromConfig(cfg)

	creds := stscreds.NewAssumeRoleProvider(stsSvc, roleArn)

	// Create a new configuration with the assumed role credentials
	newCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(creds))
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to assume role: %v", err)
	}

	return newCfg, nil
}

// ListAccounts lists all AWS accounts in the organization
func ListAccounts(cfg aws.Config) ([]string, error) {
	svc := organizations.NewFromConfig(cfg)
	var accountIDs []string
	var nextToken *string

	for {
		resp, err := svc.ListAccounts(context.TODO(), &organizations.ListAccountsInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to list accounts: %v", err)
		}

		for _, account := range resp.Accounts {
			accountIDs = append(accountIDs, aws.ToString(account.Id))
		}

		if resp.NextToken == nil {
			break
		}
		nextToken = resp.NextToken
	}

	return accountIDs, nil
}
