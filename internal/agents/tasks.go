package agents

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ListTasks retrieves all tasks (cron jobs) for a specific agent.
// Set includeDisabled to true to include disabled tasks in the response.
func ListTasks(agentID string, includeDisabled bool) (interface{}, error) {
	path := fmt.Sprintf("/%s/tasks?includeDisabled=%t", agentID, includeDisabled)

	var result interface{}
	err := doJSON(http.MethodGet, path, nil, &result)
	if err != nil {
		return nil, err
	}

	formattedJSON, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		return nil, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return result, nil
}

// CreateTask creates a new task (cron job) for a specific agent.
func CreateTask(agentID string, body CreateTaskBody) (interface{}, error) {
	path := fmt.Sprintf("/%s/tasks", agentID)

	var result interface{}
	err := doJSON(http.MethodPost, path, body, &result)
	if err != nil {
		return nil, err
	}

	formattedJSON, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		return nil, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return result, nil
}

// UpdateTask updates an existing task (cron job) for a specific agent.
func UpdateTask(agentID, jobID string, body UpdateTaskBody) (interface{}, error) {
	path := fmt.Sprintf("/%s/tasks/%s", agentID, jobID)

	var result interface{}
	err := doJSON(http.MethodPut, path, body, &result)
	if err != nil {
		return nil, err
	}

	formattedJSON, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		return nil, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return result, nil
}

// DeleteTask deletes a task (cron job) from a specific agent.
func DeleteTask(agentID, jobID string) error {
	path := fmt.Sprintf("/%s/tasks/%s", agentID, jobID)

	err := doJSON(http.MethodDelete, path, nil, nil)
	if err != nil {
		return err
	}

	fmt.Println("Task deleted")

	return nil
}

// ToggleTask enables or disables a task (cron job) for a specific agent.
func ToggleTask(agentID, jobID string, enabled bool) error {
	path := fmt.Sprintf("/%s/tasks/%s/toggle", agentID, jobID)

	body := ToggleTaskBody{
		Enabled: enabled,
	}

	err := doJSON(http.MethodPost, path, body, nil)
	if err != nil {
		return err
	}

	if enabled {
		fmt.Println("Task enabled")
	} else {
		fmt.Println("Task disabled")
	}

	return nil
}

// RunTask manually triggers a task (cron job) to run immediately.
func RunTask(agentID, jobID string) error {
	path := fmt.Sprintf("/%s/tasks/%s/run", agentID, jobID)

	err := doJSON(http.MethodPost, path, nil, nil)
	if err != nil {
		return err
	}

	fmt.Println("Task triggered")

	return nil
}

// GetTaskHistory retrieves the run history for a specific task.
// The limit parameter controls how many history entries to return.
func GetTaskHistory(agentID, jobID string, limit int) (interface{}, error) {
	path := fmt.Sprintf("/%s/tasks/%s/runs?limit=%d", agentID, jobID, limit)

	var result interface{}
	err := doJSON(http.MethodGet, path, nil, &result)
	if err != nil {
		return nil, err
	}

	formattedJSON, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		return nil, errors.New("failed to format JSON")
	}

	fmt.Println(string(formattedJSON))

	return result, nil
}
