package drift

import (
	"drifter/aws"
	"encoding/json"
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

func DetectDrift(stateData []byte, cfg initAws.Config) (map[string]struct{}, map[string]struct{}, error) {
	var tfState TerraformState
	err := json.Unmarshal(stateData, &tfState)
	if err != nil {
		return nil, nil, err
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

	currentInstances, err := aws.ListEC2Instances(cfg)
	if err != nil {
		return nil, nil, err
	}

	managedResources := make(map[string]struct{})
	unmanagedResources := make(map[string]struct{})

	for _, instanceID := range currentInstances {
		if _, ok := managedResourcesMap[instanceID]; ok {
			managedResources[instanceID] = struct{}{}
		} else {
			unmanagedResources[instanceID] = struct{}{}
		}
	}

	return managedResources, unmanagedResources, nil
}

func FormatOutput(managedResources, unmanagedResources map[string]struct{}) ([]byte, error) {
	managedList := make([]string, 0, len(managedResources))
	for id := range managedResources {
		managedList = append(managedList, id)
	}

	unmanagedList := make([]string, 0, len(unmanagedResources))
	for id := range unmanagedResources {
		unmanagedList = append(unmanagedList, id)
	}

	output := Output{
		ManagedByIAC: managedList,
		NotManaged:   unmanagedList,
		TotalManaged: len(managedList),
		Unmanaged:    len(unmanagedList),
	}

	outputJSON, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return nil, err
	}

	return outputJSON, nil
}
