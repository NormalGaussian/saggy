package saggy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	defaultSecretsDir     = "./secrets"
	secretsDir            = getEnv("SAGGY_SECRETS_DIR", defaultSecretsDir)
	defaultKeyFile        = filepath.Join(secretsDir, "age.key")
	keyFile               = getEnv("SAGGY_KEY_FILE", defaultKeyFile)
	defaultPublicKeysFile = filepath.Join(secretsDir, "public-age-keys.json")
	publicKeysFile        = getEnv("SAGGY_PUBLIC_KEYS_FILE", defaultPublicKeysFile)
	defaultKeyName        = strings.ToLower(getHostname())
	keyName               = getEnv("SAGGY_KEYNAME", defaultKeyName)
	sopsAgeKeyFile        = keyFile
	agePublicKeys         = getAgePublicKeys()
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

func getAgePublicKeys() string {
	if _, err := os.Stat(publicKeysFile); os.IsNotExist(err) {
		return ""
	}

	data, err := ioutil.ReadFile(publicKeysFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read public keys file:", err)
		os.Exit(1)
	}

	var keys map[string]string
	if err := json.Unmarshal(data, &keys); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse public keys file:", err)
		os.Exit(1)
	}

	args := []string{"--age"}
	for _, key := range keys {
		args = append(args, key)
	}

	return strings.Join(args, " ")
}
