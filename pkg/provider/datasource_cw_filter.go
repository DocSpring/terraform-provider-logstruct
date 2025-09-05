package provider

import (
    "context"
    "fmt"
    "sort"
    "strings"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

type cwFilterDataSource struct{ client *MetadataClient }

func NewCloudWatchFilterDataSource() datasource.DataSource { return &cwFilterDataSource{} }

type cwFilterModel struct {
    Struct     types.String         `tfsdk:"struct"`
    Event      types.String         `tfsdk:"event"`
    Predicates map[string][]string  `tfsdk:"predicates"`

    Pattern types.String `tfsdk:"pattern"`
}

func (d *cwFilterDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = "logstruct_cloudwatch_filter"
}

func (d *cwFilterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "struct": schema.StringAttribute{Required: true, Description: "LogStruct struct name e.g. ActionMailer"},
            "event":  schema.StringAttribute{Required: true, Description: "Event value (serialized), validated against struct"},
            "predicates": schema.MapAttribute{
                Optional: true,
                ElementType: types.ListType{ElemType: types.StringType},
                Description: "Additional equality predicates map[field] = [values...]",
            },
            "pattern": schema.StringAttribute{Computed: true, Description: "Generated CloudWatch Logs filter pattern"},
        },
    }
}

func (d *cwFilterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
    if req.ProviderData == nil { return }
    if c, ok := req.ProviderData.(*MetadataClient); ok { d.client = c }
}

func (d *cwFilterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data cwFilterModel
    diags := req.Config.Get(ctx, &data)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() { return }

    client := d.client
    if client == nil {
        resp.Diagnostics.AddError("Provider not configured", "Missing metadata client")
        return
    }
    pattern, err := generatePattern(client, data)
    if err != nil {
        resp.Diagnostics.AddError("Pattern generation failed", err.Error())
        return
    }
    data.Pattern = types.StringValue(pattern)

    diags = resp.State.Set(ctx, &data)
    resp.Diagnostics.Append(diags...)
}

func contains(arr []string, s string) bool { for _, v := range arr { if v == s { return true } }; return false }
func escape(s string) string { return strings.ReplaceAll(s, "\"", "\\\"") }

// generatePattern builds a CloudWatch filter pattern for the given model using the catalog client.
func generatePattern(client *MetadataClient, data cwFilterModel) (string, error) {
    structName := data.Struct.ValueString()
    event := data.Event.ValueString()
    allowed, _, err := client.AllowedEventsForStruct(structName)
    if err != nil { return "", err }
    if !contains(allowed, event) {
        return "", fmt.Errorf("event %q not allowed for %s (allowed: %v)", event, structName, allowed)
    }
    var parts []string
    // add evt
    evtKey, ok := client.Keys["event"]
    if !ok { return "", fmt.Errorf("evt key missing from exports") }
    parts = append(parts, fmt.Sprintf("$.%s = \"%s\"", evtKey, event))
    // add source
    if src, fixed, err := client.FixedSourceForStruct(structName); err != nil {
        return "", err
    } else if fixed {
        srcKey, ok := client.Keys["source"]; if !ok { return "", fmt.Errorf("src key missing from exports") }
        parts = append(parts, fmt.Sprintf("$.%s = \"%s\"", srcKey, src))
    }
    // add extra predicates
    if data.Predicates != nil {
        keys := make([]string, 0, len(data.Predicates))
        for k := range data.Predicates { keys = append(keys, k) }
        sort.Strings(keys)
        for _, k := range keys {
            values := data.Predicates[k]
            if len(values) == 0 { continue }
            keyPath := fmt.Sprintf("$.%s", k)
            if len(values) == 1 {
                parts = append(parts, fmt.Sprintf("%s = \"%s\"", keyPath, escape(values[0])))
            } else {
                ors := make([]string, 0, len(values))
                for _, v := range values { ors = append(ors, fmt.Sprintf("%s = \"%s\"", keyPath, escape(v))) }
                parts = append(parts, fmt.Sprintf("(%s)", strings.Join(ors, " || ")))
            }
        }
    }
    return fmt.Sprintf("{ %s }", strings.Join(parts, " && ")), nil
}
