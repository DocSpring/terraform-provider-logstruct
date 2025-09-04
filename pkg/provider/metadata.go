package provider

import (
    "fmt"
    "github.com/DocSpring/terraform-provider-logstruct/pkg/data"
)

type MetadataClient struct {
    Keys    map[string]string
    Structs map[string]StructCatalog
}

type StructCatalog = data.StructCatalog

func NewMetadataClient() (*MetadataClient, error) {
    return &MetadataClient{Keys: data.CatalogData.Keys, Structs: data.CatalogData.Structs}, nil
}

func (c *MetadataClient) AllowedEventsForStruct(structName string) ([]string, bool, error) {
    si, ok := c.Structs[structName]
    if !ok { return nil, false, fmt.Errorf("unknown struct: %s", structName) }
    if len(si.AllowedEvents) == 1 { return si.AllowedEvents, true, nil }
    if len(si.AllowedEvents) > 1 { return si.AllowedEvents, false, nil }
    return nil, false, fmt.Errorf("struct %s has no allowed events", structName)
}

func (c *MetadataClient) FixedSourceForStruct(structName string) (string, bool, error) {
    si, ok := c.Structs[structName]
    if !ok { return "", false, fmt.Errorf("unknown struct: %s", structName) }
    if si.FixedSource != nil { return *si.FixedSource, true, nil }
    return "", false, nil
}
