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
var _ datasource.DataSource = &ServerDataSource{}

func NewServerDataSource() datasource.DataSource {
	return &ServerDataSource{}
}

// ServerDataSource defines the data source implementation.
type ServerDataSource struct {
	client *ShieldooClient
}

// ServerDataSourceModel describes the data source data model.
type ServerDataSourceModel struct {
	Name          types.String `tfsdk:"name"`
	Id            types.String `tfsdk:"id"`
	IpAddress     types.String `tfsdk:"ip_address"`
	Configuration types.String `tfsdk:"configuration"`
}

func (d *ServerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *ServerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Server data source",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Server name",
				Required:            true,
			},
			"ip_address": schema.StringAttribute{
				MarkdownDescription: "Server IP address",
				Computed:            true,
			},
			"configuration": schema.StringAttribute{
				MarkdownDescription: "Server configuration data (secret)",
				Computed:            true,
				Sensitive:           true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Server identifier",
				Computed:            true,
			},
		},
	}
}

func (d *ServerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServerDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	server, err := d.client.GetServer(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("ERROR: %s", err.Error()))
		tflog.Error(ctx, "Client Error", map[string]interface{}{"err": err.Error()})
		return
	}

	data.Id = types.StringValue(server.Id)
	data.IpAddress = types.StringValue(server.IpAddress)
	data.Configuration = types.StringValue(server.Configuration)
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
