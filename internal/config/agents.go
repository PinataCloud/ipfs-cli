package config

import (
	"os"
	"path/filepath"
	"strings"
)

// GetAgentsHost returns the Agents API host URL from environment or default
func GetAgentsHost() string {
	return getEnv("PINATA_AGENTS_HOST", "agents.pinata.cloud")
}

// GetChatModel returns the chat model to use for agents
// It checks PINATA_CHAT_MODEL env var first, then ~/.pinata-agents-model file,
// then returns empty string if neither is set
func GetChatModel() string {
	// First check environment variable
	model := os.Getenv("PINATA_CHAT_MODEL")
	if len(model) > 0 {
		return model
	}

	// Then check config file
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	p := filepath.Join(home, ".pinata-agents-model")
	data, err := os.ReadFile(p)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}
