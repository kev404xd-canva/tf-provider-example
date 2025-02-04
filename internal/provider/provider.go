package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type NetprobesProvider struct {
	version string
}

type NetprobeProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &NetprobesProvider{
			version: version,
		}
	}
}

func (p *NetprobesProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "netprobes"
	resp.Version = p.version
}

func (p *NetprobesProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required:    true,
				Description: "The API server URL (e.g., https://api.example.com)",
			},
			"username": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "API authentication username",
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "API authentication password",
			},
		},
	}
}

func (p *NetprobesProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config NetprobeProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate required fields
	if config.Host.IsUnknown() || config.Host.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing API Host",
			"Provider requires a host to be specified",
		)
	}

	if config.Username.IsUnknown() || config.Username.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing API Username",
			"Provider requires a username to be specified",
		)
	}

	if config.Password.IsUnknown() || config.Password.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing API Password",
			"Provider requires a password to be specified",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the API client with configured credentials
	client := NewAPIClient(
		config.Host.ValueString(),
		config.Username.ValueString(),
		config.Password.ValueString(),
	)

	// Make the client available to all resources
	resp.ResourceData = client
	resp.DataSourceData = client

	tflog.Info(ctx, "Configured Netprobes API client", map[string]any{
		"host": config.Host.ValueString(),
	})
}

func (p *NetprobesProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewTargetResource,
	}
}

func (p *NetprobesProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// No data sources needed for your use case
	}
}
