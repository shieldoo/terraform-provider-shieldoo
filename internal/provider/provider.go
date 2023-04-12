package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ShieldooProvider satisfies various provider interfaces.
var _ provider.Provider = &ShieldooProvider{}

// ShieldooProvider defines the provider implementation.
type ShieldooProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// SshieldooProviderModel describes the provider data model.
type ShieldooProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	ApiKey   types.String `tfsdk:"apikey"`
}

func (p *ShieldooProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "shieldoo"
	resp.Version = p.version
}

func (p *ShieldooProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Shieldoo API endpoint",
				Optional:            true,
			},
			"apikey": schema.StringAttribute{
				MarkdownDescription: "Shieldoo API Key",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *ShieldooProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ShieldooProviderModel

	apiKey := os.Getenv("SHIELDOO_API_KEY")
	endpoint := os.Getenv("SHIELDOO_ENDPOINT")

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Endpoint.ValueString() != "" {
		endpoint = data.Endpoint.ValueString()
	}

	if data.ApiKey.ValueString() != "" {
		apiKey = data.ApiKey.ValueString()
	}

	// Configuration values are now available.
	if endpoint == "" {
		resp.Diagnostics.AddError(
			"endpoint is required",
			"Please set the endpoint in the provider configuration block.",
		)
		return
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"apikey is required",
			"Please set the apikey in the provider configuration block.",
		)
		return
	}

	// Example client configuration for data sources and resources
	client := &ShieldooClient{
		uri:    endpoint,
		apiKey: apiKey,
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ShieldooProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewFirewallResource,
		NewServerResource,
	}
}

func (p *ShieldooProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewFirewallDataSource,
		NewServerDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ShieldooProvider{
			version: version,
		}
	}
}
