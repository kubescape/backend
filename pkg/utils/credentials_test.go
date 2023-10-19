package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadCredentialsFromFile(t *testing.T) {
	secretPath, err := os.MkdirTemp(t.TempDir(), "secret")
	if err != nil {
		t.Fatal(err)
	}

	// When account and access key are missing, return an error
	_, err = LoadCredentialsFromFile(secretPath)
	assert.Error(t, err)

	// create access key file and test that the token is loaded
	accessKey := "myAccessKey"
	err = os.WriteFile(secretPath+"/"+AccessKeySecretKey, []byte(accessKey), 0644)
	assert.NoError(t, err)
	credentials, err := LoadCredentialsFromFile(secretPath)
	assert.NoError(t, err)
	assert.Equal(t, accessKey, credentials.AccessKey)
	assert.Empty(t, credentials.Account)

	// create account file and test that the account is loaded
	account := "xxxxx-xxxxx"
	err = os.WriteFile(secretPath+"/"+AccountSecretKey, []byte(account), 0644)
	assert.NoError(t, err)
	credentials, err = LoadCredentialsFromFile(secretPath)
	assert.NoError(t, err)
	assert.Equal(t, accessKey, credentials.AccessKey)
	assert.Equal(t, account, credentials.Account)
}
