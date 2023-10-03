package utils

import (
	"encoding/base64"
	"os"
	"testing"
)

func TestLoadTokenFromSecret(t *testing.T) {
	// Create a temporary file with a base64-encoded token
	tmpfile, err := os.CreateTemp(t.TempDir(), "secret-test.txt")
	if err != nil {
		t.Fatal(err)
	}

	token := "mySecretToken"
	encodedToken := base64.StdEncoding.EncodeToString([]byte(token))
	err = os.WriteFile(tmpfile.Name(), []byte(encodedToken), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test loading the token from the temporary file
	secretData, err := LoadTokenFromSecret(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadTokenFromSecret returned an error: %v", err)
	}

	if secretData.Token != token {
		t.Errorf("Expected token %s, but got %s", token, secretData.Token)
	}
}
