#!/bin/bash
set -euo pipefail

# Function to securely prompt for the decryption password
get_password() {
  local password
  printf "Enter password to decrypt secret keys: " > /dev/tty
  read -rs password </dev/tty
  printf "\n" > /dev/tty
  if [ -z "$password" ]; then
    return 1
  fi
  echo "$password"
}

# Prompt user for password and store it in PASSWORD variable
if ! PASSWORD=$(get_password); then
  echo "Password cannot be empty." > /dev/tty
  exit 1
fi

# Save password to file
echo "Creating secure password file..."
echo "$PASSWORD" | sudo tee /etc/ixios-validator-password > /dev/null
sudo chmod 600 /etc/ixios-validator-password
sudo chown root:root /etc/ixios-validator-password
echo "Password file created securely at /etc/ixios-validator-password"

# Determine the encrypted secret keys file by checking the three possible locations
files=(./validator_secretkeys_*.tar.enc)
if [ -e "${files[0]}" ]; then
  SECRET_KEYS_ENC="${files[0]}"
else
  files=(~/validator_secretkeys_*.tar.enc)
  if [ -e "${files[0]}" ]; then
    SECRET_KEYS_ENC="${files[0]}"
  else
    files=(/tmp/ixios-secure-keygen/validator_secretkeys_*.tar.enc)
    if [ -e "${files[0]}" ]; then
      SECRET_KEYS_ENC="${files[0]}"
    else
      echo "ERROR: No encrypted secret keys file found in the expected locations."
      exit 1
    fi
  fi
fi

# Verify the encrypted file exists and is not empty
if [ ! -f "$SECRET_KEYS_ENC" ] || [ ! -s "$SECRET_KEYS_ENC" ]; then
  echo "ERROR: Encrypted secret keys file does not exist or is empty: $SECRET_KEYS_ENC"
  exit 1
fi

# Create a secure temporary directory.
TEMP_DIR=$(mktemp -d)
if [ ! -d "$TEMP_DIR" ]; then
  echo "ERROR: Failed to create a temporary directory."
  exit 1
fi
chmod 700 "$TEMP_DIR"

# Define temporary file for the decrypted tar archive
TEMP_TAR="${TEMP_DIR}/secret_keys.tar"

# Decrypt the secret keys archive using the provided password
echo "Decrypting private keys..."
if ! openssl_output=$(openssl enc -d -aes256 -salt -pbkdf2 -iter 100000 -pass "pass:$PASSWORD" \
  -in "$SECRET_KEYS_ENC" -out "$TEMP_TAR" 2>&1); then
  echo "ERROR: Decryption failed. Incorrect password."
  echo "$openssl_output"
  rm -rf "$TEMP_DIR"
  exit 1
fi

# Verify the decrypted tar file exists and is not empty
if [ ! -f "$TEMP_TAR" ] || [ ! -s "$TEMP_TAR" ]; then
  echo "ERROR: Decrypted tar file not found or is empty."
  rm -rf "$TEMP_DIR"
  exit 1
fi

# Extract the decrypted tar archive
OUTPUT_DIR="./secret_keys"
mkdir -p "$OUTPUT_DIR"
echo "Extracting secret keys to ${OUTPUT_DIR}..."
if ! tar -xvf "$TEMP_TAR" -C "$OUTPUT_DIR"; then
  echo "ERROR: Extraction of secret keys failed."
  rm -rf "$TEMP_DIR"
  exit 1
fi

echo "Secret keys successfully decrypted and extracted to ${OUTPUT_DIR}."

# Cleanup temporary directory
rm -rf "$TEMP_DIR"

# Import secp256k1 key
SECP256K1_PRIVATE_KEY=$(cat ./secret_keys/tmp/ixios-secure-keygen/ixios-validator-keygen/secp256k1_secret.key)
SECP256K1_PRIVATE_KEY="${SECP256K1_PRIVATE_KEY#0x}"

# Create temporary files for the import process
echo "$PASSWORD" > ./password_file.temp
echo "$SECP256K1_PRIVATE_KEY" > ./key_file.temp

# Import just the secp256k1 key for now
ixiosSpark --aetherBloom account import --password ./password_file.temp ./key_file.temp # aetherbloom testnet
IMPORT_OUTPUT=$(ixiosSpark account import --password ./password_file.temp ./key_file.temp) # mainnet
VALIDATOR_ADDRESS=$(echo "$IMPORT_OUTPUT" | grep -o '{[0-9a-fA-F]\{64\}}')
if [ -n "$VALIDATOR_ADDRESS" ]; then
    echo "Saving validator address to /etc/ixios-validator-address..."
    echo "$VALIDATOR_ADDRESS" | sudo tee /etc/ixios-validator-address > /dev/null
    sudo chmod 600 /etc/ixios-validator-address
    sudo chown root:root /etc/ixios-validator-address
    echo "Validator address saved to /etc/ixios-validator-address"
else
    echo "Error: Could not extract validator address from import output."
fi

# Final Cleanup
shred -u ./password_file.temp
shred -u ./key_file.temp
shred ./secret_keys/tmp/ixios-secure-keygen/ixios-validator-keygen/secp256k1_secret.key
shred ./secret_keys/tmp/ixios-secure-keygen/ixios-validator-keygen/dilithium5_secret.key
shred ./secret_keys/tmp/ixios-secure-keygen/ixios-validator-keygen/sphincs_secret.key
rm -rf ./secret_keys