package provider

import (
    "context"
    "sort"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/attr"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

type filtersDataSource struct{ client *MetadataClient }

func NewFiltersDataSource() datasource.DataSource { return &filtersDataSource{} }

type filtersModel struct {
    // maps of struct name -> event -> pattern
    Filters types.Map `tfsdk:"filters"`
    // struct name -> list of events
    Events types.Map `tfsdk:"events"`
    // struct name -> fixed source (if any)
    Sources types.Map `tfsdk:"sources"`
    // canonical key names (evt, src, etc.)
    Keys types.Map `tfsdk:"keys"`
}

func (d *filtersDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = "logstruct_filters"
}

func (d *filtersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "filters": schema.MapAttribute{
                Computed:    true,
                Description: "Map of struct name -> (map of event -> CloudWatch filter pattern)",
                ElementType: types.MapType{ElemType: types.StringType},
            },
            "events": schema.MapAttribute{
                Computed:    true,
                Description: "Map of struct name -> list of allowed events",
                ElementType: types.ListType{ElemType: types.StringType},
            },
            "sources": schema.MapAttribute{
                Computed:    true,
                Description: "Map of struct name -> fixed source (if any)",
                ElementType: types.StringType,
            },
            "keys": schema.MapAttribute{
                Computed:    true,
                Description: "Canonical key names for common fields (evt, src, etc.)",
                ElementType: types.StringType,
            },
        },
    }
}

func (d *filtersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
    if req.ProviderData == nil { return }
    if c, ok := req.ProviderData.(*MetadataClient); ok { d.client = c }
}

func (d *filtersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var state filtersModel

    client := d.client
    if client == nil {
        resp.Diagnostics.AddError("Provider not configured", "Missing metadata client")
        return
    }

    // Build events map and filters map
    // Sort struct names for stable output
    structNames := make([]string, 0, len(client.Structs))
    for name := range client.Structs { structNames = append(structNames, name) }
    sort.Strings(structNames)

    // events: map[string]types.List
    eventsElems := make(map[string]types.List, len(structNames))
    // sources: map[string]types.String
    sourceElems := make(map[string]types.String, len(structNames))
    // filters: map[string]types.Map (event -> pattern)
    filterElems := make(map[string]types.Map, len(structNames))

    for _, sname := range structNames {
        allowed, _, err := client.AllowedEventsForStruct(sname)
        if err != nil { resp.Diagnostics.AddError("Lookup error", err.Error()); return }
        // events list
        evVals := make([]string, 0, len(allowed))
        evVals = append(evVals, allowed...)
        // convert []string -> []types.Value
        tv := make([]attr.Value, 0, len(evVals))
        for _, s := range evVals { tv = append(tv, types.StringValue(s)) }
        eventsElems[sname] = types.ListValueMust(types.StringType, tv)

        // fixed source
        if src, fixed, err := client.FixedSourceForStruct(sname); err == nil && fixed {
            sourceElems[sname] = types.StringValue(src)
        } else {
            sourceElems[sname] = types.StringNull()
        }

        // patterns for each event
        evMap := make(map[string]types.String, len(allowed))
        for _, ev := range allowed {
            // reuse existing pattern generator
            model := cwFilterModel{Struct: types.StringValue(sname), Event: types.StringValue(ev)}
            pat, err := generatePattern(client, model)
            if err != nil { resp.Diagnostics.AddError("Pattern generation failed", err.Error()); return }
            evMap[ev] = types.StringValue(pat)
        }
        // cast to map[string]types.Value
        mm := make(map[string]attr.Value, len(evMap))
        for k, v := range evMap { mm[k] = v }
        filterElems[sname] = types.MapValueMust(types.StringType, mm)
    }

    // keys map
    keyElems := make(map[string]types.String, len(client.Keys))
    for k, v := range client.Keys { keyElems[k] = types.StringValue(v) }

    // assign to state
    state.Events = types.MapValueMust(types.ListType{ElemType: types.StringType}, castMapListToValues(eventsElems))
    state.Sources = types.MapValueMust(types.StringType, castMapStringToValues(sourceElems))
    state.Filters = types.MapValueMust(types.MapType{ElemType: types.StringType}, castMapMapStringToValues(filterElems))
    // keys is a simple map[string]string
    kv := make(map[string]attr.Value, len(keyElems))
    for k, v := range keyElems { kv[k] = v }
    state.Keys = types.MapValueMust(types.StringType, kv)

    diags := resp.State.Set(ctx, &state)
    resp.Diagnostics.Append(diags...)
}

// helpers to convert typed maps into map[string]types.Value for MapValueMust
func castMapListToValues(in map[string]types.List) map[string]attr.Value {
    out := make(map[string]attr.Value, len(in))
    for k, v := range in { out[k] = v }
    return out
}
func castMapStringToValues(in map[string]types.String) map[string]attr.Value {
    out := make(map[string]attr.Value, len(in))
    for k, v := range in { out[k] = v }
    return out
}
func castMapMapStringToValues(in map[string]types.Map) map[string]attr.Value {
    out := make(map[string]attr.Value, len(in))
    for k, v := range in { out[k] = v }
    return out
}
