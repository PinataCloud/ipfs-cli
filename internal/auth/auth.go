package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"pinata/internal/config"
	"pinata/internal/gateways"
	"pinata/internal/utils"
	"time"
)

func SaveJWT() error {
	jwt, err := utils.GetInput("Enter your Pinata JWT")
	if err != nil {
		return err
	}

	if jwt == "" {
		return errors.New("JWT cannot be empty")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	p := filepath.Join(home, ".pinata-files-cli")
	err = os.WriteFile(p, []byte(jwt), 0600)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://%s/data/testAuthentication", config.GetAPIHost())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+jwt)

	client := &http.Client{
		Timeout: time.Duration(time.Second * 3),
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	status := resp.StatusCode
	if status != 200 {
		return errors.New("Authentication failed, make sure you are using the Pinata JWT")
	}

	fmt.Println("Authentication Successful!")
	err = gateways.SetGateway("")
	if err != nil {
		return err
	}

	return nil
}

func GetHost() string {
	return GetEnv("PINATA_HOST", "api.pinata.cloud")
}

func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}
