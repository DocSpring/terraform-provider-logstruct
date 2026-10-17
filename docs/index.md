# LogStruct Provider

The LogStruct Terraform provider builds CloudWatch Logs filter patterns for LogStruct's log lines, checked at plan
time against the sources and events LogStruct writes. Its catalog of sources and events is compiled into the
provider, from the LogStruct release with the same version number.

## Example Usage

```hcl
terraform {
  required_providers {
    logstruct = {
      source  = "DocSpring/logstruct"
      version = "~> 0.2.0"
    }
  }
}

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

Prefer a higher-level experience?
- Use the AWS module to wrap AWS resources with validated patterns:
  https://registry.terraform.io/modules/DocSpring/logstruct/aws
  (Provider remains available for custom metrics or advanced usage.)

## Argument Reference

This provider uses an embedded catalog exported from LogStruct releases, so it requires no configuration.

## Import

This provider has no importable resources.
