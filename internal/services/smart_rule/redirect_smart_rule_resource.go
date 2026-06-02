// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package smart_rule

import (
	"context"
	"fmt"

	"github.com/Feng-Brasil/terraform-provider-gocache/internal/client/gocache"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var allowedRedirectType = []int64{301, 302}

type redirectSmartRuleModel struct {
	ID     types.String         `tfsdk:"id"`
	Domain types.String         `tfsdk:"domain"`
	Name   types.String         `tfsdk:"name"`
	Status types.Bool           `tfsdk:"status"`
	Notes  types.String         `tfsdk:"notes"`
	Match  *redirectMatchModel  `tfsdk:"match"`
	Action *redirectActionModel `tfsdk:"action"`
}

type redirectActionModel struct {
	RedirectType types.Int64  `tfsdk:"redirect_type"`
	RedirectTo   types.String `tfsdk:"redirect_to"`
}

var (
	_ resource.Resource                = &redirectSmartRuleResource{}
	_ resource.ResourceWithConfigure   = &redirectSmartRuleResource{}
	_ resource.ResourceWithImportState = &redirectSmartRuleResource{}
)

func NewRedirectSmartRuleResource() resource.Resource {
	return &redirectSmartRuleResource{}
}

type redirectSmartRuleResource struct {
	client *gocache.Client
}

func (r *redirectSmartRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redirect_smart_rule"
}

func (r *redirectSmartRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a GoCache redirect Smart Rule. Redirect Smart Rules issue HTTP 301/302 redirects for requests that match a set of criteria.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Smart Rule identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Description: "Domain the Smart Rule belongs to. Example: `gocache.com.br`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Rule name.",
				Optional:    true,
			},
			"status": schema.BoolAttribute{
				Description: "Whether the rule is active. Defaults to `true`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"notes": schema.StringAttribute{
				Description: "Rule description.",
				Optional:    true,
			},
			"match": matchSchema(false),
			"action": schema.SingleNestedAttribute{
				Description: "Redirect applied to requests that match the criteria.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"redirect_type": schema.Int64Attribute{
						Description: "Redirect type. Allowed values: `301`, `302`.",
						Required:    true,
						Validators: []validator.Int64{
							int64validator.OneOf(allowedRedirectType...),
						},
					},
					"redirect_to": schema.StringAttribute{
						Description: "Redirect URL. The wildcard (`*`) value(s) from `match.request` can be used through the variables `$1`, `$2`, .... Example: `https://blog.gocache.com.br/$1`.",
						Required:    true,
					},
				},
			},
		},
	}
}

func (r *redirectSmartRuleResource) buildRequest(ctx context.Context, plan redirectSmartRuleModel, diags *diag.Diagnostics) gocache.RedirectSmartRuleRequest {
	return gocache.RedirectSmartRuleRequest{
		Name:   stringPtr(plan.Name),
		Status: boolPtr(plan.Status),
		Notes:  stringPtr(plan.Notes),
		Match:  redirectMatchToClient(ctx, plan.Match, diags),
		Action: gocache.RedirectSmartRuleAction{
			RedirectType: int64Ptr(plan.Action.RedirectType),
			RedirectTo:   stringPtr(plan.Action.RedirectTo),
		},
	}
}

func (r *redirectSmartRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan redirectSmartRuleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request := r.buildRequest(ctx, plan, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateRedirectSmartRule(plan.Domain.ValueString(), request)

	if err != nil {
		resp.Diagnostics.AddError("Error creating GoCache redirect Smart Rule", err.Error())

		return
	}

	plan.ID = types.StringValue(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *redirectSmartRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state redirectSmartRuleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rules, err := r.client.ListRedirectSmartRules(state.Domain.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Error reading GoCache redirect Smart Rules", err.Error())

		return
	}

	for _, rule := range rules {
		if rule.ID != state.ID.ValueString() {
			continue
		}

		if rule.Metadata.Name != "" {
			state.Name = types.StringValue(rule.Metadata.Name)
		}

		if rule.Metadata.Notes != "" {
			state.Notes = types.StringValue(rule.Metadata.Notes)
		}

		if rule.Metadata.Status != "" {
			state.Status = types.BoolValue(rule.Metadata.Status == "true")
		}

		if rule.Action.RedirectType != nil {
			state.Action.RedirectType = types.Int64Value(*rule.Action.RedirectType)
		}

		if rule.Action.RedirectTo != nil {
			state.Action.RedirectTo = types.StringValue(*rule.Action.RedirectTo)
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *redirectSmartRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan redirectSmartRuleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request := r.buildRequest(ctx, plan, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateRedirectSmartRule(plan.Domain.ValueString(), plan.ID.ValueString(), request)

	if err != nil {
		resp.Diagnostics.AddError("Error updating GoCache redirect Smart Rule", err.Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *redirectSmartRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state redirectSmartRuleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRedirectSmartRule(state.Domain.ValueString(), state.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Error deleting GoCache redirect Smart Rule", err.Error())

		return
	}
}

func (r *redirectSmartRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importSmartRuleState(ctx, req, resp)
}

func (r *redirectSmartRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*gocache.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *gocache.Client, got: %T.", req.ProviderData),
		)

		return
	}

	r.client = client
}
