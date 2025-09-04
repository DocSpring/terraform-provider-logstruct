package provider

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
)

type MetadataClient struct {
    Keys         map[string]string
    Enums        map[string][]EnumValue
    Structs      map[string]StructInfo
}

type EnumValue struct {
    Name  string `json:"name"`
    Value string `json:"value"`
}

type FieldInfo struct {
    Optional  bool   `json:"optional"`
    Type      string `json:"type"`
    BaseEnum  string `json:"base_enum,omitempty"`
    EnumValue string `json:"enum_value,omitempty"`
    // for enum_union
    EnumValues []string `json:"enum_values,omitempty"`
}

type StructInfo struct {
    Name   string               `json:"name"`
    Fields map[string]FieldInfo `json:"fields"`
}

func NewMetadataClient(dir string) (*MetadataClient, error) {
    c := &MetadataClient{}

    // keys
    keysPath := filepath.Join(dir, "log-keys.json")
    b, err := os.ReadFile(keysPath)
    if err != nil { return nil, fmt.Errorf("read log-keys.json: %w", err) }
    if err := json.Unmarshal(b, &c.Keys); err != nil { return nil, fmt.Errorf("parse log-keys.json: %w", err) }

    // enums
    enumsPath := filepath.Join(dir, "sorbet-enums.json")
    b, err = os.ReadFile(enumsPath)
    if err != nil { return nil, fmt.Errorf("read sorbet-enums.json: %w", err) }
    if err := json.Unmarshal(b, &c.Enums); err != nil { return nil, fmt.Errorf("parse sorbet-enums.json: %w", err) }

    // structs
    structsPath := filepath.Join(dir, "sorbet-log-structs.json")
    b, err = os.ReadFile(structsPath)
    if err != nil { return nil, fmt.Errorf("read sorbet-log-structs.json: %w", err) }
    if err := json.Unmarshal(b, &c.Structs); err != nil { return nil, fmt.Errorf("parse sorbet-log-structs.json: %w", err) }

    return c, nil
}

func (c *MetadataClient) AllowedEventsForStruct(structName string) ([]string, bool, error) {
    si, ok := c.Structs[structName]
    if !ok { return nil, false, fmt.Errorf("unknown struct: %s", structName) }
    f, ok := si.Fields["event"]
    if !ok { return nil, false, fmt.Errorf("struct %s missing event field", structName) }
    switch f.Type {
    case "enum_single":
        if f.EnumValue == "" { return nil, false, fmt.Errorf("struct %s event enum_single missing value", structName) }
        return []string{f.EnumValue}, true, nil
    case "enum_union":
        if len(f.EnumValues) == 0 { return nil, false, fmt.Errorf("struct %s event enum_union empty", structName) }
        return f.EnumValues, false, nil
    default:
        return nil, false, fmt.Errorf("struct %s event field not enum: %s", structName, f.Type)
    }
}

func (c *MetadataClient) FixedSourceForStruct(structName string) (string, bool, error) {
    si, ok := c.Structs[structName]
    if !ok { return "", false, fmt.Errorf("unknown struct: %s", structName) }
    f, ok := si.Fields["source"]
    if !ok { return "", false, fmt.Errorf("struct %s missing source field", structName) }
    if f.Type == "enum_single" && f.EnumValue != "" {
        return f.EnumValue, true, nil
    }
    return "", false, nil
}

