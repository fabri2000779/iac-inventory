package controller

import (
	"bytes"
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

	var detectDrifts [][]byte

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

			detectDrift, err := drift.DetectDrift(stateData, cfg)
			if err != nil {
				log.Printf("failed to detect drift in bucket %s, key %s: %v", bucket, key, err)
				continue
			}

			detectDrifts = append(detectDrifts, detectDrift)
		}
	}

	return bytes.Join(detectDrifts, []byte("\n")), nil
}
