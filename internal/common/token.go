package common

import (
	"errors"
	"os"
	"path/filepath"
)

// FindToken extracts the JWT token from the .pinata-files-cli file
func FindToken() ([]byte, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dotFilePath := filepath.Join(homeDir, ".pinata-files-cli")
	JWT, err := os.ReadFile(dotFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("JWT not found. Please authorize first using the 'auth' command")
		} else {
			return nil, err
		}
	}
	return JWT, err
}

// FindGatewayDomain gets the gateway domain from the config file
func FindGatewayDomain() ([]byte, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dotFilePath := filepath.Join(homeDir, ".pinata-files-cli-gateway")
	Domain, err := os.ReadFile(dotFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("Gateway domain not found. Please set a gateway first")
		} else {
			return nil, err
		}
	}
	return Domain, err
}
