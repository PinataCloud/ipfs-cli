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

// buildFeedbackURL constructs the full URL for the feedback endpoint.
// Note: Feedback endpoint is /v0/feedback, NOT under /v0/agents.
func buildFeedbackURL() string {
	return fmt.Sprintf("https://%s/%s/feedback", config.GetAgentsHost(), apiVersion)
}

// SubmitFeedback submits user feedback.
func SubmitFeedback(message string) error {
	jwt, err := common.FindToken()
	if err != nil {
		return err
	}

	body := FeedbackBody{
		Message: message,
	}

	jsonPayload, err := json.Marshal(body)
	if err != nil {
		return errors.Join(err, errors.New("failed to marshal request body"))
	}

	url := buildFeedbackURL()
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return errors.Join(err, errors.New("failed to create the request"))
	}

	req.Header.Set("Authorization", "Bearer "+string(jwt))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Join(err, errors.New("failed to send the request"))
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	fmt.Println("Feedback submitted successfully")

	return nil
}
