# logstruct_source (Data Source)

Validates a canonical `source` (e.g., `mailer`, `job`, `rails`, `storage`) and returns:

- `canonical`: The canonical source string (echoed).
- `events`: A map of allowed events for this source (keys and values are the same), suitable for `contains(keys(...), var.event)` validation.

## Example Usage

```hcl
data "logstruct_source" "mailer" {
  source = "mailer"
}

variable "event" {
  type = string
  validation {
    condition     = contains(keys(data.logstruct_source.mailer.events), var.event)
    error_message = "Invalid event for source=mailer"
  }
}
```

## Argument Reference

- `source` (String, Required)

## Attributes Reference

- `canonical` (String)
- `events` (Map(String))

