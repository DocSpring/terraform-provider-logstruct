# logstruct_filters (Data Source)

Aggregated, catalog-driven helpers for LogStruct:

- `filters`: Map of struct name → (map of event → CloudWatch filter `pattern`).
- `events`: Map of struct name → list of allowed events.
- `sources`: Map of struct name → fixed source (if any).
- `keys`: Canonical key names for common fields (e.g., `evt`, `src`).

## Example Usage

```hcl
data "logstruct_filters" "all" {}

locals {
  email_delivered = data.logstruct_filters.all.filters["ActionMailer"]["delivered"]
}

resource "aws_cloudwatch_log_metric_filter" "email_delivered_count" {
  name           = "Email Delivered Count"
  log_group_name = var.log_group.app
  pattern        = local.email_delivered

  metric_transformation {
    name      = "app_email_delivered_count"
    namespace = var.namespace.logs
    value     = "1"
    unit      = "Count"
  }
}
```

## Attributes Reference

- `filters` (Map(Map(String)))
- `events` (Map(List(String)))
- `sources` (Map(String))
- `keys` (Map(String))

