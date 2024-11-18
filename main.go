package main

import (
	"errors"
	"fmt"
	"os"
	"saggy"
	_ "embed"
	"path/filepath"
)

var (
	secretsDir     			= getEnv("SAGGY_SECRETS_DIR", "./secrets")
	keyFile        			= getEnv("SAGGY_KEY_FILE", filepath.Join(secretsDir, "age.key"))
	publicKeysFile 			= getEnv("SAGGY_PUBLIC_KEYS_FILE", filepath.Join(secretsDir, "public-age-keys.json"))
)

type CLIError struct {
	Code       int
	Message    string
	PrintUsage bool
	Err        error
}

func (e *CLIError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	} else {
		return e.Message
	}
}

func NewCLIError(code int, message string, err error, printUsage bool) *CLIError {
	return &CLIError{Code: code, Message: message, Err: err, PrintUsage: printUsage}
}

//go:embed LICENSE
var LICENSE_SAGGY string
//go:embed LICENSE_AGE
var LICENSE_AGE string
//go:embed LICENSE_SOPS
var LICENSE_SOPS string

func cli(argv []string) error {
	if len(argv) < 2 {
		printUsage()
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
		return saggy.Encrypt(source, destination)
	case "decrypt":
		if len(args) < 1 {
			return NewCLIError(1, "Nothing provided to decrypt", nil, true)
		}
		source := args[0]
		destination := ""
		if len(args) > 1 {
			destination = args[1]
		}
		return saggy.Decrypt(source, destination)
	case "keygen":
		if len(args) > 0 {
			if args[0] == "-" {
				return saggy.KeygenToStdout("age")
			}
		}
		return saggy.KeygenToFile(keyFile, publicKeysFile)
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
		return saggy.With(target, command, mode)
	case "license":
		if len(args) >= 1 && args[0] == "--full" {
			fmt.Println("The saggy source is primarily licensed under the BSD-3-Clause license; its dependencies are licensed under their respective licenses, listed as follows:")
			fmt.Println("")
			fmt.Println("Saggy is licensed under the BSD-3-Clause license:")
			fmt.Println(LICENSE_SAGGY)
			fmt.Println("")
			fmt.Println("parts of the age project is bundled, and is licensed under the BSD-3-Clause license:")
			fmt.Println(LICENSE_AGE)
			fmt.Println("")
			fmt.Println("parts of the sops project is bundled, and is licensed under the MPL 2.0 license:")
			fmt.Println(LICENSE_SOPS)
			return nil
		}
		fmt.Println("The saggy source is primarily licensed under the BSD-3-Clause license; its dependencies are licensed under their respective licenses, listed as follows:")
		fmt.Println("")
		fmt.Println("Saggy is licensed under the BSD-3-Clause license")
		fmt.Println("parts of the age project is bundled, and is licensed under the BSD-3-Clause license")
		fmt.Println("parts of the sops project is bundled, and is licensed under the MPL 2.0 license")
		fmt.Println("")
		fmt.Println("to see the full license text, run `saggy license --full`")
		return nil
	default:
		return NewCLIError(1, "Unknown command: "+cmd, nil, true)
	}
}

func main() {
	// Invoke the CLI
	if err := cli(os.Args); err != nil {
		var SilentError *saggy.SilentError
		if errors.As(err, &SilentError) {
			os.Exit(SilentError.ExitCode)
		}
		var cliErr *CLIError
		if errors.As(err, &cliErr) {
			fmt.Fprintln(os.Stderr, err.Error())
			if cliErr.PrintUsage {
				printUsage()
			}
			os.Exit(cliErr.Code)
		}
		saggyError := &saggy.SaggyError{}
		if errors.As(err, &saggyError) {
			fmt.Fprintln(os.Stderr, saggyError.Error())
			os.Exit(2)
		}

		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
	os.Exit(0)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage:

  saggy keygen
	 - Generate a new key and add it to the public keys file

  saggy with <target> [-w] -- <command>
	 - Run the command with the target decrypted
	   The target is decrypted and into a temporary file or folder
	   Any {} in the command is replaced with the temporary file or folder
	   If the -w flag is provided, changes to the decrypted file or folder are encrypted again
	   Otherwise, the decrypted file or folder is deleted and changes are not preserved
  
  saggy encrypt <target>
	 - Encrypt the target, storing the result in a file with the same name but with a .sops pre-suffix
	   e.g myfile.yaml -> myfile.sops.yaml.
		   myfile -> myfile.sops

  saggy encrypt <target> <destination>
	 - Encrypt the target, storing the result in the destination file

  saggy decrypt <target>
	 - Decrypt the target, storing the result in a file with the same name but without a .sops pre-suffix
	   e.g myfile.sops.yaml -> myfile.yaml.
		   myfile.sops -> myfile

  saggy decrypt <target> <destination>
	 - Decrypt the target, storing the result in the destination file

Environment Variables:
  SAGGY_SECRETS_DIR       - the directory containing the secrets
							(default: ./secrets)
  SAGGY_KEY_FILE          - the file containing the AGE key
							(default: $SAGGY_SECRETS_DIR/age.key)
  SAGGY_PUBLIC_KEYS_FILE  - the json file containing the public keys
							(default: $SAGGY_SECRETS_DIR/public-age-keys.json)
  SAGGY_KEYNAME           - the name with which to save the public key when using keygen
							(default: the lowercased hostname)`)
}
