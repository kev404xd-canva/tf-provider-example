package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &netprobeProvider{}
)

type NetprobesProvider struct {
	version string
}

// netprobeProviderModel maps provider schema data to a Go type.
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

// Metadata returns the provider type name.
func (p *NetprobesProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "netprobes"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *NetprobesProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Optional:   true,
				Sensitive:  true,
				Validators: []validator.String{},
			},
			"password": schema.StringAttribute{
				Optional:   true,
				Sensitive:  true,
				Validators: []validator.String{},
			},
			"hostname": schema.StringAttribute{
				Optional:    true,
				Description: "The base URL for the Netprobes API. If not provided, the provider will use the environment variable NETPROBES_HOSTNAME.",
			},
		},
	}
}

func (p *NetprobeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring provider client")

	// Retrieve provider data from configuration
	var config NetprobeProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.
	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Netprobe API Host",
			"The provider cannot create the Netprobe API client as there is an unknown configuration value for the Netprobe API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_HOST environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Netprobe API Username",
			"The provider cannot create the Netprobe API client as there is an unknown configuration value for the Netprobe API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Netprobe API Password",
			"The provider cannot create the Netprobe API client as there is an unknown configuration value for the Netprobe API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the HASHICUPS_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	username := os.Getenv("NETPROBES_USERNAME")
	password := os.Getenv("NETPROBES_PASSWORD")
	hostname := os.Getenv("NETPROBES_HOSTNAME")

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}
	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}
	if !config.Hostname.IsNull() {
		hostname = config.Hostname.ValueString()
	}

	if username == "" || password == "" || hostname == "" {
		resp.Diagnostics.AddError(
			"Missing Credentials or Hostname",
			"Username, password, and hostname must be provided either via environment variables or provider configuration.",
		)
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "netprobes_host", host)
	ctx = tflog.SetField(ctx, "netprobes_username", username)
	ctx = tflog.SetField(ctx, "netprobes_password", password)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "netprobes_password")

	tflog.Debug(ctx, "Creating netprobes client")

	// Create a new Netprobe client using the configuration values
	client, err := netprobe.NewClient(&host, &username, &password)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create netprobes API Client",
			"An unexpected error occurred when creating the netprobes API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"netprobes Client Error: "+err.Error(),
		)
		return
	}

	// Make the Netprobe client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Netprobe client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *netprobeProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCoffeesDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *netprobeProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOrderResource,
	}
}
