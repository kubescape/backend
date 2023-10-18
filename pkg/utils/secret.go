package utils

import (
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

// LoadTokenFromFile loads the token and the account id from a file
func LoadTokenFromFile(secretPath string) (*TokenSecretData, error) {
	token, err := os.ReadFile(secretPath + "/" + TokenSecretKey)
	if err != nil {
		return nil, err
	}
	account, err := os.ReadFile(secretPath + "/" + AccountSecretKey)
	if err != nil {
		return nil, err
	}

	return &TokenSecretData{
		Token:   string(token),
		Account: string(account),
	}, nil
}
