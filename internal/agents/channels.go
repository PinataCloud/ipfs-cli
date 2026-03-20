package agents

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// GetChannelStatus retrieves the channel configuration status for an agent.
func GetChannelStatus(agentID string) (*ChannelStatusResponse, error) {
	var response ChannelStatusResponse
	err := doJSON(http.MethodGet, "/"+agentID+"/channels", nil, &response)
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

// ConfigureChannel configures a specific channel for an agent.
// channel must be one of: telegram, slack, discord, whatsapp
func ConfigureChannel(agentID, channel string, botToken, appToken, dmPolicy string, allowFrom []string) error {
	body := ConfigureChannelBody{
		BotToken:  botToken,
		AppToken:  appToken,
		DmPolicy:  dmPolicy,
		AllowFrom: allowFrom,
	}

	var response ConfigureChannelResponse
	err := doJSON(http.MethodPost, "/"+agentID+"/channels/"+channel, body, &response)
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

// RemoveChannel removes a channel configuration from an agent.
// channel must be one of: telegram, slack, discord, whatsapp
func RemoveChannel(agentID, channel string) error {
	err := doJSON(http.MethodDelete, "/"+agentID+"/channels/"+channel, nil, nil)
	if err != nil {
		return err
	}

	fmt.Println("Channel removed")

	return nil
}
