package agents

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"pinata/internal/config"
)

// Templates API uses a different base URL than agents
func doTemplatesJSON(method, path string, body interface{}, result interface{}) error {
	host := config.GetAgentsHost()
	url := fmt.Sprintf("https://%s/v0/templates%s", host, path)

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

// ListTemplates retrieves all published templates, optionally filtered.
func ListTemplates(category string, featured bool) (*TemplateListResponse, error) {
	path := ""
	sep := "?"

	if category != "" {
		path += sep + "category=" + category
		sep = "&"
	}
	if featured {
		path += sep + "featured=true"
	}

	var response TemplateListResponse
	err := doTemplatesJSON(http.MethodGet, path, nil, &response)
	if err != nil {
		return nil, err
	}

	formattedJSON, err := json.MarshalIndent(response.Templates, "", "    ")
	if err != nil {
		return nil, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return &response, nil
}

// GetTemplate retrieves a template by its slug.
func GetTemplate(slug string) (*TemplateDetailResponse, error) {
	var response TemplateDetailResponse
	err := doTemplatesJSON(http.MethodGet, "/"+slug, nil, &response)
	if err != nil {
		return nil, err
	}

	formattedJSON, err := json.MarshalIndent(response.Template, "", "    ")
	if err != nil {
		return nil, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return &response, nil
}
