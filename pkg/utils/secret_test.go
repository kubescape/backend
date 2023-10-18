package utils

import (
	"os"
	"testing"
)

func TestLoadTokenFromFile(t *testing.T) {
	secretPath, err := os.MkdirTemp(t.TempDir(), "secret")
	if err != nil {
		t.Fatal(err)
	}
	token := "mySecretToken"
	err = os.WriteFile(secretPath+"/"+TokenSecretKey, []byte(token), 0644)
	if err != nil {
		t.Fatal(err)
	}
	account := "xxxxx-xxxxx"
	err = os.WriteFile(secretPath+"/"+AccountSecretKey, []byte(account), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test loading the token from the temporary file
	secretData, err := LoadTokenFromFile(secretPath)
	if err != nil {
		t.Fatalf("LoadTokenFromFile returned an error: %v", err)
	}

	if secretData.Token != token {
		t.Errorf("Expected token %s, but got %s", token, secretData.Token)
	}

	if secretData.Account != account {
		t.Errorf("Expected account %s, but got %s", account, secretData.Account)
	}
}
