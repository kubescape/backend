package utils

import (
	"fmt"
	"os"
)

const (
	TokenSecretKey   = "token"
	AccountSecretKey = "account"
)

type TokenSecretData struct {
	Account string `mapstructure:"account"`
	Token   string `mapstructure:"token"`
}

// LoadTokenFromFile loads the token and the account from a file
func LoadTokenFromFile(secretPath string) (*TokenSecretData, error) {
	tokenPath := secretPath + "/" + TokenSecretKey
	accountPath := secretPath + "/" + AccountSecretKey

	// if both files are missing, return an error as we need at least one of them
	_, errTokenPath := os.Stat(tokenPath)
	_, errAccountPath := os.Stat(accountPath)
	if os.IsNotExist(errTokenPath) && os.IsNotExist(errAccountPath) {
		return nil, fmt.Errorf("token and account files are missing in path %s", secretPath)
	}

	t := &TokenSecretData{}
	if token, err := os.ReadFile(tokenPath); err == nil {
		t.Token = string(token)
	}
	if account, err := os.ReadFile(accountPath); err == nil {
		t.Account = string(account)
	}

	return t, nil
}
