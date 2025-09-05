This folder contains helper scripts to generate and manage a dedicated GPG key for the Terraform provider releases.

Security note: Do not commit private keys. These scripts write keys into `terraform-provider-logstruct/.gpg/` which is ignored by git (and the whole provider dir is ignored in this repo).

Quick start

1) Generate a CI keypair and export ASCII-armored keys:

   bash scripts/gpg/generate_provider_gpg.sh

   Outputs:
   - terraform-provider-logstruct/.gpg/public_gpg_key.asc
   - terraform-provider-logstruct/.gpg/private_gpg_key.asc

2) Upload public key to the Terraform Registry namespace:
   - Copy contents of `public_gpg_key.asc` to the HashiCorp Registry UI → Namespace → GPG Keys → Add.

3) Store private key in GitHub Actions secrets for the provider repo:
   - In DocSpring/terraform-provider-logstruct → Settings → Secrets and variables → Actions → New repository secret
     - `GPG_PRIVATE_KEY`: paste the entire contents of `private_gpg_key.asc`
     - (Optional) `GPG_PASSPHRASE`: if you set a passphrase during key generation

4) Verify a release (on v* tag in provider):
   - Provider CI imports the key and GoReleaser signs checksums per `.goreleaser.yaml`.

