package agents

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ListDevices retrieves all devices for a specific agent.
func ListDevices(agentID string) (*DeviceListResponse, error) {
	var response DeviceListResponse
	err := doJSON(http.MethodGet, "/"+agentID+"/devices", nil, &response)
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

// ApproveDevice approves a pending device pairing request.
func ApproveDevice(agentID, requestID string) error {
	var response ApproveDeviceResponse
	err := doJSON(http.MethodPost, "/"+agentID+"/devices/"+requestID+"/approve", nil, &response)
	if err != nil {
		return err
	}

	fmt.Printf("Device %s approved successfully\n", requestID)

	return nil
}

// ApproveAllDevices approves all pending device pairing requests.
func ApproveAllDevices(agentID string) (*ApproveAllResponse, error) {
	var response ApproveAllResponse
	err := doJSON(http.MethodPost, "/"+agentID+"/devices/approve-all", nil, &response)
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
