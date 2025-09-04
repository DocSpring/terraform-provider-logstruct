package provider

import (
    "context"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/attr"
)

type structDataSource struct{ client *MetadataClient }

func NewStructDataSource() datasource.DataSource { return &structDataSource{} }

type structDataModel struct {
    Struct        types.String   `tfsdk:"struct"`
    FixedSource   types.String   `tfsdk:"fixed_source"`
    AllowedEvents []types.String `tfsdk:"allowed_events"`
    Keys          types.Map      `tfsdk:"keys"`
}

func (d *structDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = "logstruct_struct"
}

func (d *structDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "struct": schema.StringAttribute{Required: true, Description: "LogStruct struct name e.g. ActionMailer"},
            "fixed_source": schema.StringAttribute{Computed: true},
            "allowed_events": schema.ListAttribute{Computed: true, ElementType: types.StringType},
            "keys": schema.MapAttribute{Computed: true, ElementType: types.StringType},
        },
    }
}

func (d *structDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
    if req.ProviderData == nil { return }
    if c, ok := req.ProviderData.(*MetadataClient); ok { d.client = c }
}

func (d *structDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data structDataModel
    diags := req.Config.Get(ctx, &data)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() { return }

    client := d.client
    if client == nil {
        resp.Diagnostics.AddError("Provider not configured", "Missing metadata client")
        return
    }
    allowed, single, err := client.AllowedEventsForStruct(data.Struct.ValueString())
    if err != nil { resp.Diagnostics.AddError("Lookup error", err.Error()); return }
    if src, fixed, err := client.FixedSourceForStruct(data.Struct.ValueString()); err != nil {
        resp.Diagnostics.AddError("Lookup error", err.Error()); return
    } else if fixed {
        data.FixedSource = types.StringValue(src)
    } else {
        data.FixedSource = types.StringNull()
    }
    // events
    data.AllowedEvents = []types.String{}
    for _, e := range allowed { data.AllowedEvents = append(data.AllowedEvents, types.StringValue(e)) }
    _ = single
    // keys map
    kv := map[string]attr.Value{}
    for k, v := range client.Keys { kv[k] = types.StringValue(v) }
    m, md := types.MapValue(types.StringType, kv)
    resp.Diagnostics.Append(md...)
    data.Keys = m

    diags = resp.State.Set(ctx, &data)
    resp.Diagnostics.Append(diags...)
}
