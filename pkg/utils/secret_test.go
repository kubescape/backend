package utils

import (
	"os"
	"testing"
)

func TestLoadTokenFromSecret(t *testing.T) {
	secretPath, err := os.MkdirTemp(t.TempDir(), "secret")
	if err != nil {
		t.Fatal(err)
	}
	token := "mySecretToken"
	err = os.WriteFile(secretPath+"/"+TokenSecretKey, []byte(token), 0644)
	if err != nil {
		t.Fatal(err)
	}
	accountId := "xxxxx-xxxxx"
	err = os.WriteFile(secretPath+"/"+AccountIdSecretKey, []byte(accountId), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test loading the token from the temporary file
	secretData, err := LoadTokenFromSecret(secretPath)
	if err != nil {
		t.Fatalf("LoadTokenFromSecret returned an error: %v", err)
	}

	if secretData.Token != token {
		t.Errorf("Expected token %s, but got %s", token, secretData.Token)
	}

	if secretData.AccountId != accountId {
		t.Errorf("Expected account id %s, but got %s", accountId, secretData.AccountId)
	}
}
