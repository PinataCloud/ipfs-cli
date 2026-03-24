package agents

import (
	"fmt"
	"net/http"
	"pinata/internal/utils"
)

// CredentialLogin prompts the user for a credential and stores it as a secret.
func CredentialLogin(prompt, secretName string) error {
	key, err := utils.GetInput(prompt, prompt)
	if err != nil {
		return fmt.Errorf("failed to read credential: %w", err)
	}
	if key == "" {
		return fmt.Errorf("credential cannot be empty")
	}

	fmt.Printf("Creating secret '%s'...\n", secretName)
	err = UpsertSecret(secretName, key)
	return err
}


// upsertSecret creates or updates a secret by name
func UpsertSecret(name, value string) error {
	var list SecretListResponse
	if err := doSecretsJSON(http.MethodGet, "", nil, &list); err != nil {
		return fmt.Errorf("failed to list secrets: %w", err)
	}
	for _, s := range list.Secrets {
		if s.Name == name {
			return doSecretsJSON(http.MethodPut, "/"+s.ID, UpdateSecretBody{Value: value}, nil)
		}
	}
	
	return doSecretsJSON(http.MethodPost, "", CreateSecretBody{Name: name, Value: value}, nil)
}
