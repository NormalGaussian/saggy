package saggy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	useBundledDependencies = getEnv("SAGGY_USE_BUNDLED_DEPENDENCIES", "false") == "true"
)

func CLI(argv []string) error {
	var (
		secretsDir     = getEnv("SAGGY_SECRETS_DIR", "./secrets")
		privateKeyFile = getEnv("SAGGY_KEY_FILE", filepath.Join(secretsDir, "age.key"))
		publicKeysFile = getEnv("SAGGY_PUBLIC_KEYS_FILE", filepath.Join(secretsDir, "public-age-keys.json"))
		keyName        = getEnv("SAGGY_KEYNAME", strings.ToLower(getHostname()))
	)

	if len(argv) < 2 {
		fmt.Fprintln(os.Stderr, USAGE_TEXT)
		return NewCLIError(1, "No command provided", nil, false)
	}

	cmd := argv[1]
	args := argv[2:]

	switch cmd {
	case "encrypt":

		if len(args) < 1 {
			return NewCLIError(1, "Nothing provided to encrypt", nil, true)
		}
		source := args[0]
		destination := ""
		if len(args) > 1 {
			destination = args[1]
		}

		keys, err := EncryptKeysFromFiles(publicKeysFile)
		if err != nil {
			return err
		}

		return Encrypt(keys, source, destination)

	case "decrypt":
		if len(args) < 1 {
			return NewCLIError(1, "Nothing provided to decrypt", nil, true)
		}
		source := args[0]
		destination := ""
		if len(args) > 1 {
			destination = args[1]
		}

		keys, err := DecryptKeysFromFiles(privateKeyFile)
		if err != nil {
			return err
		}

		return Decrypt(keys, source, destination)

	case "keygen":
		if len(args) > 0 {
			if args[0] == "-" {
				return KeygenToStdout("age")
			}
		}
		keyFileNames := &KeyFileNames{
			privateKeyFilepath: privateKeyFile,
			publicKeysFilepath: publicKeysFile,
		}
		return KeygenToFile(keyFileNames, keyName)

	case "with":
		if len(args) < 2 {
			return NewCLIError(1, "Usage: with <target> [-w] -- <command>", nil, true)
		}
		target := args[0]
		mode := "read"
		commandIndex := 1
		if args[1] == "-w" {
			mode = "write"
			commandIndex = 2
		}
		command := args[commandIndex+1:]

		keys, err := KeysFromFiles(publicKeysFile, privateKeyFile)
		if err != nil {
			return err
		}

		return With(keys, target, command, mode)

	case "license":
		if len(args) >= 1 && args[0] == "--full" {
			fmt.Println(LICENSE_TEXT_FULL)
		} else {
			fmt.Println(LICENSE_TEXT)
		}
		return nil

	default:
		return NewCLIError(1, "Unknown command: "+cmd, nil, true)
	}
}
