package utils

import (
	"os"
)

const (
	TokenSecretKey     = "token"
	AccountIdSecretKey = "account-id"
)

type TokenSecretData struct {
	AccountId string `mapstructure:"accountId"`
	Token     string `mapstructure:"token"`
}

// LoadTokenFromSecret loads the token and the account id from a loaded secret path
func LoadTokenFromSecret(secretPath string) (*TokenSecretData, error) {
	token, err := os.ReadFile(secretPath + "/" + TokenSecretKey)
	if err != nil {
		return nil, err
	}
	accountID, err := os.ReadFile(secretPath + "/" + AccountIdSecretKey)
	if err != nil {
		return nil, err
	}

	return &TokenSecretData{
		Token:     string(token),
		AccountId: string(accountID),
	}, nil
}
