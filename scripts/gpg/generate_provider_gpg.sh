#!/usr/bin/env bash
set -euo pipefail

# Generates a dedicated GPG key for Terraform provider releases
# and exports ASCII-armored public/private keys into
#   .gpg/
#
# Defaults are safe for CI: RSA4096, no passphrase (non-interactive signing).
# If you want a passphrase, set GPG_PASSPHRASE env var and export will prompt once.

NAME_DEFAULT="DocSpring Terraform Provider"
EMAIL_DEFAULT="docspring-bot@docspring.com"
KEY_DIR=".gpg"

NAME="${GPG_NAME:-$NAME_DEFAULT}"
EMAIL="${GPG_EMAIL:-$EMAIL_DEFAULT}"

mkdir -p "$KEY_DIR"

cat >"$KEY_DIR"/gpg-batch.cfg <<CFG
Key-Type: RSA
Key-Length: 4096
Subkey-Type: RSA
Subkey-Length: 4096
Name-Real: $NAME
Name-Email: $EMAIL
Expire-Date: 1y
%no-protection
%commit
CFG

if [[ -n "${GPG_PASSPHRASE:-}" ]]; then
  # Replace %no-protection with passphrase config
  sed -i'' -e '/^%no-protection$/d' "$KEY_DIR"/gpg-batch.cfg || true
  {
    echo "Passphrase: $GPG_PASSPHRASE"
  } >> "$KEY_DIR"/gpg-batch.cfg
fi

echo "Generating GPG key for $NAME <$EMAIL> ..."
gpg --batch --gen-key "$KEY_DIR"/gpg-batch.cfg

echo "Exporting ASCII-armored keys..."
gpg --armor --export    "$EMAIL" > "$KEY_DIR"/public_gpg_key.asc
gpg --armor --export-secret-keys "$EMAIL" > "$KEY_DIR"/private_gpg_key.asc

echo "Done. Files created in $KEY_DIR:"
ls -la "$KEY_DIR"

cat <<'NEXT'

Next steps:
- Upload the public key to the Terraform Registry namespace (copy from public_gpg_key.asc).
- Add provider repo secrets:
  - GPG_PRIVATE_KEY = contents of private_gpg_key.asc
  - (Optional) GPG_PASSPHRASE if you set one
- Re-run provider release to sign checksums.

For safekeeping, move private_gpg_key.asc into your password manager.
NEXT
