package terraform

import (
	"encoding/json"
	"fmt"
)

func ParseTerraformState(stateData []byte) (map[string]interface{}, error) {
	var state map[string]interface{}
	err := json.Unmarshal(stateData, &state)
	if err != nil {
		return nil, fmt.Errorf("error parsing Terraform state: %v", err)
	}
	return state, nil
}
