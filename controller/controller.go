package controller

import (
	initAws "github.com/aws/aws-sdk-go-v2/aws"
	"log"

	"drifter/aws"
	"drifter/drift"
)

func Run(cfg initAws.Config) ([]byte, error) {
	buckets, err := aws.ListBuckets(cfg)
	if err != nil {
		return nil, err
	}

	managedResources := make(map[string]struct{})
	unmanagedResources := make(map[string]struct{})

	for _, bucket := range buckets {
		keys, err := aws.ListObjects(cfg, bucket)
		if err != nil {
			log.Printf("failed to list objects in bucket %s: %v", bucket, err)
			continue
		}

		for _, key := range keys {
			stateData, err := aws.GetTerraformState(cfg, bucket, key)
			if err != nil {
				log.Printf("failed to get Terraform state from bucket %s, key %s: %v", bucket, key, err)
				continue
			}

			managed, unmanaged, err := drift.DetectDrift(stateData, cfg)
			if err != nil {
				log.Printf("failed to detect drift in bucket %s, key %s: %v", bucket, key, err)
				continue
			}

			for id := range managed {
				managedResources[id] = struct{}{}
			}
			for id := range unmanaged {
				unmanagedResources[id] = struct{}{}
			}
		}
	}

	output, err := drift.FormatOutput(managedResources, unmanagedResources)
	if err != nil {
		return nil, err
	}

	return output, nil
}
