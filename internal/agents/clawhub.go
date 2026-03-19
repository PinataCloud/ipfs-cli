package agents

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"pinata/internal/config"
)

// ClawHub API uses a different base path
func doClawHubJSON(method, path string, body interface{}, result interface{}) error {
	host := config.GetAgentsHost()
	url := fmt.Sprintf("https://%s/v0/clawhub%s", host, path)

	resp, err := doRequestURL(method, url, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ErrorResponse
		if decodeErr := json.NewDecoder(resp.Body).Decode(&errResp); decodeErr == nil && errResp.Error != "" {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return errors.Join(err, errors.New("failed to decode response"))
		}
	}

	return nil
}

// ListHubSkills retrieves skills from ClawHub with optional filters.
func ListHubSkills(category, sort string, featured bool, cursor string) (*HubSkillListResponse, error) {
	path := ""
	sep := "?"

	if category != "" {
		path += sep + "category=" + category
		sep = "&"
	}
	if sort != "" {
		path += sep + "sort=" + sort
		sep = "&"
	}
	if featured {
		path += sep + "featured=true"
		sep = "&"
	}
	if cursor != "" {
		path += sep + "cursor=" + cursor
	}

	var response HubSkillListResponse
	err := doClawHubJSON(http.MethodGet, path, nil, &response)
	if err != nil {
		return nil, err
	}

	formattedJSON, err := json.MarshalIndent(response.Skills, "", "    ")
	if err != nil {
		return nil, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return &response, nil
}

// GetHubSkill retrieves a single skill from ClawHub by slug.
func GetHubSkill(slug string) (*HubSkillDetailResponse, error) {
	var response HubSkillDetailResponse
	err := doClawHubJSON(http.MethodGet, "/"+slug, nil, &response)
	if err != nil {
		return nil, err
	}

	formattedJSON, err := json.MarshalIndent(response.Skill, "", "    ")
	if err != nil {
		return nil, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return &response, nil
}

// InstallHubSkill installs a skill from ClawHub into the user's library.
func InstallHubSkill(hubSkillID string) (*InstallHubSkillResponse, error) {
	var response InstallHubSkillResponse
	err := doClawHubJSON(http.MethodPost, "/"+hubSkillID+"/install", nil, &response)
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
