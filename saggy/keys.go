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

func (encryptKeys *EncryptKeys) Read(filepath string) error {
	keys := make(map[string]string)

	// Open the file
	var filedata_string string
	filedata_bytes, err := os.ReadFile(filepath)
	if err != nil && os.IsNotExist(err) {
		// No such file, therefore no data to read
		filedata_string = ""
	} else if err != nil {
		// The file might exist, but for some other reason we can't open it
		return NewSaggyError("Failed to open public keys file", err)
	} else {
		filedata_string = string(filedata_bytes)
	}

	// If the file is empty, there are no keys to read
	if len(filedata_string) > 0 {
		// Read the keys from the file
		if err := json.NewDecoder(strings.NewReader(filedata_string)).Decode(&keys); err != nil {
			return NewSaggyError("Failed to parse public keys file", err)
		}
	}

	encryptKeys.publicKeys = &keys
	encryptKeys.publicKeysFilepath = filepath

	return nil
}

func (decryptKey *DecryptKey) Read(filepath string) error {
	// Open the file
	filedata_bytes, err := os.ReadFile(filepath)
	if err != nil && os.IsNotExist(err) {
		// Forward the error if the file doesn't exist so the caller can handle it
		return err
	} else if err != nil {
		// The file might exist, but for some other reason we can't open it
		return NewSaggyError("Failed to open private key file", err)
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
		return NewSaggyError("Failed to find the private key in the file", nil)
	}

	decryptKey.privateKeyFilepath = filepath
	decryptKey.privateKey = privateKey

	return nil
}

func DecryptKeysFromFile(privateKeyFilepath string) (*DecryptKey, error) {
	decryptKey := &DecryptKey{}
	if err := decryptKey.Read(privateKeyFilepath); err != nil {
		return nil, err
	}
	return decryptKey, nil
}

func EncryptKeysFromFile(publicKeysFilepath string) (*EncryptKeys, error) {
	encryptKeys := &EncryptKeys{}
	if err := encryptKeys.Read(publicKeysFilepath); err != nil {
		return nil, err
	}
	return encryptKeys, nil
}

func KeysFromFiles(publicKeysFilepath, privateKeyFilepath string) (*Keys, error) {
	encryptKeys, err := EncryptKeysFromFile(publicKeysFilepath)
	if err != nil {
		return nil, err
	}
	decryptKey, err := DecryptKeysFromFile(privateKeyFilepath)
	if err != nil {
		return nil, err
	}
	return &Keys{
		EncryptKeys: encryptKeys,
		DecryptKey:  decryptKey,
	}, nil
}

func (keyFiles *KeyFiles) Open(keyFileNames *KeyFileNames) error {
	privateKeyFile, err := os.Open(keyFileNames.privateKeyFilepath)
	if err != nil {
		return NewSaggyError("Failed to open private key file", err)
	}

	publicKeysFile, err := os.Open(keyFileNames.publicKeysFilepath)
	if err != nil {
		return NewSaggyError("Failed to open public keys file", err)
	}

	keyFiles.privateKeyFile = privateKeyFile
	keyFiles.publicKeysFile = publicKeysFile

	return nil
}
func (keyFiles *KeyFiles) Close() {
	if keyFiles.privateKeyFile != nil {
		keyFiles.privateKeyFile.Close()
	}
	if keyFiles.publicKeysFile != nil {
		keyFiles.publicKeysFile.Close()
	}
}
