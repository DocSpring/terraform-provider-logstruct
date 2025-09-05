package provider

import (
    "context"
    "sort"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/attr"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

type sourceDataSource struct{ client *MetadataClient }

func NewSourceDataSource() datasource.DataSource { return &sourceDataSource{} }

type sourceModel struct {
    Source types.String `tfsdk:"source"`
    // outputs
    Canonical types.String `tfsdk:"canonical"`
    Events    types.Map    `tfsdk:"events"` // map[string]string
}

func (d *sourceDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = "logstruct_source"
}

func (d *sourceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "source": schema.StringAttribute{Required: true, Description: "Canonical source value (e.g., mailer, job, rails, storage)"},
            "canonical": schema.StringAttribute{Computed: true, Description: "Canonical source name (echoed)"},
            "events": schema.MapAttribute{Computed: true, Description: "Map of allowed events for this source", ElementType: types.StringType},
        },
    }
}

func (d *sourceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
    if req.ProviderData == nil { return }
    if c, ok := req.ProviderData.(*MetadataClient); ok { d.client = c }
}

func (d *sourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data sourceModel
    diags := req.Config.Get(ctx, &data)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() { return }

    client := d.client
    if client == nil {
        resp.Diagnostics.AddError("Provider not configured", "Missing metadata client")
        return
    }
    src := data.Source.ValueString()
    if src == "" {
        resp.Diagnostics.AddError("Invalid source", "source cannot be empty")
        return
    }

    // Collect all structs that have this fixed source
    var structs []string
    for name, sc := range client.Structs {
        if sc.FixedSource != nil && *sc.FixedSource == src {
            structs = append(structs, name)
        }
    }
    if len(structs) == 0 {
        resp.Diagnostics.AddError("Unknown source", "No structs found with fixed source = "+src)
        return
    }
    sort.Strings(structs)

    // Union of events across all matching structs
    evset := map[string]struct{}{}
    for _, sname := range structs {
        allowed, _, err := client.AllowedEventsForStruct(sname)
        if err != nil { resp.Diagnostics.AddError("Lookup error", err.Error()); return }
        for _, ev := range allowed { evset[ev] = struct{}{} }
    }
    // Build map[string]string where key==value
    evMap := make(map[string]attr.Value, len(evset))
    // stable order
    var keys []string
    for ev := range evset { keys = append(keys, ev) }
    sort.Strings(keys)
    for _, ev := range keys { evMap[ev] = types.StringValue(ev) }

    data.Canonical = types.StringValue(src)
    data.Events = types.MapValueMust(types.StringType, evMap)

    diags = resp.State.Set(ctx, &data)
    resp.Diagnostics.Append(diags...)
}

