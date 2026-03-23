package agents

import (
	"fmt"
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
	_, err = CreateSecret(secretName, key)
	return err
}
