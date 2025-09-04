package provider

import (
    "context"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/provider"
    "github.com/hashicorp/terraform-plugin-framework/provider/schema"
    "github.com/hashicorp/terraform-plugin-framework/schema/validator"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

type logstructProvider struct {
    version string
}

type providerModel struct {
    ExportDir types.String `tfsdk:"export_dir"`
}

func New(version string) func() provider.Provider {
    return func() provider.Provider {
        return &logstructProvider{version: version}
    }
}

func (p *logstructProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
    resp.TypeName = "logstruct"
    resp.Version = p.version
}

func (p *logstructProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "export_dir": schema.StringAttribute{
                Optional:    true,
                Description: "Directory containing LogStruct JSON exports (sorbet-enums.json, sorbet-log-structs.json, log-keys.json). Defaults to ./site/lib/log-generation",
                Validators:  []validator.String{},
            },
        },
    }
}

func (p *logstructProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    var cfg providerModel
    diags := req.Config.Get(ctx, &cfg)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
    exportDir := "site/lib/log-generation"
    if !cfg.ExportDir.IsNull() && !cfg.ExportDir.IsUnknown() {
        exportDir = cfg.ExportDir.ValueString()
    }
    client, err := NewMetadataClient(exportDir)
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

func (p *logstructProvider) Resources(context.Context) []func() provider.Resource {
    return nil
}

