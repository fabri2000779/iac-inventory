package drift

import (
	"drifter/aws"
	"encoding/json"
	"fmt"
	initAws "github.com/aws/aws-sdk-go-v2/aws"
)

type Output struct {
	ManagedByIAC []string `json:"Managed by IaC"`
	NotManaged   []string `json:"Not managed"`
	TotalManaged int      `json:"total resources managed"`
	Unmanaged    int      `json:"unmanaged resources"`
}

type TerraformState struct {
	Resources []struct {
		Mode      string `json:"mode"`
		Type      string `json:"type"`
		Name      string `json:"name"`
		Provider  string `json:"provider"`
		Instances []struct {
			SchemaVersion int `json:"schema_version"`
			Attributes    struct {
				ID string `json:"id"`
			} `json:"attributes"`
		} `json:"instances"`
	} `json:"resources"`
}

func DetectDrift(stateData []byte, cfg initAws.Config) ([]byte, error) {
	var tfState TerraformState
	err := json.Unmarshal(stateData, &tfState)
	if err != nil {
		return nil, err
	}

	managedResourcesMap := make(map[string]struct{})
	for _, resource := range tfState.Resources {
		if resource.Type == "aws_instance" {
			for _, instance := range resource.Instances {
				id := instance.Attributes.ID
				managedResourcesMap[id] = struct{}{}
			}
		}
	}

	fmt.Println(managedResourcesMap)

	var managedResources, unmanagedResources []string
	currentInstances, err := aws.ListEC2Instances(cfg)
	if err != nil {
		return nil, err
	}

	for _, instanceID := range currentInstances {
		if _, ok := managedResourcesMap[instanceID]; ok {
			managedResources = append(managedResources, instanceID)
		} else {
			unmanagedResources = append(unmanagedResources, instanceID)
		}
	}

	output := Output{
		ManagedByIAC: managedResources,
		NotManaged:   unmanagedResources,
		TotalManaged: len(managedResources),
		Unmanaged:    len(unmanagedResources),
	}

	outputJSON, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return nil, err
	}

	return outputJSON, nil
}
