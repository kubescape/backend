package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadTokenFromFile(t *testing.T) {
	secretPath, err := os.MkdirTemp(t.TempDir(), "secret")
	if err != nil {
		t.Fatal(err)
	}

	// When account and token are missing, return an error
	_, err = LoadTokenFromFile(secretPath)
	assert.Error(t, err)

	// create token file and test that the token is loaded
	token := "mySecretToken"
	err = os.WriteFile(secretPath+"/"+TokenSecretKey, []byte(token), 0644)
	assert.NoError(t, err)
	secretData, err := LoadTokenFromFile(secretPath)
	assert.NoError(t, err)
	assert.Equal(t, token, secretData.Token)
	assert.Empty(t, secretData.Account)

	// create account file and test that the account is loaded
	account := "xxxxx-xxxxx"
	err = os.WriteFile(secretPath+"/"+AccountSecretKey, []byte(account), 0644)
	assert.NoError(t, err)
	secretData, err = LoadTokenFromFile(secretPath)
	assert.NoError(t, err)
	assert.Equal(t, token, secretData.Token)
	assert.Equal(t, account, secretData.Account)
}
