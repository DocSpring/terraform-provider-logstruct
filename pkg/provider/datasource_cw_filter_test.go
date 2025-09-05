package provider

import (
    "strings"
    "testing"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

// minimal harness to exercise Read logic without full TF runtime
func TestCloudWatchFilter_BuildsPattern(t *testing.T) {
    client, err := NewMetadataClient()
    if err != nil { t.Fatalf("client: %v", err) }
    // Build inputs directly and call helper
    cfg := cwFilterModel{Struct: types.StringValue("ActionMailer"), Event: types.StringValue("delivered"), Predicates: map[string][]string{"mailer_class": {"UserMailer"}}}
    p, err := generatePattern(client, cfg)
    if err != nil { t.Fatalf("generatePattern error: %v", err) }
    if p == "" { t.Fatalf("expected pattern to be set") }
    if !strings.Contains(p, "$.evt = \"delivered\"") { t.Fatalf("missing evt in pattern: %s", p) }
}
