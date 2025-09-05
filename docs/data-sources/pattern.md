# logstruct_pattern (Data Source)

Compiles a CloudWatch Logs JSON filter pattern for a given `source` and `event`.
The provider validates the combination at plan-time against the embedded catalog.

## Example Usage

```hcl
data "logstruct_pattern" "email_delivered" {
  source = "mailer"
  event  = var.event
}

resource "aws_cloudwatch_log_metric_filter" "email_delivered_count" {
  name           = "Email Delivered Count"
  log_group_name = var.log_group.app
  pattern        = data.logstruct_pattern.email_delivered.pattern

  metric_transformation {
    name      = "app_email_delivered_count"
    namespace = var.namespace.logs
    value     = "1"
    unit      = "Count"
  }
}
```

## Argument Reference

- `source` (String, Required) — Canonical source value (e.g., `mailer`, `job`).
- `event` (String, Required) — Serialized event value (e.g., `delivered`, `finish`).

## Attributes Reference

- `pattern` (String) — Compiled CloudWatch filter pattern.

