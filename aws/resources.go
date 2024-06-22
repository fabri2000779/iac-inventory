package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	autoscalingType "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
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

// ListAutoScalingGroups lists all Auto Scaling Groups in the account
func ListAutoScalingGroups(cfg aws.Config) ([]autoscalingType.AutoScalingGroup, error) {
	svc := autoscaling.NewFromConfig(cfg)
	resp, err := svc.DescribeAutoScalingGroups(context.TODO(), &autoscaling.DescribeAutoScalingGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Auto Scaling Groups: %v", err)
	}
	return resp.AutoScalingGroups, nil
}

// ListResourcesInRegion lists all relevant resources in the specified region
func ListResourcesInRegion(cfg aws.Config, region string) (map[string][]string, error) {
	newCfg := cfg.Copy()
	newCfg.Region = region

	resources := make(map[string][]string)

	ec2Instances, err := ListEC2Instances(newCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to list EC2 instances in region %s: %v", region, err)
	}
	resources["aws_instance"] = ec2Instances

	rdsInstances, err := ListRDSInstances(newCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to list RDS instances in region %s: %v", region, err)
	}
	var rdsIdentifiers []string
	for _, instance := range rdsInstances {
		rdsIdentifiers = append(rdsIdentifiers, *instance.DBInstanceIdentifier)
	}
	resources["aws_db_instance"] = rdsIdentifiers

	lambdaFunctions, err := ListLambdaFunctions(newCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to list Lambda functions in region %s: %v", region, err)
	}
	resources["aws_lambda_function"] = lambdaFunctions

	asgs, err := ListAutoScalingGroups(newCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to list Auto Scaling Groups in region %s: %v", region, err)
	}
	for _, asg := range asgs {
		resources["aws_autoscaling_group"] = append(resources["aws_autoscaling_group"], *asg.AutoScalingGroupName)
		for _, instance := range asg.Instances {
			resources["aws_instance"] = append(resources["aws_instance"], *instance.InstanceId)
		}
	}

	return resources, nil
}
