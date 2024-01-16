package utils

import (
	"fmt"
	"os"
	"strings"
)

const (
	AccessKeySecretKey = "accessKey"
	AccountSecretKey   = "account"
)

type Credentials struct {
	Account   string `mapstructure:"account"`
	AccessKey string `mapstructure:"accessKey"`
}

// LoadTokenFromFile loads the access key and the account from a mounted secret directory
func LoadCredentialsFromFile(secretPath string) (*Credentials, error) {
	accessKeyPath := secretPath + "/" + AccessKeySecretKey
	accountPath := secretPath + "/" + AccountSecretKey

	// if both files are missing, return an error as we need at least one of them
	_, errAccessKeyPath := os.Stat(accessKeyPath)
	_, errAccountPath := os.Stat(accountPath)
	if os.IsNotExist(errAccessKeyPath) && os.IsNotExist(errAccountPath) {
		return nil, fmt.Errorf("access key and account files are missing in path %s", secretPath)
	}

	t := &Credentials{}
	if accessKey, err := os.ReadFile(accessKeyPath); err == nil {
		t.AccessKey = strings.TrimSuffix(string(accessKey), "\n")
	}
	if account, err := os.ReadFile(accountPath); err == nil {
		t.Account = strings.TrimSuffix(string(account), "\n")
	}

	return t, nil
}
