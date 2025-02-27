package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	NetworkPublic  = "public"
	NetworkPrivate = "private"
)

// SetDefaultNetwork saves the user's preferred network to the config file
func SetDefaultNetwork(network string) error {
	// Validate network value
	if network != NetworkPublic && network != NetworkPrivate {
		return fmt.Errorf("invalid network: %s. Must be either 'public' or 'private'", network)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	p := filepath.Join(home, ".pinata-files-cli-network")
	err = os.WriteFile(p, []byte(network), 0600)
	if err != nil {
		return err
	}

	fmt.Printf("Default network set to '%s'\n", network)
	return nil
}

// GetDefaultNetwork retrieves the user's preferred network from config
// If no preference is set, it returns "public" as the default
func GetDefaultNetwork() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	p := filepath.Join(home, ".pinata-files-cli-network")
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, return default
			return NetworkPublic, nil
		}
		return "", err
	}

	network := string(data)
	if network != NetworkPublic && network != NetworkPrivate {
		return NetworkPublic, errors.New("invalid network in config file")
	}

	return network, nil
}

// GetNetworkParam returns the network to use based on input param or default
// If the passed network is empty, it will use the user's preference
func GetNetworkParam(network string) (string, error) {
	if network == "" {
		// No network specified, use default
		return GetDefaultNetwork()
	}

	// Validate the provided network
	if network != NetworkPublic && network != NetworkPrivate {
		return "", fmt.Errorf("invalid network: %s. Must be either 'public' or 'private'", network)
	}

	return network, nil
}
