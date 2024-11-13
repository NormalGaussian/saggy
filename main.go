
package main

import (
	"fmt"
	"os"
	"strings"
	"saggy"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "encrypt":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "Nothing provided to encrypt")
			os.Exit(1)
		}
		source := args[0]
		destination := ""
		if len(args) > 1 {
			destination = args[1]
		}
		saggy.Encrypt(source, destination)
	case "decrypt":
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "Nothing provided to decrypt")
			os.Exit(1)
		}
		source := args[0]
		destination := ""
		if len(args) > 1 {
			destination = args[1]
		}
		saggy.Decrypt(source, destination)
	case "keygen":
		saggy.Keygen()
	case "with":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: with <target> [-w] -- <command>")
			os.Exit(1)
		}
		target := args[0]
		mode := "read"
		commandIndex := 1
		if args[1] == "-w" {
			mode = "write"
			commandIndex = 2
		}
		command := strings.Join(args[commandIndex+1:], " ")
		saggy.With(target, command, mode)
	default:
		printUsage()
		os.Exit(1)
	}
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
							(default: the hostname)`)
}