package drift

import (
	"drifter/aws"
	"encoding/json"
	"fmt"
	initAws "github.com/aws/aws-sdk-go-v2/aws"
)

type ARN struct {
	AccountID  string
	ResourceID string
	Type       string
}

type Output struct {
	ManagedByIAC map[string][]string `json:"Managed by IaC"`
	NotManaged   map[string][]string `json:"Not managed"`
	TotalManaged int                 `json:"total resources managed"`
	Unmanaged    int                 `json:"unmanaged resources"`
}

type TerraformState struct {
	Resources []struct {
		Type      string `json:"type"`
		Instances []struct {
			Attributes map[string]interface{} `json:"attributes"`
		} `json:"instances"`
	} `json:"resources"`
}

// ExtractResourceIdentifiers extracts identifiers for EC2 instances, RDS instances, and Lambda functions from the Terraform state data
func ExtractResourceIdentifiers(stateData []byte) ([]ARN, error) {
	var tfState TerraformState
	err := json.Unmarshal(stateData, &tfState)
	if err != nil {
		return nil, fmt.Errorf("error parsing Terraform state: %v", err)
	}

	var arns []ARN
	for _, resource := range tfState.Resources {
		if resource.Type == "aws_instance" || resource.Type == "aws_db_instance" || resource.Type == "aws_lambda_function" {
			for _, instance := range resource.Instances {
				var resourceID string
				if resource.Type == "aws_db_instance" {
					if identifier, ok := instance.Attributes["identifier"].(string); ok {
						resourceID = identifier
					} else if id, ok := instance.Attributes["id"].(string); ok {
						resourceID = id
					}
				} else {
					if id, ok := instance.Attributes["id"].(string); ok {
						resourceID = id
					}
				}
				if resourceID != "" {
					arns = append(arns, ARN{
						ResourceID: resourceID,
						Type:       resource.Type,
					})
				}
			}
		}
	}

	return arns, nil
}

// DetectDriftForResources detects drift for EC2 instances, RDS instances, and Lambda functions
func DetectDriftForResources(resourceIdentifiers []ARN, cfg initAws.Config) (map[string]map[string]struct{}, map[string]map[string]struct{}, error) {
	managedResources := make(map[string]map[string]struct{})
	unmanagedResources := make(map[string]map[string]struct{})

	// Initialize maps for each resource type
	resourceTypes := []string{"aws_instance", "aws_db_instance", "aws_lambda_function"}
	for _, resourceType := range resourceTypes {
		managedResources[resourceType] = make(map[string]struct{})
		unmanagedResources[resourceType] = make(map[string]struct{})
	}

	for _, arn := range resourceIdentifiers {
		managedResources[arn.Type][arn.ResourceID] = struct{}{}
	}

	// Check for EC2 instances
	currentInstances, err := aws.ListEC2Instances(cfg)
	if err != nil {
		return nil, nil, err
	}
	for _, instanceID := range currentInstances {
		if _, exists := managedResources["aws_instance"][instanceID]; !exists {
			unmanagedResources["aws_instance"][instanceID] = struct{}{}
		} else {
			delete(unmanagedResources["aws_instance"], instanceID) // Ensure it's not in unmanaged if it is managed
		}
	}

	// Check for RDS instances
	currentRDSInstances, err := aws.ListRDSInstances(cfg)
	if err != nil {
		return nil, nil, err
	}
	for _, dbInstance := range currentRDSInstances {
		dbInstanceIdentifier := *dbInstance.DBInstanceIdentifier
		if _, exists := managedResources["aws_db_instance"][dbInstanceIdentifier]; !exists {
			unmanagedResources["aws_db_instance"][dbInstanceIdentifier] = struct{}{}
		} else {
			delete(unmanagedResources["aws_db_instance"], dbInstanceIdentifier) // Ensure it's not in unmanaged if it is managed
		}
	}

	// Check for Lambda functions
	currentLambdas, err := aws.ListLambdaFunctions(cfg)
	if err != nil {
		return nil, nil, err
	}
	for _, functionName := range currentLambdas {
		if _, exists := managedResources["aws_lambda_function"][functionName]; !exists {
			unmanagedResources["aws_lambda_function"][functionName] = struct{}{}
		} else {
			delete(unmanagedResources["aws_lambda_function"], functionName) // Ensure it's not in unmanaged if it is managed
		}
	}

	return managedResources, unmanagedResources, nil
}

func FormatOutput(managedResources, unmanagedResources map[string]map[string]struct{}) ([]byte, error) {
	managedOutput := make(map[string][]string)
	unmanagedOutput := make(map[string][]string)

	for resourceType, resources := range managedResources {
		for id := range resources {
			managedOutput[resourceType] = append(managedOutput[resourceType], id)
		}
	}

	for resourceType, resources := range unmanagedResources {
		for id := range resources {
			unmanagedOutput[resourceType] = append(unmanagedOutput[resourceType], id)
		}
	}

	output := Output{
		ManagedByIAC: managedOutput,
		NotManaged:   unmanagedOutput,
		TotalManaged: len(managedOutput["aws_instance"]) + len(managedOutput["aws_db_instance"]) + len(managedOutput["aws_lambda_function"]),
		Unmanaged:    len(unmanagedOutput["aws_instance"]) + len(unmanagedOutput["aws_db_instance"]) + len(unmanagedOutput["aws_lambda_function"]),
	}

	outputJSON, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return nil, err
	}

	return outputJSON, nil
}
