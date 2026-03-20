package agents

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"pinata/internal/common"
	"pinata/internal/config"
)

// buildSecretsURL constructs the full URL for a Secrets API endpoint
func buildSecretsURL(path string) string {
	return fmt.Sprintf("https://%s/%s/secrets%s", config.GetAgentsHost(), apiVersion, path)
}

// doSecretsRequest makes an authenticated HTTP request to the Secrets API
func doSecretsRequest(method, path string, body interface{}) (*http.Response, error) {
	jwt, err := common.FindToken()
	if err != nil {
		return nil, err
	}

	var reqBody io.Reader
	if body != nil {
		jsonPayload, err := json.Marshal(body)
		if err != nil {
			return nil, errors.Join(err, errors.New("failed to marshal request body"))
		}
		reqBody = bytes.NewBuffer(jsonPayload)
	}

	url := buildSecretsURL(path)
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to create the request"))
	}

	req.Header.Set("Authorization", "Bearer "+string(jwt))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to send the request"))
	}

	return resp, nil
}

// doSecretsJSON makes an authenticated request to the Secrets API and decodes the JSON response
func doSecretsJSON(method, path string, body interface{}, result interface{}) error {
	resp, err := doSecretsRequest(method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		err = json.NewDecoder(resp.Body).Decode(result)
		if err != nil {
			return errors.Join(err, errors.New("failed to decode response"))
		}
	}

	return nil
}

// ListSecrets retrieves all secrets for the authenticated user.
func ListSecrets() ([]SecretWithAgents, error) {
	var response SecretListResponse
	err := doSecretsJSON(http.MethodGet, "", nil, &response)
	if err != nil {
		return nil, err
	}

	formattedJSON, err := json.MarshalIndent(response.Secrets, "", "    ")
	if err != nil {
		return nil, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return response.Secrets, nil
}

// CreateSecret creates a new secret with the specified name and value.
func CreateSecret(name, value string) (*CreateSecretResponse, error) {
	body := CreateSecretBody{
		Name:  name,
		Value: value,
	}

	var response CreateSecretResponse
	err := doSecretsJSON(http.MethodPost, "", body, &response)
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

// UpdateSecret updates an existing secret's value.
func UpdateSecret(secretID, value string) error {
	body := UpdateSecretBody{
		Value: value,
	}

	err := doSecretsJSON(http.MethodPut, "/"+secretID, body, nil)
	if err != nil {
		return err
	}

	fmt.Println("Secret updated")

	return nil
}

// DeleteSecret deletes a secret by ID.
func DeleteSecret(secretID string) error {
	err := doSecretsJSON(http.MethodDelete, "/"+secretID, nil, nil)
	if err != nil {
		return err
	}

	fmt.Println("Secret deleted")

	return nil
}

// AttachSecrets attaches secrets to an agent.
func AttachSecrets(agentID string, secretIds []string) error {
	body := AddSecretsBody{
		SecretIds: secretIds,
	}

	var response AddSecretsResponse
	err := doJSON(http.MethodPost, "/"+agentID+"/secrets", body, &response)
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

// DetachSecret detaches a secret from an agent.
func DetachSecret(agentID, secretID string) error {
	err := doJSON(http.MethodDelete, "/"+agentID+"/secrets/"+secretID, nil, nil)
	if err != nil {
		return err
	}

	fmt.Println("Secret detached from agent")

	return nil
}
