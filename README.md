# terraform-provider-logstruct

Terraform provider for LogStruct: type-safe CloudWatch filter patterns and LogStruct metadata validation at plan time.

Links:

- Website: https://logstruct.com/
- LogStruct (Ruby gem): https://github.com/DocSpring/logstruct
- Provider (this repo): https://github.com/DocSpring/terraform-provider-logstruct

## Features

- Validate that a `struct` and `event` combination is valid based on LogStruct's typed definitions.
- Generate CloudWatch Logs filter patterns without stringly-typed values.
- Fail fast during `terraform validate/plan` if LogStruct enums/keys drift.

## Data Sources

### `logstruct_struct`

Inputs:

- `struct` (string): e.g. `"ActionMailer"`

Outputs:

- `fixed_source` (string, null if not fixed)
- `allowed_events` (list of strings)
- `keys` (map): canonical key names, e.g. `evt`, `src`, etc.

### `logstruct_cloudwatch_filter`

Inputs:

- `struct` (string)
- `event` (string, serialized value as emitted by LogStruct)
- `predicates` (map(string => list(string)), optional): additional equality clauses

Outputs:

- `pattern` (string): CloudWatch filter pattern `{ $.src = "mailer" && $.evt = "delivered" ... }`

## Installation

```hcl
terraform {
  required_providers {
    logstruct = {
      source  = "DocSpring/logstruct"
      version = ">= 0.1.0"
    }
  }
}
```

## Example

```hcl
data "logstruct_cloudwatch_filter" "email_delivered" {
  struct = "ActionMailer"
  event  = "delivered"
}

resource "aws_cloudwatch_log_metric_filter" "email_delivered_count" {
  name           = "Email Delivered Count"
  log_group_name = var.log_group.docspring
  pattern        = data.logstruct_cloudwatch_filter.email_delivered.pattern

  metric_transformation {
    name          = "docspring_email_delivered_count"
    namespace     = var.namespace.logs
    value         = "1"
    default_value = "0"
    unit          = "Count"
  }
}
```

See more examples at https://logstruct.com/docs/terraform.

## Releasing

Use GoReleaser to build and publish GitHub releases with platform-specific zips and checksums. Tags must be semantic versions prefixed with `v` (e.g. `v0.1.0`).

The release workflow signs the checksums with the `GPG_PRIVATE_KEY` repository secret. Its public counterpart is
[`.gpg/public_gpg_key.asc`](.gpg/public_gpg_key.asc), with fingerprint
`531ED8661E17EAC44124AE57BD2DE4CA3CF5BA4D`.

The current public key is already registered for the DocSpring namespace. When rotating the signing key, update the
repository's public key, the release secret and the namespace's GPG keys together. Claimed namespaces manage keys in
HCP Terraform under **Registry > Public namespaces > Settings**. Keep earlier public keys registered so users can
verify older releases.

After the GitHub release finishes, verify that the new version appears in the
[Terraform Registry](https://registry.terraform.io/providers/DocSpring/logstruct/latest). Claimed namespaces receive
release events through the Terraform Cloud GitHub App; a repository webhook is not required. If a version is missing,
check the App's repository access and the namespace's version history. Review the **Resync** preview before confirming:
it can delete Registry versions that it cannot find on GitHub. Do not confirm a preview that would delete releases
which still exist on GitHub. Confirm the new version is listed by the
[Registry API](https://registry.terraform.io/v1/providers/DocSpring/logstruct/versions), then run
`terraform init` in a configuration pinned to that version using the normal Registry installation path.

### Registry ingestion issue

As of 2026-10-21, the Registry lists versions through 0.1.12, while 0.2.0, 0.2.1 and 0.2.2 are published on GitHub.
[HashiCorp issue #67](https://github.com/hashicorp/terraform-registry-support/issues/67) tracks the ingestion failure and
the Resync preview that proposes removing existing versions. The current signing key is already registered. The
absence of a repository webhook is expected for this claimed namespace, as
[HashiCorp explains](https://github.com/hashicorp/terraform-registry-support/issues/53#issuecomment-5797889249).

Provider 0.1.12 can still be installed through the Registry. Its `mailer` events match 0.2.2, and both versions generate
the same `mailer` / `delivered` pattern and reject the invalid event `deliverd`. Pin `version = "= 0.1.12"` for that
example while ingestion is unavailable. Other integrations have catalog changes, so this is not a substitute for
0.2.2 when using its newer events.

For 0.2.2, Terraform supports a
[filesystem mirror](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-installation). Download the
published archive for your platform, verify its checksum against the signed SHA256SUMS, and put it under
`<mirror>/registry.terraform.io/docspring/logstruct/`. Configure `filesystem_mirror` to include only
`registry.terraform.io/docspring/logstruct` and exclude that address from `direct`. Keep the provider source address
`DocSpring/logstruct` and pin `version = "= 0.2.2"`. This installation path passed `terraform init` and the valid and
invalid mailer plans on macOS ARM64; it does not verify Registry ingestion.

See HashiCorp's [namespace key management guide](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/organizations/public-namespace/manage),
[artifact management guide](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/organizations/public-namespace/artifacts)
and [provider publishing guide](https://developer.hashicorp.com/terraform/registry/providers/publishing).
