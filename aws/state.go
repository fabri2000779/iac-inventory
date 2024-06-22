package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"strings"
)

// ListBuckets lists all the S3 buckets in the account
func ListBuckets(cfg aws.Config) ([]string, error) {
	svc := s3.NewFromConfig(cfg)
	resp, err := svc.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("unable to list buckets: %v", err)
	}

	var buckets []string
	for _, bucket := range resp.Buckets {
		buckets = append(buckets, aws.ToString(bucket.Name))
	}

	return buckets, nil
}

func GetBucketRegion(bucket string, cfg aws.Config) (string, error) {
	svc := s3.NewFromConfig(cfg)
	resp, err := svc.GetBucketLocation(context.TODO(), &s3.GetBucketLocationInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return "", fmt.Errorf("unable to get bucket location, %v", err)
	}

	if resp.LocationConstraint == "" {
		return "us-east-1", nil
	}

	return string(resp.LocationConstraint), nil
}

// ListObjects lists the objects in the specified S3 bucket with the specified prefix
func ListObjects(cfg aws.Config, bucket string) ([]string, error) {
	region, err := GetBucketRegion(bucket, cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to get bucket region, %v", err)
	}

	cfg.Region = region
	svc := s3.NewFromConfig(cfg)

	var keys []string
	paginator := s3.NewListObjectsV2Paginator(svc, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to fetch page, %v", err)
		}

		for _, item := range page.Contents {
			key := aws.ToString(item.Key)
			if strings.HasSuffix(key, ".tfstate") {
				keys = append(keys, key)
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("unable to list objects, %v", err)
	}

	return keys, nil
}

// GetTerraformState gets the Terraform state file from the specified S3 bucket and key
func GetTerraformState(cfg aws.Config, bucket, key string) ([]byte, error) {
	// Get the correct region for the bucket
	region, err := GetBucketRegion(bucket, cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to get bucket region, %v", err)
	}

	// Create a new AWS client with the correct region
	cfg.Region = region
	svc := s3.NewFromConfig(cfg)

	resp, err := svc.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to download item from S3, %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read object body, %v", err)
	}

	return body, nil
}
