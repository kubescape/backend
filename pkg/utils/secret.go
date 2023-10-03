package utils

import (
	"encoding/base64"
	"os"
)

type SecretData struct {
	Token string `mapstructure:"token"`
}

// LoadSecret loads a token from a secret file and decodes it from base64
func LoadTokenFromSecret(path string) (*SecretData, error) {
	secretBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	token, err := decodeBase64(secretBytes)
	if err != nil {
		return nil, err
	}

	return &SecretData{
		Token: token,
	}, nil
}

func decodeBase64(data []byte) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
