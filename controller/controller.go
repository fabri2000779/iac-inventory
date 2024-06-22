package controller

import (
	"fmt"
	initAws "github.com/aws/aws-sdk-go-v2/aws"
	"log"
	"sync"

	"drifter/aws"
	"drifter/drift"
)

const (
	readRoleArnFormat = "arn:aws:iam::%s:role/ReadOnlyRole" //
)

func Run(cfg initAws.Config, regions []string) ([]byte, error) {
	// List all accounts in the organization using the management configuration
	accountIDs, err := aws.ListAccounts(cfg)
	if err != nil {
		return nil, err
	}

	managedResources := make(map[string]map[string]struct{})
	unmanagedResources := make(map[string]map[string]struct{})

	// Initialize maps for each resource type
	resourceTypes := []string{"aws_instance", "aws_db_instance", "aws_lambda_function", "aws_autoscaling_group"}
	for _, resourceType := range resourceTypes {
		managedResources[resourceType] = make(map[string]struct{})
		unmanagedResources[resourceType] = make(map[string]struct{})
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Iterate over each account to find S3 buckets with Terraform state files
	for _, accountID := range accountIDs {
		wg.Add(1)
		go func(accountID string) {
			defer wg.Done()

			readRoleArn := fmt.Sprintf(readRoleArnFormat, accountID)
			accountCfg, err := aws.AssumeRole(cfg, readRoleArn)
			if err != nil {
				log.Printf("failed to assume role for account %s: %v", accountID, err)
				return
			}

			buckets, err := aws.ListBuckets(accountCfg)
			if err != nil {
				log.Printf("failed to list buckets for account %s: %v", accountID, err)
				return
			}

			for _, bucket := range buckets {
				keys, err := aws.ListObjects(accountCfg, bucket)
				if err != nil {
					log.Printf("failed to list objects in bucket %s: %v", bucket, err)
					continue
				}

				for _, key := range keys {
					stateData, err := aws.GetTerraformState(accountCfg, bucket, key)
					if err != nil {
						log.Printf("failed to get Terraform state from bucket %s, key %s: %v", bucket, key, err)
						continue
					}

					// Extract resource identifiers from the state data
					arns, err := drift.ExtractResourceIdentifiers(stateData)
					if err != nil {
						log.Printf("failed to extract resource identifiers from state data in bucket %s, key %s: %v", bucket, key, err)
						continue
					}

					// Detect drift for the resources
					managed, unmanaged, err := drift.DetectDriftForResources(arns, accountCfg, regions)
					if err != nil {
						log.Printf("failed to detect drift for account %s: %v", accountID, err)
						continue
					}

					mu.Lock()
					for resourceType, resources := range managed {
						for id := range resources {
							managedResources[resourceType][id] = struct{}{}
							delete(unmanagedResources[resourceType], id) // Ensure it's not in unmanaged if it is managed
						}
					}
					for resourceType, resources := range unmanaged {
						for id := range resources {
							if _, exists := managedResources[resourceType][id]; !exists {
								unmanagedResources[resourceType][id] = struct{}{}
							}
						}
					}
					mu.Unlock()
				}
			}
		}(accountID)
	}

	wg.Wait()

	output, err := drift.FormatOutput(managedResources, unmanagedResources)
	if err != nil {
		return nil, err
	}

	return output, nil
}
