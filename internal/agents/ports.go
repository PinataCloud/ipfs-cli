package agents

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ListPorts retrieves the port forwarding rules for a specific agent.
func ListPorts(agentID string) (*PortForwardingResponse, error) {
	var response PortForwardingResponse
	err := doJSON(http.MethodGet, "/"+agentID+"/port-forwarding", nil, &response)
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

// UpdatePorts updates the port forwarding rules for a specific agent.
func UpdatePorts(agentID string, mappings []PortForwarding) (*UpdatePortForwardingResponse, error) {
	body := UpdatePortForwardingBody{
		Mappings: mappings,
	}

	var response UpdatePortForwardingResponse
	err := doJSON(http.MethodPut, "/"+agentID+"/port-forwarding", body, &response)
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
