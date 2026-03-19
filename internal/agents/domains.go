package agents

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ListDomains retrieves all custom domains for a specific agent.
func ListDomains(agentID string) (*CustomDomainListResponse, error) {
	var response CustomDomainListResponse
	err := doJSON(http.MethodGet, "/"+agentID+"/domains", nil, &response)
	if err != nil {
		return nil, err
	}

	formattedJSON, err := json.MarshalIndent(response.Domains, "", "    ")
	if err != nil {
		return nil, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return &response, nil
}

// CreateDomain registers a new custom domain for an agent.
func CreateDomain(agentID string, subdomain, customDomain string, targetPort int, protected bool) (*CreateCustomDomainResponse, error) {
	body := CreateCustomDomainBody{
		Subdomain:    subdomain,
		CustomDomain: customDomain,
		TargetPort:   targetPort,
		Protected:    protected,
	}

	var response CreateCustomDomainResponse
	err := doJSON(http.MethodPost, "/"+agentID+"/domains", body, &response)
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

// UpdateDomain updates an existing custom domain mapping.
func UpdateDomain(agentID, domainID string, subdomain, customDomain string, targetPort *int, protected *bool) (*UpdateCustomDomainResponse, error) {
	body := UpdateCustomDomainBody{
		Subdomain:    subdomain,
		CustomDomain: customDomain,
		TargetPort:   targetPort,
		Protected:    protected,
	}

	var response UpdateCustomDomainResponse
	err := doJSON(http.MethodPut, "/"+agentID+"/domains/"+domainID, body, &response)
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

// DeleteDomain removes a custom domain mapping.
func DeleteDomain(agentID, domainID string) (*DeleteCustomDomainResponse, error) {
	var response DeleteCustomDomainResponse
	err := doJSON(http.MethodDelete, "/"+agentID+"/domains/"+domainID, nil, &response)
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
