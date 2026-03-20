package agents

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// ExecCommand executes a command on the agent's console.
func ExecCommand(agentID, command, cwd string) (*ExecResponse, error) {
	body := ExecRequest{
		Command: command,
		Cwd:     cwd,
	}

	var response ExecResponse
	err := doJSON(http.MethodPost, "/"+agentID+"/console/exec", body, &response)
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

// ReadFile reads a file from the agent's filesystem.
func ReadFile(agentID, filePath string) ([]byte, error) {
	encodedPath := url.QueryEscape(filePath)
	resp, err := doRequest(http.MethodGet, "/"+agentID+"/files?path="+encodedPath, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to read response body"))
	}

	fmt.Println(string(content))

	return content, nil
}
