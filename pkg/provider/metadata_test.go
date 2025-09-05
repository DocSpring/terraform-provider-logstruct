package provider

import "testing"

func TestMetadataClient_Basics(t *testing.T) {
    c, err := NewMetadataClient()
    if err != nil { t.Fatalf("client: %v", err) }
    if c.Keys == nil || len(c.Keys) == 0 { t.Fatalf("expected keys exported") }

    // Known struct from generated catalog
    events, fixed, err := c.AllowedEventsForStruct("ActionMailer")
    if err != nil { t.Fatalf("allowed: %v", err) }
    if len(events) == 0 { t.Fatalf("expected events for ActionMailer") }

    // If source is fixed, ensure FixedSourceForStruct returns it
    if fixed {
        src, isFixed, err := c.FixedSourceForStruct("ActionMailer")
        if err != nil { t.Fatalf("fixed source: %v", err) }
        if !isFixed || src == "" { t.Fatalf("expected fixed source for ActionMailer") }
    }
}

