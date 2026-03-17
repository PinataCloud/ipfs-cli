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

// buildSkillsURL constructs the full URL for a Skills API endpoint
func buildSkillsURL(path string) string {
	return fmt.Sprintf("https://%s/%s/skills%s", config.GetAgentsHost(), apiVersion, path)
}

// doSkillsRequest makes an authenticated HTTP request to the Skills API
func doSkillsRequest(method, path string, body interface{}) (*http.Response, error) {
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

	url := buildSkillsURL(path)
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

// doSkillsJSON makes an authenticated request to the Skills API and decodes the JSON response
func doSkillsJSON(method, path string, body interface{}, result interface{}) error {
	resp, err := doSkillsRequest(method, path, body)
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

// ListSkills retrieves all skills for the authenticated user.
func ListSkills() ([]Skill, error) {
	var response SkillListResponse
	err := doSkillsJSON(http.MethodGet, "", nil, &response)
	if err != nil {
		return nil, err
	}

	formattedJSON, err := json.MarshalIndent(response.Skills, "", "    ")
	if err != nil {
		return nil, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return response.Skills, nil
}

// CreateSkill creates a new skill with the specified parameters.
func CreateSkill(skillCid, name, description string, envVars []string, fileId string) (*CreateSkillResponse, error) {
	body := CreateSkillBody{
		SkillCid:    skillCid,
		Name:        name,
		Description: description,
		EnvVars:     envVars,
		FileID:      fileId,
	}

	var response CreateSkillResponse
	err := doSkillsJSON(http.MethodPost, "", body, &response)
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

// DeleteSkill deletes a skill by its CID.
func DeleteSkill(skillCid string) error {
	var response DeleteSkillResponse
	err := doSkillsJSON(http.MethodDelete, "/"+skillCid, nil, &response)
	if err != nil {
		return err
	}

	fmt.Println("Skill deleted")

	return nil
}

// AttachSkills attaches skills to an agent.
func AttachSkills(agentID string, skillCids []string) error {
	body := AddSkillsBody{
		SkillCids: skillCids,
	}

	var response AddSkillsResponse
	err := doJSON(http.MethodPost, "/"+agentID+"/skills", body, &response)
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

// DetachSkill detaches a skill from an agent.
func DetachSkill(agentID, skillID string) error {
	err := doJSON(http.MethodDelete, "/"+agentID+"/skills/"+skillID, nil, nil)
	if err != nil {
		return err
	}

	fmt.Println("Skill detached from agent")

	return nil
}
