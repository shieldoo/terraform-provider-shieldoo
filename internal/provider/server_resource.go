package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ServerResource{}
var _ resource.ResourceWithImportState = &ServerResource{}

func NewServerResource() resource.Resource {
	return &ServerResource{}
}

// ServerResource defines the resource implementation.
type ServerResource struct {
	client *ShieldooClient
}

// ServerResourceModel describes the resource data model.
type ServerResourceModel struct {
	Name                    types.String                     `tfsdk:"name"`
	Id                      types.String                     `tfsdk:"id"`
	Configuration           types.String                     `tfsdk:"configuration"`
	Description             types.String                     `tfsdk:"description"`
	IpAddress               types.String                     `tfsdk:"ip_address"`
	FirewallId              types.String                     `tfsdk:"firewall_id"`
	GroupIds                types.List                       `tfsdk:"group_ids"`
	GroupObjectIds          types.List                       `tfsdk:"group_object_ids"`
	GroupNames              types.List                       `tfsdk:"group_names"`
	Listeners               ServerResourceModelListenerValue `tfsdk:"listeners"`
	Autoupdate              types.Bool                       `tfsdk:"autoupdate"`
	OSUpdateEnabled         types.Bool                       `tfsdk:"os_update_enabled"`
	OSSecurityUpdateEnabled types.Bool                       `tfsdk:"os_security_update_enabled"`
	OSAllUpdateEnabled      types.Bool                       `tfsdk:"os_all_update_enabled"`
	OSRestartAfterUpdate    types.Bool                       `tfsdk:"os_restart_after_update"`
	OSUpdateHour            types.Int64                      `tfsdk:"os_update_hour"`
}

type ServerResourceModelListenerType struct {
	types.ListType
}

func (c ServerResourceModelListenerType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := c.ListType.ValueFromTerraform(ctx, in)

	s, ok := val.(types.List)
	if !ok {
		return nil, fmt.Errorf("value is not a list")
	}

	return ServerResourceModelListenerValue{s}, err
}

type ServerResourceModelListenerValue struct {
	types.List
}

func (c ServerResourceModelListenerValue) ParseServerListenersFromModel(ctx context.Context) []Listener {
	var listeners []Listener
	for _, listener := range c.Elements() {
		listener, ok := listener.(types.Object)
		if !ok {
			tflog.Warn(ctx, "listener is not an object", map[string]interface{}{"listener": listener})
			continue
		}
		listenport, ok := listener.Attributes()["listen_port"].(types.Int64)
		if !ok {
			tflog.Warn(ctx, "listen_port is not an int64", map[string]interface{}{"listen": listener})
			continue
		}
		protocol, ok := listener.Attributes()["protocol"].(types.String)
		if !ok {
			tflog.Warn(ctx, "protocol is not a string", map[string]interface{}{"listen": listener})
			continue
		}
		forwardport, ok := listener.Attributes()["forward_port"].(types.Int64)
		if !ok {
			tflog.Warn(ctx, "forward_port is not an int64", map[string]interface{}{"listen": listener})
			continue
		}
		forwardhost, ok := listener.Attributes()["forward_host"].(types.String)
		if !ok {
			tflog.Warn(ctx, "forward_host is not a string", map[string]interface{}{"listen": listener})
			continue
		}
		description, ok := listener.Attributes()["description"].(types.String)
		if !ok {
			tflog.Warn(ctx, "description is not a string", map[string]interface{}{"listen": listener})
			continue
		}

		r := Listener{
			// unchecked type assertion
			ListenPort:  int(listenport.ValueInt64()),
			Protocol:    protocol.ValueString(),
			ForwardPort: int(forwardport.ValueInt64()),
			ForwardHost: forwardhost.ValueString(),
			Description: description.ValueString(),
		}
		listeners = append(listeners, r)
	}
	return listeners
}

func (r *ServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *ServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Server resource",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Server name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Server description",
				Optional:            true,
			},
			"autoupdate": schema.BoolAttribute{
				MarkdownDescription: "Autoupdate",
				Optional:            true,
			},
			"os_update_enabled": schema.BoolAttribute{
				MarkdownDescription: "OS Update Enabled",
				Optional:            true,
			},
			"os_security_update_enabled": schema.BoolAttribute{
				MarkdownDescription: "OS Security Update Enabled",
				Optional:            true,
			},
			"os_all_update_enabled": schema.BoolAttribute{
				MarkdownDescription: "OS All Update Enabled",
				Optional:            true,
			},
			"os_restart_after_update": schema.BoolAttribute{
				MarkdownDescription: "OS Restart After Update",
				Optional:            true,
			},
			"os_update_hour": schema.Int64Attribute{
				MarkdownDescription: "OS Update Hour (0=anytime, then hour in day in GMT)",
				Optional:            true,
			},
			"firewall_id": schema.StringAttribute{
				MarkdownDescription: "Firewall ID",
				Required:            true,
			},
			"ip_address": schema.StringAttribute{
				MarkdownDescription: "IP Address (if omitted, will be assigned automatically)",
				Optional:            true,
			},
			"group_ids": schema.ListAttribute{
				MarkdownDescription: "Group IDs",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"group_object_ids": schema.ListAttribute{
				MarkdownDescription: "Group Object IDs",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"group_names": schema.ListAttribute{
				MarkdownDescription: "Group Names",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"listeners": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Server listeners",
				CustomType: ServerResourceModelListenerType{
					types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"listen_port":  types.Int64Type,
								"protocol":     types.StringType,
								"forward_port": types.Int64Type,
								"forward_host": types.StringType,
								"description":  types.StringType,
							},
						},
					},
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"listen_port": schema.Int64Attribute{
							Required:            true,
							MarkdownDescription: "Listen port",
						},
						"protocol": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Protocol",
						},
						"forward_port": schema.Int64Attribute{
							Required:            true,
							MarkdownDescription: "Forward port",
						},
						"forward_host": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Forward host",
						},
						"description": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Description",
						},
					},
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Server identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"configuration": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Server configuration data (secret)",
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ShieldooClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ShieldooConfigureData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ServerResource) NormalizeServerListener(listener *Listener) error {
	if listener.ListenPort < 1 || listener.ListenPort > 65535 {
		return fmt.Errorf("listen_port must be between 1 and 65535")
	}
	if listener.ForwardPort < 1 || listener.ForwardPort > 65535 {
		return fmt.Errorf("forward_port must be between 1 and 65535")
	}
	if listener.Protocol != "tcp" && listener.Protocol != "udp" {
		return fmt.Errorf("protocol must be tcp or udp")
	}
	return nil
}

func (r *ServerResource) NormalizeServer(server *Server) error {
	for i := range server.Listeners {
		if err := r.NormalizeServerListener(&server.Listeners[i]); err != nil {
			return err
		}
	}
	return nil
}

func (r *ServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ServerResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	server := &Server{
		Name:        data.Name.ValueString(),
		Autoupdate:  data.Autoupdate.ValueBool(),
		Description: data.Description.ValueString(),
		IpAddress:   data.IpAddress.ValueString(),
		Firewall:    Firewall{Id: data.FirewallId.ValueString()},
		Listeners:   data.Listeners.ParseServerListenersFromModel(ctx),
		OSUpdatePolicy: ServerOSAutoupdatePolicy{
			Enabled:                   data.OSUpdateEnabled.ValueBool(),
			SecurityAutoupdateEnabled: data.OSSecurityUpdateEnabled.ValueBool(),
			AllAutoupdateEnabled:      data.OSAllUpdateEnabled.ValueBool(),
			RestartAfterUpdate:        data.OSRestartAfterUpdate.ValueBool(),
			UpdateHour:                int(data.OSUpdateHour.ValueInt64()),
		},
	}

	if !data.GroupIds.IsNull() {
		for _, g := range data.GroupIds.Elements() {
			g, ok := g.(types.String)
			if !ok {
				resp.Diagnostics.AddError("Error parsing group id", "Error parsing group id")
				tflog.Error(ctx, "error parsing group id")
				return
			}
			server.Groups = append(server.Groups, Group{Id: g.ValueString()})
		}
	}

	if !data.GroupNames.IsNull() {
		for _, g := range data.GroupNames.Elements() {
			g, ok := g.(types.String)
			if !ok {
				resp.Diagnostics.AddError("Error parsing group name", "Error parsing group name")
				tflog.Error(ctx, "error parsing group name")
				return
			}
			server.Groups = append(server.Groups, Group{Name: g.ValueString()})
		}
	}

	if !data.GroupObjectIds.IsNull() {
		for _, g := range data.GroupObjectIds.Elements() {
			g, ok := g.(types.String)
			if !ok {
				resp.Diagnostics.AddError("Error parsing group object id", "Error parsing group object id")
				tflog.Error(ctx, "error parsing group object id")
				return
			}
			server.Groups = append(server.Groups, Group{ObjectId: g.ValueString()})
		}
	}

	if err := r.NormalizeServer(server); err != nil {
		resp.Diagnostics.AddError("Error normalizing Server", err.Error())
		tflog.Error(ctx, "error normalizing Server", map[string]interface{}{"error": err.Error()})
		return
	}

	Server, err := r.client.CreateServer(server)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Server", err.Error())
		tflog.Error(ctx, "error creating Server", map[string]interface{}{"error": err.Error()})
		return
	}

	// For the purposes of this Server code, hardcoding a response value to
	// save into the Terraform state.
	data.Id = types.StringValue(Server.Id)
	data.Configuration = types.StringValue(Server.Configuration)
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ServerResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	server, err := r.client.GetServer(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Server, got error: %s", err))
		tflog.Error(ctx, "Client Error", map[string]interface{}{"err": err.Error()})
		return
	}

	data.Id = types.StringValue(server.Id)
	data.Configuration = types.StringValue(server.Configuration)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ServerResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	server := &Server{
		Id:          data.Id.ValueString(),
		Autoupdate:  data.Autoupdate.ValueBool(),
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		IpAddress:   data.IpAddress.ValueString(),
		Firewall:    Firewall{Id: data.FirewallId.ValueString()},
		Listeners:   data.Listeners.ParseServerListenersFromModel(ctx),
		OSUpdatePolicy: ServerOSAutoupdatePolicy{
			Enabled:                   data.OSUpdateEnabled.ValueBool(),
			SecurityAutoupdateEnabled: data.OSSecurityUpdateEnabled.ValueBool(),
			AllAutoupdateEnabled:      data.OSAllUpdateEnabled.ValueBool(),
			RestartAfterUpdate:        data.OSRestartAfterUpdate.ValueBool(),
			UpdateHour:                int(data.OSUpdateHour.ValueInt64()),
		},
	}

	if !data.GroupIds.IsNull() {
		for _, g := range data.GroupIds.Elements() {
			g, ok := g.(types.String)
			if !ok {
				resp.Diagnostics.AddError("Error parsing group id", "Error parsing group id")
				tflog.Error(ctx, "error parsing group id")
				return
			}
			server.Groups = append(server.Groups, Group{Id: g.ValueString()})
		}
	}

	if !data.GroupNames.IsNull() {
		for _, g := range data.GroupNames.Elements() {
			g, ok := g.(types.String)
			if !ok {
				resp.Diagnostics.AddError("Error parsing group name", "Error parsing group name")
				tflog.Error(ctx, "error parsing group name")
				return
			}
			server.Groups = append(server.Groups, Group{Name: g.ValueString()})
		}
	}

	if !data.GroupObjectIds.IsNull() {
		for _, g := range data.GroupObjectIds.Elements() {
			g, ok := g.(types.String)
			if !ok {
				resp.Diagnostics.AddError("Error parsing group object id", "Error parsing group object id")
				tflog.Error(ctx, "error parsing group object id")
				return
			}
			server.Groups = append(server.Groups, Group{ObjectId: g.ValueString()})
		}
	}

	if err := r.NormalizeServer(server); err != nil {
		resp.Diagnostics.AddError("Error normalizing Server", err.Error())
		tflog.Error(ctx, "error normalizing Server", map[string]interface{}{"error": err.Error()})
		return
	}

	server, err := r.client.UpdateServer(server)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Server", err.Error())
		tflog.Error(ctx, "error updating Server", map[string]interface{}{"error": err.Error()})
		return
	}

	data.Configuration = types.StringValue(server.Configuration)
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ServerResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteServer(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Server, got error: %s", err))
		tflog.Error(ctx, "Client Error", map[string]interface{}{"err": err.Error()})
		return
	}
}

func (r *ServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
