package saggy

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var (
	secretsDir             = getEnv("SAGGY_SECRETS_DIR", "./secrets")
	keyFile                = getEnv("SAGGY_KEY_FILE", filepath.Join(secretsDir, "age.key"))
	publicKeysFile         = getEnv("SAGGY_PUBLIC_KEYS_FILE", filepath.Join(secretsDir, "public-age-keys.json"))
	keyName                = getEnv("SAGGY_KEYNAME", strings.ToLower(getHostname()))
	useBundledDependencies = getEnv("SAGGY_USE_BUNDLED_DEPENDENCIES", "false") == "true"
)

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
