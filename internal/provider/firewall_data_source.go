package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &FirewallDataSource{}

func NewFirewallDataSource() datasource.DataSource {
	return &FirewallDataSource{}
}

// FirewallDataSource defines the data source implementation.
type FirewallDataSource struct {
	client *ShieldooClient
}

// FirewallDataSourceModel describes the data source data model.
type FirewallDataSourceModel struct {
	Name types.String `tfsdk:"name"`
	Id   types.String `tfsdk:"id"`
}

func (d *FirewallDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall"
}

func (d *FirewallDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Firewall data source",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Firewall name",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Firewall identifier",
				Computed:            true,
			},
		},
	}
}

func (d *FirewallDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ShieldooClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ShieldooConfigureData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *FirewallDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FirewallDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	firewall, err := d.client.GetFirewall(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ERROR: %s", err.Error()))
		tflog.Error(ctx, "Client Error", map[string]interface{}{"err": err.Error()})
		return
	}

	data.Id = types.StringValue(firewall.Id)
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
