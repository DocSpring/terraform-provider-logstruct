# LogStruct Provider

The LogStruct Terraform provider offers type-safe helpers for working with LogStruct JSON logs:

- Validate struct/event combinations at plan-time using the provider’s embedded catalog.
- Generate CloudWatch Logs filter patterns without hand-writing stringly-typed expressions.

## Example Usage

```hcl
terraform {
  required_providers {
    logstruct = {
      source  = "DocSpring/logstruct"
      version = ">= 0.0.3"
    }
  }
}

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

Prefer a higher-level experience?
- Use the AWS module to wrap AWS resources with validated patterns:
  https://registry.terraform.io/modules/DocSpring/logstruct/aws
  (Provider remains available for custom metrics or advanced usage.)

## Argument Reference

This provider uses an embedded catalog exported from LogStruct releases, so it requires no configuration.

## Import

This provider has no importable resources.
