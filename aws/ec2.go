package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func ListEC2Instances(cfg aws.Config) ([]string, error) {
	svc := ec2.NewFromConfig(cfg)
	resp, err := svc.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list EC2 instances: %v", err)
	}

	var instanceIds []string
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			instanceIds = append(instanceIds, aws.ToString(instance.InstanceId))
		}
	}

	return instanceIds, nil
}

// ListRDSInstances lists all RDS instances in the account
func ListRDSInstances(cfg aws.Config) ([]types.DBInstance, error) {
	svc := rds.NewFromConfig(cfg)
	resp, err := svc.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list RDS instances: %v", err)
	}

	return resp.DBInstances, nil
}

// ListLambdaFunctions lists all Lambda functions in the account
func ListLambdaFunctions(cfg aws.Config) ([]string, error) {
	svc := lambda.NewFromConfig(cfg)
	resp, err := svc.ListFunctions(context.TODO(), &lambda.ListFunctionsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Lambda functions: %v", err)
	}

	var functionNames []string
	for _, function := range resp.Functions {
		functionNames = append(functionNames, aws.ToString(function.FunctionName))
	}

	return functionNames, nil
}
