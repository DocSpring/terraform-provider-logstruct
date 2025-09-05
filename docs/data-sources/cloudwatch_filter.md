# logstruct_cloudwatch_filter (Data Source)

Compiles a CloudWatch Logs JSON filter pattern for a given LogStruct `struct` and `event`. The provider validates the combination at plan time against the embedded catalog generated from LogStruct releases.

## Example Usage

```hcl
data "logstruct_cloudwatch_filter" "email_delivered" {
  struct = "ActionMailer"
  event  = "delivered"
}

resource "aws_cloudwatch_log_metric_filter" "email_delivered_count" {
  name           = "Email Delivered Count"
  log_group_name = var.log_group.app
  pattern        = data.logstruct_cloudwatch_filter.email_delivered.pattern

  metric_transformation {
    name      = "app_email_delivered_count"
    namespace = var.namespace.logs
    value     = "1"
    unit      = "Count"
  }
}
```

## Argument Reference

- `struct` (String, Required) — LogStruct struct name (e.g., `ActionMailer`, `GoodJob`, `SQL`).
- `event` (String, Required) — Serialized event value for the struct (e.g., `delivered`, `finish`, `database`).
- `predicates` (Map(List(String)), Optional) — Additional equality predicates to include in the filter. Keys are JSON paths relative to the root (without `$.`), values are lists of accepted values.

## Attributes Reference

- `pattern` (String) — Compiled CloudWatch Logs filter pattern, for example:

  ```
  { $.src = "mailer" && $.evt = "delivered" }
  ```

