package agents

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// GetConfig retrieves the openclaw config for a specific agent.
func GetConfig(agentID string) (*ConfigResponse, error) {
	var response ConfigResponse
	err := doJSON(http.MethodGet, "/"+agentID+"/config", nil, &response)
	if err != nil {
		return nil, err
	}

	formattedJSON, err := json.MarshalIndent(response.Config, "", "    ")
	if err != nil {
		return nil, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return &response, nil
}

// SetConfig updates the openclaw config for a specific agent.
func SetConfig(agentID string, configData interface{}) error {
	body := UpdateConfigBody{
		Config: configData,
	}

	var response map[string]interface{}
	err := doJSON(http.MethodPut, "/"+agentID+"/config", body, &response)
	if err != nil {
		return err
	}

	formattedJSON, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		return errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return nil
}

// ValidateConfig validates the openclaw config for a specific agent.
func ValidateConfig(agentID string) (*ValidateConfigResponse, error) {
	var response ValidateConfigResponse
	err := doJSON(http.MethodPost, "/"+agentID+"/config/validate", nil, &response)
	if err != nil {
		return nil, err
	}

	formattedJSON, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		return nil, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return &response, nil
}

// CheckUpdate checks if there's an openclaw update available for an agent.
func CheckUpdate(agentID string) (*UpdateCheckResponse, error) {
	var response UpdateCheckResponse
	err := doJSON(http.MethodGet, "/"+agentID+"/update", nil, &response)
	if err != nil {
		return nil, err
	}

	formattedJSON, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		return nil, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return &response, nil
}

// ApplyUpdate applies an openclaw update to an agent.
func ApplyUpdate(agentID, tag string) (*UpdateApplyResponse, error) {
	var body *UpdateApplyBody
	if tag != "" {
		body = &UpdateApplyBody{Tag: tag}
	}

	var response UpdateApplyResponse
	err := doJSON(http.MethodPost, "/"+agentID+"/update", body, &response)
	if err != nil {
		return nil, err
	}

	formattedJSON, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		return nil, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return &response, nil
}

// GetAvailableVersions retrieves available agent versions.
func GetAvailableVersions(agentID string) (*VersionsResponse, error) {
	var response VersionsResponse
	err := doJSON(http.MethodGet, "/"+agentID+"/available-agent-versions", nil, &response)
	if err != nil {
		return nil, err
	}

	formattedJSON, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		return nil, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return &response, nil
}
