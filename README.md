# terraform-provider-logstruct

Terraform provider for [LogStruct](https://logstruct.com/): CloudWatch Logs filter patterns for LogStruct's log lines,
checked at plan time against the sources and events LogStruct writes.

Links:

- Docs: https://logstruct.com/docs/terraform
- LogStruct (Ruby gem): https://github.com/DocSpring/logstruct
- AWS module that wraps the CloudWatch resources: https://registry.terraform.io/modules/DocSpring/logstruct/aws
- Provider (this repo): https://github.com/DocSpring/terraform-provider-logstruct

## How it works

The provider has a catalog of LogStruct's sources (`src`) and the events (`evt`) each one writes, compiled into the
provider (`pkg/data/catalog.json`). Each provider release is built from the LogStruct release with the same version
number: LogStruct's Release workflow regenerates the catalog, commits it here and tags this repository. An unknown
source or event fails `terraform validate` and `terraform plan`.

The provider has no settings and no resources.

## Data Sources

### `logstruct_source`

Checks a source and lists its events.

- `source` (String, required): a source, e.g. `mailer` or `job`.
- `canonical` (String): the source.
- `events` (Map of String): the source's events, keyed by themselves, for `contains(keys(...), var.event)` checks.

### `logstruct_pattern`

Builds a CloudWatch Logs filter pattern for a source and an event.

- `source` (String, required): a source, e.g. `mailer`.
- `event` (String, required): one of the source's events, e.g. `delivered`.
- `pattern` (String): the pattern, e.g. `{ $.src = "mailer" && $.evt = "delivered" }`.

## Installation

```hcl
terraform {
  required_providers {
    logstruct = {
      source  = "DocSpring/logstruct"
      version = "~> 0.2.0"
    }
  }
}
```

Use the provider version that matches the LogStruct version your app runs.

## Example

```hcl
data "logstruct_pattern" "email_delivered" {
  source = "mailer"
  event  = "delivered"
}

resource "aws_cloudwatch_log_metric_filter" "email_delivered_count" {
  name           = "Email Delivered Count"
  log_group_name = var.log_group_name
  pattern        = data.logstruct_pattern.email_delivered.pattern

  metric_transformation {
    name      = "mailer_delivered"
    namespace = var.namespace
    value     = "1"
    unit      = "Count"
  }
}
```

More examples: https://logstruct.com/docs/terraform

## Releasing

Releases are tagged by LogStruct's Release workflow. A `v*` tag runs `.github/workflows/release.yml`, which builds the
platform zips and checksums with GoReleaser and publishes them as a GitHub release.
