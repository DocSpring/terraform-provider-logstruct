package provider

import (
    "context"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
    "github.com/hashicorp/terraform-plugin-framework/provider/schema"
    "github.com/hashicorp/terraform-plugin-framework/schema/validator"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/resource"
)

type logstructProvider struct {
    version string
}

type providerModel struct {
    ExportDir types.String `tfsdk:"export_dir"`
}

func New(version string) func() fwprovider.Provider {
    return func() fwprovider.Provider {
        return &logstructProvider{version: version}
    }
}

func (p *logstructProvider) Metadata(_ context.Context, _ fwprovider.MetadataRequest, resp *fwprovider.MetadataResponse) {
    resp.TypeName = "logstruct"
    resp.Version = p.version
}

func (p *logstructProvider) Schema(_ context.Context, _ fwprovider.SchemaRequest, resp *fwprovider.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "export_dir": schema.StringAttribute{Optional: true, Validators: []validator.String{}},
        },
    }
}

func (p *logstructProvider) Configure(ctx context.Context, req fwprovider.ConfigureRequest, resp *fwprovider.ConfigureResponse) {
    var cfg providerModel
    diags := req.Config.Get(ctx, &cfg)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
    client, err := NewMetadataClient()
    if err != nil {
        resp.Diagnostics.AddError("Metadata Load Error", err.Error())
        return
    }
    resp.DataSourceData = client
}

func (p *logstructProvider) DataSources(context.Context) []func() datasource.DataSource {
    return []func() datasource.DataSource{
        NewStructDataSource,
        NewCloudWatchFilterDataSource,
    }
}

func (p *logstructProvider) Resources(context.Context) []func() resource.Resource {
    return []func() resource.Resource{}
}
