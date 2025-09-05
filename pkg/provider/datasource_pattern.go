package provider

import (
    "context"
    "sort"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

type patternDataSource struct{ client *MetadataClient }

func NewPatternDataSource() datasource.DataSource { return &patternDataSource{} }

type patternModel struct {
    Source  types.String `tfsdk:"source"`
    Event   types.String `tfsdk:"event"`
    Pattern types.String `tfsdk:"pattern"`
}

func (d *patternDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = "logstruct_pattern"
}

func (d *patternDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "source": schema.StringAttribute{Required: true, Description: "Canonical source value (e.g., mailer, job, rails, storage)"},
            "event":  schema.StringAttribute{Required: true, Description: "Serialized event value for the struct (e.g., delivered, finish, database)"},
            "pattern": schema.StringAttribute{Computed: true, Description: "Compiled CloudWatch Logs filter pattern"},
        },
    }
}

func (d *patternDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
    if req.ProviderData == nil { return }
    if c, ok := req.ProviderData.(*MetadataClient); ok { d.client = c }
}

func (d *patternDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data patternModel
    diags := req.Config.Get(ctx, &data)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() { return }

    client := d.client
    if client == nil {
        resp.Diagnostics.AddError("Provider not configured", "Missing metadata client")
        return
    }
    src := data.Source.ValueString()
    ev := data.Event.ValueString()
    if src == "" || ev == "" {
        resp.Diagnostics.AddError("Invalid input", "source and event are required")
        return
    }

    // find structs that match the source
    var candidates []string
    for name, sc := range client.Structs {
        if sc.FixedSource != nil && *sc.FixedSource == src {
            candidates = append(candidates, name)
        }
    }
    if len(candidates) == 0 {
        resp.Diagnostics.AddError("Unknown source", "No structs found with fixed source = "+src)
        return
    }
    sort.Strings(candidates)

    // pick the first struct that allows the event
    var chosen string
    for _, sname := range candidates {
        allowed, _, err := client.AllowedEventsForStruct(sname)
        if err != nil { resp.Diagnostics.AddError("Lookup error", err.Error()); return }
        for _, a := range allowed { if a == ev { chosen = sname; break } }
        if chosen != "" { break }
    }
    if chosen == "" {
        resp.Diagnostics.AddError("Invalid event", "event "+ev+" is not allowed for source "+src)
        return
    }

    // build pattern
    model := cwFilterModel{Struct: types.StringValue(chosen), Event: types.StringValue(ev)}
    pat, err := generatePattern(client, model)
    if err != nil { resp.Diagnostics.AddError("Pattern generation failed", err.Error()); return }
    data.Pattern = types.StringValue(pat)

    diags = resp.State.Set(ctx, &data)
    resp.Diagnostics.Append(diags...)
}

