package provider

import (
	"context"
	"fmt"
	"regexp"

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
var _ resource.Resource = &FirewallResource{}
var _ resource.ResourceWithImportState = &FirewallResource{}

func NewFirewallResource() resource.Resource {
	return &FirewallResource{}
}

// FirewallResource defines the resource implementation.
type FirewallResource struct {
	client *ShieldooClient
}

// FirewallResourceModel describes the resource data model.
type FirewallResourceModel struct {
	Name          types.String                   `tfsdk:"name"`
	Id            types.String                   `tfsdk:"id"`
	RulesInbound  FirewallResourceModelRuleValue `tfsdk:"rules_inbound"`
	RulesOutbound FirewallResourceModelRuleValue `tfsdk:"rules_outbound"`
}

type FirewallResourceModelRuleType struct {
	types.ListType
}

func (c FirewallResourceModelRuleType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := c.ListType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	f, ok := val.(types.List)
	if !ok {
		return nil, fmt.Errorf("expected types.List, got %T", val)
	}

	return FirewallResourceModelRuleValue{f}, nil
}

type FirewallResourceModelRuleValue struct {
	types.List
}

func (c FirewallResourceModelRuleValue) ParseFirewallRulesFromModel(ctx context.Context) []FirewallRule {
	var rules []FirewallRule
	for _, rule := range c.Elements() {
		tflog.Debug(ctx, "parse rule", map[string]interface{}{"rule": rule})

		rule, ok := rule.(types.Object)
		if !ok {
			tflog.Warn(ctx, "rule is not an object", map[string]interface{}{"rule": rule})
			continue
		}
		port, ok := rule.Attributes()["port"].(types.String)
		if !ok {
			tflog.Warn(ctx, "rule has no port", map[string]interface{}{"rule": rule})
			continue
		}
		protocol, ok := rule.Attributes()["protocol"].(types.String)
		if !ok {
			tflog.Warn(ctx, "rule has no protocol", map[string]interface{}{"rule": rule})
			continue
		}
		r := FirewallRule{
			Port:     port.ValueString(),
			Protocol: protocol.ValueString(),
		}
		if rule.Attributes()["group_ids"] != nil {
			group_ids, ok := rule.Attributes()["group_ids"].(types.List)
			if !ok {
				tflog.Warn(ctx, "rule has no group_ids", map[string]interface{}{"rule": rule})
				continue
			}
			for _, grp := range group_ids.Elements() {
				tmp, ok := grp.(types.String)
				if !ok {
					tflog.Warn(ctx, "rule has no group_ids", map[string]interface{}{"rule": rule})
					continue
				}
				r.Groups = append(r.Groups, Group{Id: tmp.ValueString()})
			}
		}
		if rule.Attributes()["group_names"] != nil {
			group_names, ok := rule.Attributes()["group_names"].(types.List)
			if !ok {
				tflog.Warn(ctx, "rule has no group_names", map[string]interface{}{"rule": rule})
				continue
			}
			for _, grp := range group_names.Elements() {
				tmp, ok := grp.(types.String)
				if !ok {
					tflog.Warn(ctx, "rule has no group_names", map[string]interface{}{"rule": rule})
					continue
				}
				r.Groups = append(r.Groups, Group{Name: tmp.ValueString()})
			}
		}
		if rule.Attributes()["group_object_ids"] != nil {
			group_object_ids, ok := rule.Attributes()["group_object_ids"].(types.List)
			if !ok {
				tflog.Warn(ctx, "rule has no group_object_ids", map[string]interface{}{"rule": rule})
				continue
			}
			for _, grp := range group_object_ids.Elements() {
				tmp, ok := grp.(types.String)
				if !ok {
					tflog.Warn(ctx, "rule has no group_object_ids", map[string]interface{}{"rule": rule})
					continue
				}
				r.Groups = append(r.Groups, Group{ObjectId: tmp.ValueString()})
			}
		}
		rules = append(rules, r)
	}
	return rules
}

func (r *FirewallResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall"
}

func (r *FirewallResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Firewall resource",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "Firewall name",
				Required:            true,
			},
			"rules_inbound": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Firewall inbound rules",
				CustomType: FirewallResourceModelRuleType{
					types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"port":     types.StringType,
								"protocol": types.StringType,
								"group_ids": types.ListType{
									ElemType: types.StringType,
								},
								"group_object_ids": types.ListType{
									ElemType: types.StringType,
								},
								"group_names": types.ListType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"port": schema.StringAttribute{
							MarkdownDescription: "Port",
							Required:            true,
						},
						"protocol": schema.StringAttribute{
							MarkdownDescription: "Protocol",
							Required:            true,
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
							MarkdownDescription: "Group names",
							Optional:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
			"rules_outbound": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Firewall outbound rules",
				CustomType: FirewallResourceModelRuleType{
					types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"port":     types.StringType,
								"protocol": types.StringType,
								"group_ids": types.ListType{
									ElemType: types.StringType,
								},
								"group_object_ids": types.ListType{
									ElemType: types.StringType,
								},
								"group_names": types.ListType{
									ElemType: types.StringType,
								},
							},
						},
					},
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"port": schema.StringAttribute{
							MarkdownDescription: "Port",
							Required:            true,
						},
						"protocol": schema.StringAttribute{
							MarkdownDescription: "Protocol",
							Required:            true,
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
							MarkdownDescription: "Group names",
							Optional:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Firewall identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *FirewallResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FirewallResource) NormalizeFirewallRule(rule *FirewallRule) error {
	if !regexp.MustCompile(`^(any|icmp|tcp|udp)$`).MatchString(rule.Protocol) {
		return fmt.Errorf("invalid protocol: %s", rule.Protocol)
	}
	if !regexp.MustCompile(`^([1-9][0-9]{0,3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$|^([1-9][0-9]{0,3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])-([1-9][0-9]{0,3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$|^any$`).MatchString(rule.Port) {
		return fmt.Errorf("invalid port: %s", rule.Port)
	}

	// normalize groups
	for _, g := range rule.Groups {
		if g.Id == "" && g.ObjectId == "" && g.Name == "" {
			return fmt.Errorf("invalid group: %v", g)
		}
	}
	if len(rule.Groups) > 0 {
		rule.Host = "group"
	} else {
		rule.Host = "any"
	}

	return nil
}

func (r *FirewallResource) NormalizeFirewall(fw *Firewall) error {
	for i := range fw.RulesIn {
		if err := r.NormalizeFirewallRule(&fw.RulesIn[i]); err != nil {
			return err
		}
	}
	for i := range fw.RulesOut {
		if err := r.NormalizeFirewallRule(&fw.RulesOut[i]); err != nil {
			return err
		}
	}
	// default OUT rules
	if len(fw.RulesOut) == 0 {
		fw.RulesOut = []FirewallRule{
			{
				Port:     "any",
				Protocol: "any",
				Host:     "any",
			},
		}
	}
	return nil
}

func (r *FirewallResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *FirewallResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	firewall := &Firewall{
		Name:     data.Name.ValueString(),
		RulesIn:  data.RulesInbound.ParseFirewallRulesFromModel(ctx),
		RulesOut: data.RulesOutbound.ParseFirewallRulesFromModel(ctx),
	}

	if err := r.NormalizeFirewall(firewall); err != nil {
		resp.Diagnostics.AddError("Error normalizing firewall", err.Error())
		tflog.Error(ctx, "error normalizing firewall", map[string]interface{}{"error": err.Error()})
		return
	}

	firewall, err := r.client.CreateFirewall(firewall)
	if err != nil {
		resp.Diagnostics.AddError("Error creating firewall", err.Error())
		tflog.Error(ctx, "error creating firewall", map[string]interface{}{"error": err.Error()})
		return
	}

	// For the purposes of this Firewall code, hardcoding a response value to
	// save into the Terraform state.
	data.Id = types.StringValue(firewall.Id)
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *FirewallResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	firewall, err := r.client.GetFirewall(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get Firewall, got error: %s", err))
		tflog.Error(ctx, "Client Error", map[string]interface{}{"err": err.Error()})
		return
	}

	data.Id = types.StringValue(firewall.Id)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *FirewallResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	firewall := &Firewall{
		Id:       data.Id.ValueString(),
		Name:     data.Name.ValueString(),
		RulesIn:  data.RulesInbound.ParseFirewallRulesFromModel(ctx),
		RulesOut: data.RulesOutbound.ParseFirewallRulesFromModel(ctx),
	}

	if err := r.NormalizeFirewall(firewall); err != nil {
		resp.Diagnostics.AddError("Error normalizing firewall", err.Error())
		tflog.Error(ctx, "error normalizing firewall", map[string]interface{}{"error": err.Error()})
		return
	}

	_, err := r.client.UpdateFirewall(firewall)
	if err != nil {
		resp.Diagnostics.AddError("Error updating firewall", err.Error())
		tflog.Error(ctx, "error updating firewall", map[string]interface{}{"error": err.Error()})
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *FirewallResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFirewall(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Firewall, got error: %s", err))
		tflog.Error(ctx, "Client Error", map[string]interface{}{"err": err.Error()})
		return
	}
}

func (r *FirewallResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
