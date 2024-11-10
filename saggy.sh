#!/bin/bash
#
# Saggy - GitOps secrets management using SOPS and AGE
#
# Assumption:
#  - There is a folder called "secrets" in the same directory as this script
#  - The "secrets" folder contains the following files:
#    - "age.key" - the private key for AGE; this can be generated using `saggy keygen`
#    - "public-age-keys.json" - a JSON file containing the public keys for age. This is automatically generated when you run `saggy keygen`
#  - There are two folders called "data.decrypted" and "data.encrypted" in the same directory as this script
#  - The "data.decrypted" folder contains a .gitignore file to ignore all of the decrypted files
#  - The "data.encrypted" folder contains a .gitignore file and all of the files that are encrypted using SOPS & age
#   
#  Prefer *not* using the decrypt or encrypt commands. Instead, use the "with" command to run a command on the decrypted files.
#  The "with" command will decrypt the files, run the command, and then delete or encrypt the files.

if ! which age 2>&1 >/dev/null; then
    echo "age is not installed. Please install age." >&2
    exit 1
fi
if ! which sops 2>&1 >/dev/null; then
    echo "sops is not installed. Please install sops." >&2
    exit 1
fi

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

DEFAULT_SAGGY_SECRETS_DIR="./secrets"
SAGGY_SECRETS_DIR="${SAGGY_SECRETS_DIR:-$DEFAULT_SAGGY_SECRETS_DIR}"

DEFAULT_SAGGY_KEY_FILE="$SAGGY_SECRETS_DIR/age.key"
SAGGY_KEY_FILE="${SAGGY_KEY_FILE:-$DEFAULT_SAGGY_KEY_FILE}"

DEFAULT_SAGGY_PUBLIC_KEYS_FILE="$SAGGY_SECRETS_DIR/public-age-keys.json"
SAGGY_PUBLIC_KEYS_FILE="${SAGGY_PUBLIC_KEYS_FILE:-$DEFAULT_SAGGY_PUBLIC_KEYS_FILE}"

export SOPS_AGE_KEY_FILE="$SAGGY_KEY_FILE"
export AGE_PUBLIC_KEYS="$(jq -r '["--age", .[]] | join(" ")' "$PUBLIC_KEYFILE")"

is_sopsified_filename() {
    FILE="$1"
    BASENAME="$(basename "$FILE")"
    return [[ "$BASENAME" == *.sops.* ]] || [[ "$BASENAME" == *.sops ]]
}

is_sopsified_dirname() {
    DIR="$1"
    BASENAME="$(basename "$DIR")"
    return [[ "$BASENAME" == *.sops ]]
}

get_sopsified_filename() {
    FILE="$1"
    BASENAME="$(basename "$FILE")"
    if [[ "$BASENAME" == *.* ]]; then
        SUFFIX="${BASENAME##*.}"
        echo "${FILE%.*}.sops.$SUFFIX"
    else
        echo "$FILE.sops"
    fi
}

get_sopsified_dirname() {
    DIR="$1"
    echo "$DIR.sops"
}

unsopsify_filename() {
    FILE="$1"
    BASENAME="$(basename "$FILE")"
    if [[ "$BASENAME" == *.sops.* ]]; then
        SUFFIX="${BASENAME##*.}"
        echo "${FILE%.sops.*}.$SUFFIX"
    elif [[ "$BASENAME" == *.sops ]]; then
        echo "${FILE%.sops}"
    else
        # TODO: or raise an error?
        echo "$FILE"
    fi
}

encrypt_file() {
    FROM="$1"
    TO="$2"

    if is_sopsified_filename "$FROM"; then
        echo "File is already encrypted: $FROM"
        exit 1
    fi

    if [[ -z "$TO" ]]; then
        TO="$(get_sopsified_filename "$FROM")"
    fi

    sops --encrypt $AGE_PUBLIC_KEYS "$FROM" > "$TO"
}

encrypt_folder() {
    FROM="$1"
    TO="$2"

    FROM="$(end_with_slash "$FROM")"
    TO="$(end_with_slash "$TO")"

    if [[ -z "$TO" ]]; then
        TO="$(get_sopsified_dirname "$FROM")"
    fi
    
    echo "Encrypting files in $FROM and saving to $TO:"
    find "$FROM" -type f | while read -r RAW_DECRYPTED_FILE; do
        DECRYPTED_FILE="${RAW_DECRYPTED_FILE#$FROM}"

        # If already a sops encrypted file, skip
        if is_sopsified_filename "$DECRYPTED_FILE"; then
            # n.b. this is only a very shallow check, and hasn't checked the contents of the file
            echo "File is already encrypted: $DECRYPTED_FILE"
            # TODO: error code?
            continue
        fi

        ENCRYPTED_FILE="$(get_sopsified_filename "$DECRYPTED_FILE")"

        echo -e "\t$DECRYPTED_FILE"

        mkdir -p "$(dirname "$TO$ENCRYPTED_FILE")"

        sops --encrypt $AGE_PUBLIC_KEYS "$FROM$DECRYPTED_FILE" > "$TO$ENCRYPTED_FILE"
    done

}

end_with_slash() {
    if [[ "$1" != */ ]]; then
        echo "$1/"
    else
        echo "$1"
    fi
}

decrypt_folder() {
    FROM="$(end_with_slash "$1")"
    TO="$(end_with_slash "$2")"

    echo "Decrypting files in $FROM and saving to $TO:" 
    find "$FROM" -type f | while read -r RAW_ENCRYPTED_FILE; do
        ENCRYPTED_FILE="${RAW_ENCRYPTED_FILE#$FROM}"

        # If not a sops file, skip
        if ! is_sopsified_filename "$ENCRYPTED_FILE"; then
            # n.b. this is only a very shallow check, and hasn't checked the contents of the file
            continue
        fi

        DECRYPTED_FILE="$(unsopsify_filename "$ENCRYPTED_FILE")"

        echo -e "\t$DECRYPTED_FILE"

        mkdir -p "$(dirname "$TO$ENCRYPTED_FILE")"

        sops --decrypt "$FROM$ENCRYPTED_FILE" > "$TO$DECRYPTED_FILE"
    done
}

with_file() {
    FILE="$1"
    COMMAND="$2"
    MODE="$3"

    if [[ ! -e "$FILE" ]]; then
        echo "File does not exist: $FILE"
        exit 1
    fi

    TMP_FILE="$(mktemp)"
    trap "rm -f $TMP_FILE" EXIT

    sops --decrypt "$FILE" > "$TMP_FILE"

    eval "$COMMAND"

    if [[ "$MODE" == "write" ]]; then
        sops --encrypt $AGE_PUBLIC_KEYS "$TMP_FILE" > "$FILE"
    fi
}

## "encrypt with <file or folder> -- <command>"
## e.g. "encrypt with herd-1 -- talosctl apply {}/talos/controlplane.yaml"
with() {
    FILE_OR_FOLDER="$1"
    shift
    MODE="read"
    if [[ -n "$1" ]] && [[ "$1" != "--" ]]; then
        if [[ "$2" == "-r" ]] || [[ "$2" == "--read" ]]; then
            MODE="read"
        elif [[ "$2" == "-w" ]] || [[ "$2" == "--write" ]]; then
            MODE="write"
        fi
        shift
    fi

    # Extract the command, which is everything after "--"
    while [[ -n "$1" ]] && [[ "$1" != "--" ]]; do
        shift
    done
    shift
    COMMAND="$@"
    if [[ -z "$COMMAND" ]]; then
        echo "No command provided"
        exit 1
    fi

    if [[ ! -e "$FILE_OR_FOLDER" ]]; then
        echo "File or folder does not exist: $FILE_OR_FOLDER"
        exit 1
    fi

    FOLDER=""
    FILE=""

    if [[ -d "$FILE_OR_FOLDER" ]]; then
        FOLDER="$FILE_OR_FOLDER"
    elif [[ -f "$FILE_OR_FOLDER" ]]; then
        FILE="$FILE_OR_FOLDER"
    else
        echo "path must be a file or a folder - not some other device: $FILE_OR_FOLDER"
        exit 1
    fi

    if [[ -n "$FOLDER" ]]; then
        # With folder
        
        # Create the temporary folder and ensure it is deleted
        TMP_FOLDER="$(mktemp -d)"
        mkdir -p "$TMP_FOLDER"
        trap "rm -rf $TMP_FOLDER" EXIT

        # Replace the {} with the folder
        COMMAND="${COMMAND//\{\}/$TMP_FOLDER}"

        # Decrypt the folder    
        decrypt_folder "$FOLDER" "$TMP_FOLDER"
        
        # Run the command
        eval "$COMMAND"

        # If mode is "write", then we want to save the changes
        # TODO: handle deleted files
        if [[ "$MODE" == "write" ]]; then
            encrypt_folder "$TMP_FOLDER" "$FOLDER"
        fi

    else
        # With file

        # Create the temporary file and ensure it is deleted
        TMP_FILE="$(mktemp)"
        trap "rm -f $TMP_FILE" EXIT
        
        # Replace the {} with the file
        COMMAND="${COMMAND//\{\}/$TMP_FILE}"

        # Decrypt the file
        sops --decrypt "$FILE" > "$TMP_FILE"

        # Run the command
        eval "$COMMAND"

        # If mode is "write", then we want to save the changes
        if [[ "$MODE" == "write" ]]; then
            sops --encrypt $AGE_PUBLIC_KEYS "$TMP_FILE" > "$FILE"
        fi
    fi
}

cmd="$1"
shift
case "$cmd" in
    encrypt)
        TARGET="$1"
        DESTINATION="$2"

        if [[ -z "$TARGET" ]]; then
            echo "Nothing provided to encrypt"
            exit 1
        fi

        if [[ -f "$TARGET" ]]; then

            encrypt_file "$TARGET" "$DESTINATION"

        elif [[ -d "$TARGET" ]]; then

            encrypt_folder "$TARGET" "$TARGET.sops"

        else
            echo "Path must be a file or a folder - not some other device: $TARGET"
            exit 1
        fi
        ;;
    decrypt)
        TARGET="$1"
        if [[ -z "$TARGET" ]]; then
            echo "Nothing provided to decrypt"
            exit 1
        fi
        if [[ -f "$TARGET" ]]; then
            sops --decrypt "$TARGET" > "${TARGET%.sops}"
        elif [[ -d "$TARGET" ]]; then
            decrypt_folder "$TARGET" "${TARGET%.sops}"
        else
            echo "Path must be a file or a folder - not some other device: $TARGET"
            exit 1
        fi
        ;;
    keygen)
        if [[ -e "$SCRIPT_DIR/secrets/age.key" ]]; then
            echo "Key already exists - to generate a new key, delete the existing key"
            echo "1. Decrypt the folders"
            echo "  $0 decrypt"
            echo "2. Delete the key"
            echo "  rm \"./secrets/age.key\""
            echo "2. Delete it from the public keys file"
            echo "  vi \"./secrets/public-age-keys.json\""
            echo "3. Run this command again"
            echo "  $0 keygen"
            echo "4. Encrypt the folders"
            echo "  $0 encrypt"
            # TODO: add command "rotate" to rotate the key
            exit 1
        fi

        # Create the key
        age-keygen -o "$SAGGY_KEY_FILE"
        PUBLIC_KEY="$(age-keygen -y "$SAGGY_KEY_FILE")"
        
        # Use the hostname as the name for the key
        HOSTNAME="$(hostname)"
        HOSTNAME="${HOSTNAME,,}"

        # Add the public key to the public keys file
        cat "$SAGGY_PUBLIC_KEYS_FILE" | jq ". + {\"${HOSTNAME,,}\": \"$PUBLIC_KEY\"}" > "$SAGGY_PUBLIC_KEYS_FILE.tmp"
        mv "$SAGGY_PUBLIC_KEYS_FILE.tmp" "$SAGGY_PUBLIC_KEYS_FILE"
        ;;

    with)
        with "$@"
        ;;

    *)
        echo "Usage:"
        echo "  $0 keygen"
        echo "     - Generate a new key and add it to the public keys file"
        echo "  $0 with <target> [-w] -- <command>"
        echo "     - Run the command with the target decrypted"
        echo "       The target is decrypted and into a temporary file or folder"
        echo "       Any {} in the command is replaced with the temporary file or folder"
        echo "       If the -w flag is provided, changes to the decrypted file or folder are encrypted again"
        echo "       Otherwise, the decrypted file or folder is deleted and changes are not preserved"
        echo "  $0 encrypt <target>"

        echo "  $0 decrypt <target>"
        echo ""
        echo "Environment Variables:"
        echo "  SAGGY_SECRETS_DIR       - the directory containing the secrets"
        echo "                            (default: ./secrets)"
        echo "  SAGGY_KEY_FILE          - the file containing the AGE key"
        echo "                            (default: \$SAGGY_SECRETS_DIR/age.key)"
        echo "  SAGGY_PUBLIC_KEYS_FILE  - the json file containing the public keys"
        echo "                            (default: \$SAGGY_SECRETS_DIR/public-age-keys.json)"
        exit 1
        ;;
esac
