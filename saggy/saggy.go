package saggy

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	secretsDir     = getEnv("SAGGY_SECRETS_DIR", "./secrets")
	keyFile        = getEnv("SAGGY_KEY_FILE", filepath.Join(secretsDir, "age.key"))
	publicKeysFile = getEnv("SAGGY_PUBLIC_KEYS_FILE", filepath.Join(secretsDir, "public-age-keys.json"))
	keyName        = getEnv("SAGGY_KEYNAME", strings.ToLower(getHostname()))
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get hostname:", err)
		os.Exit(1)
	}
	return hostname
}

func getAgePublicKeys() ([]string, error) {
	if _, err := os.Stat(publicKeysFile); errors.Is(err, os.ErrNotExist) {
		return []string{}, nil
	}

	data, err := os.ReadFile(publicKeysFile)
	if err != nil {
		return nil, NewSaggyError("Failed to read public keys file", err)
	}

	var keys map[string]string
	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, NewSaggyError("Failed to parse public keys file", err)
	}

	publicKeys := []string{}
	for _, value := range keys {
		publicKeys = append(publicKeys, value)
	}

	return publicKeys, nil
}
