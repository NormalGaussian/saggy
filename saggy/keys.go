package saggy

import (
	"encoding/json"
	"os"
	"strings"
)

type EncryptKeys struct {
	publicKeys         *map[string]string
	publicKeysFilepath string
}

type DecryptKey struct {
	privateKeyFilepath string
	privateKey         string
}

type GenerateKeys struct {
	publicKey  string
	privateKey string
}

type KeyFileNames struct {
	privateKeyFilepath string
	publicKeysFilepath string
}

type KeyFiles struct {
	privateKeyFile *os.File
	publicKeysFile *os.File
}

type Keys struct {
	*EncryptKeys
	*DecryptKey
	*GenerateKeys
}

func ReadPublicKeys(filepath string) (*map[string]string, error) {
	keys := make(map[string]string)

	// Open the file
	filedata_bytes, err := os.ReadFile(filepath)
	if err != nil && os.IsNotExist(err) {
		// No such file, return an empty map as there are no keys
		return &keys, nil
	} else if err != nil {
		// The file might exist, but for some other reason we can't open it
		return nil, NewSaggyError("Failed to open public keys file", err)
	}

	// If the file is empty, there are no keys to read
	if len(filedata_bytes) == 0 {
		return &keys, nil
	}

	// Read the keys from the file
	if err := json.NewDecoder(strings.NewReader(string(filedata_bytes))).Decode(&keys); err != nil {
		return nil, NewSaggyError("Failed to parse public keys file", err)
	}
	return &keys, nil
}

func ReadPrivateKey(file string) (string, error) {
	// Open the file
	filedata_bytes, err := os.ReadFile(file)
	if err != nil && os.IsNotExist(err) {
		// Forward the error if the file doesn't exist so the caller can handle it
		return "", err
	} else if err != nil {
		// The file might exist, but for some other reason we can't open it
		return "", NewSaggyError("Failed to open private key file", err)
	}

	// Read the key
	filedata_string := string(filedata_bytes)
	privateKey := ""
	for _, line := range strings.Split(filedata_string, "\n") {
		if strings.HasPrefix(line, "AGE-SECRET-KEY-") {
			privateKey = line
			break
		}
	}
	if privateKey == "" {
		return "", NewSaggyError("Failed to find the private key in the file", nil)
	}
	return privateKey, nil
}

func DecryptKeysFromFiles(privateKeyFilepath string) (*DecryptKey, error) {
	privateKey, err := ReadPrivateKey(privateKeyFilepath)
	if err != nil {
		return nil, err
	}
	return &DecryptKey{
		privateKeyFilepath: privateKeyFilepath,
		privateKey:         privateKey,
	}, nil
}

func EncryptKeysFromFiles(publicKeysFilepath string) (*EncryptKeys, error) {
	publicKeys, err := ReadPublicKeys(publicKeysFilepath)
	if err != nil {
		return nil, err
	}
	return &EncryptKeys{
		publicKeys:         publicKeys,
		publicKeysFilepath: publicKeysFilepath,
	}, nil
}

func KeysFromFiles(publicKeysFilepath, privateKeyFilepath string) (*Keys, error) {
	encryptKeys, err := EncryptKeysFromFiles(publicKeysFilepath)
	if err != nil {
		return nil, err
	}
	decryptKey, err := DecryptKeysFromFiles(privateKeyFilepath)
	if err != nil {
		return nil, err
	}
	return &Keys{
		EncryptKeys: encryptKeys,
		DecryptKey:  decryptKey,
	}, nil
}
