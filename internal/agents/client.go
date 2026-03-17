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

const apiVersion = "v0"

// buildURL constructs the full URL for an Agents API endpoint
func buildURL(path string) string {
	return fmt.Sprintf("https://%s/%s/agents%s", config.GetAgentsHost(), apiVersion, path)
}

// doRequest makes an authenticated HTTP request to the Agents API
func doRequest(method, path string, body interface{}) (*http.Response, error) {
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

	url := buildURL(path)
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

// doJSON makes an authenticated request and decodes the JSON response
func doJSON(method, path string, body interface{}, result interface{}) error {
	resp, err := doRequest(method, path, body)
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
