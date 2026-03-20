package agents

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ListSnapshots retrieves all snapshots for a specific agent.
func ListSnapshots(agentID string) (*AgentSnapshotsResponse, error) {
	var response AgentSnapshotsResponse
	err := doJSON(http.MethodGet, "/"+agentID+"/snapshots", nil, &response)
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

// CreateSnapshot triggers a snapshot sync for a specific agent.
func CreateSnapshot(agentID string) (*StorageSyncResponse, error) {
	var response StorageSyncResponse
	err := doJSON(http.MethodPost, "/"+agentID+"/snapshots/sync", nil, &response)
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

// GetSyncStatus retrieves the storage sync status for a specific agent.
func GetSyncStatus(agentID string) (*StorageStatusResponse, error) {
	var response StorageStatusResponse
	err := doJSON(http.MethodGet, "/"+agentID+"/snapshots/sync", nil, &response)
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

// ResetSnapshot resets the agent to a specific snapshot.
func ResetSnapshot(agentID, snapshotCid string) (*ResetSnapshotResponse, error) {
	body := ResetSnapshotBody{
		SnapshotCid: snapshotCid,
	}

	var response ResetSnapshotResponse
	err := doJSON(http.MethodPost, "/"+agentID+"/snapshots/reset", body, &response)
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
