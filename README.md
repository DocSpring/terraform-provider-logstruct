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

Register this public key in the Terraform Registry under **User Settings > Signing Keys**. When rotating the signing
key, update the repository's public key, the release secret and the Registry registration together. Keep earlier public
keys registered so users can verify older releases.

After the GitHub release finishes, verify that the new version appears in the
[Terraform Registry](https://registry.terraform.io/providers/DocSpring/logstruct/latest). If it is missing, use
**Resync** in the provider's Registry settings. Resync also ensures the GitHub release webhook exists. Confirm the
version is listed by the [Registry API](https://registry.terraform.io/v1/providers/DocSpring/logstruct/versions), then run
`terraform init` in a configuration pinned to that version using the normal Registry installation path.

See HashiCorp's [provider publishing guide](https://developer.hashicorp.com/terraform/registry/providers/publishing)
and [Registry FAQ](https://developer.hashicorp.com/terraform/registry/faq) for signing-key and webhook requirements.
