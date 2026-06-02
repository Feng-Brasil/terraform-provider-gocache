// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package smart_rule

import (
	"context"
	"fmt"

	"github.com/Feng-Brasil/terraform-provider-gocache/internal/client/gocache"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	allowedCachingBehavior = []string{"default", "ignore_query_string"}
	allowedCacheMode       = []string{"off", "default", "full"}
	allowedSSLMode         = []string{"partial", "full"}
	allowedConnection      = []string{"close", "keep-alive"}
	allowedSignedURLType   = []string{"s3qs", "off"}
	allowedWAFMode         = []string{"block", "simulate", "challenge"}
	allowedWAFLevel        = []string{"low", "medium", "high"}
	allowedExpiresTTL      = []int64{-1, 3600, 7200, 14400, 43200, 86400, 172800, 345600, 604800, 1296000, 2592000, 15552000, 31536000}
	allowedCacheTTL        = []int64{300, 600, 900, 1800, 3600, 7200, 14400, 86400, 172800, 604800, 1296000}
	allowedImageLevel      = []int64{0, 65, 75, 90}
)

type generalSmartRuleModel struct {
	ID     types.String        `tfsdk:"id"`
	Domain types.String        `tfsdk:"domain"`
	Name   types.String        `tfsdk:"name"`
	Status types.Bool          `tfsdk:"status"`
	Notes  types.String        `tfsdk:"notes"`
	Match  *generalMatchModel  `tfsdk:"match"`
	Action *generalActionModel `tfsdk:"action"`
}

type generalActionModel struct {
	SetHost                  types.String `tfsdk:"set_host"`
	Backend                  types.String `tfsdk:"backend"`
	SetURI                   types.String `tfsdk:"set_uri"`
	CustomCacheKey           types.String `tfsdk:"custom_cache_key"`
	ExpiresTTL               types.Int64  `tfsdk:"expires_ttl"`
	CachingBehavior          types.String `tfsdk:"caching_behavior"`
	CacheMode                types.String `tfsdk:"cache_mode"`
	CacheTTL                 types.Int64  `tfsdk:"cache_ttl"`
	CORS                     types.String `tfsdk:"cors"`
	SSLMode                  types.String `tfsdk:"ssl_mode"`
	Connection               types.String `tfsdk:"connection"`
	Hide                     types.String `tfsdk:"hide"`
	Set                      types.String `tfsdk:"set"`
	SetReqHeader             types.List   `tfsdk:"set_req_header"`
	SignedURLKey             types.String `tfsdk:"signed_url_key"`
	SignedURLType            types.String `tfsdk:"signed_url_type"`
	Cache301                 types.Bool   `tfsdk:"cache_301"`
	Cache302                 types.Bool   `tfsdk:"cache_302"`
	Cache404                 types.Bool   `tfsdk:"cache_404"`
	GzipStatus               types.Bool   `tfsdk:"gzip_status"`
	IgnoreCacheControl       types.Bool   `tfsdk:"ignore_cache_control"`
	IgnoreExpires            types.Bool   `tfsdk:"ignore_expires"`
	IgnoreVary               types.Bool   `tfsdk:"ignore_vary"`
	WAFStatus                types.Bool   `tfsdk:"waf_status"`
	WAFMode                  types.String `tfsdk:"waf_mode"`
	WAFLevel                 types.String `tfsdk:"waf_level"`
	RateLimitStatus          types.Bool   `tfsdk:"rate_limit_status"`
	ImageOptimize            types.Bool   `tfsdk:"image_optimize"`
	ImageOptimizeWebp        types.Bool   `tfsdk:"image_optimize_webp"`
	ImageOptimizeProgressive types.Bool   `tfsdk:"image_optimize_progressive"`
	ImageOptimizeMetadata    types.Bool   `tfsdk:"image_optimize_metadata"`
	ImageOptimizeLevel       types.Int64  `tfsdk:"image_optimize_level"`
}

var (
	_ resource.Resource                = &generalSmartRuleResource{}
	_ resource.ResourceWithConfigure   = &generalSmartRuleResource{}
	_ resource.ResourceWithImportState = &generalSmartRuleResource{}
)

func NewGeneralSmartRuleResource() resource.Resource {
	return &generalSmartRuleResource{}
}

type generalSmartRuleResource struct {
	client *gocache.Client
}

func (r *generalSmartRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_general_smart_rule"
}

func (r *generalSmartRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a GoCache general Smart Rule. General Smart Rules apply cache, performance, security and header behaviors to requests that match a set of criteria. Available on the GROWTH plan.",
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
			"match": matchSchema(true),
			"action": schema.SingleNestedAttribute{
				Description: "Behaviors applied to requests that match the criteria. At least one attribute should be set.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"set_host": schema.StringAttribute{
						Description: "Overrides the request Host header. Example: `mybucket.s3.amazonaws.com`.",
						Optional:    true,
					},
					"backend": schema.StringAttribute{
						Description: "Sets the destination IP or hostname. Example: `s3.amazonaws.com`.",
						Optional:    true,
					},
					"set_uri": schema.StringAttribute{
						Description: "Overrides the URI. The wildcard (`*`) value from `match.request_uri` can be used through the `$1` variable. Example: `/imagens/$1`.",
						Optional:    true,
					},
					"custom_cache_key": schema.StringAttribute{
						Description: "Customizes the cache key. Supports the variables `$cookie_NAME` and `$http_NAME` to use a specific cookie/header value, and `$geoip2_data_country_code`/`$geoip2_data_continent_code` for country/continent. Example: `logado,$cookie_user`. URL parameters `$1`, `$2`, ... can be used when the criteria is a URL with a wildcard.",
						Optional:    true,
					},
					"expires_ttl": schema.Int64Attribute{
						Description: "Browser cache time, in seconds. The value `-1` disables browser cache. Allowed values: `-1`, `3600`, `7200`, `14400`, `43200`, `86400`, `172800`, `345600`, `604800`, `1296000`, `2592000`, `15552000`, `31536000`.",
						Optional:    true,
						Validators: []validator.Int64{
							int64validator.OneOf(allowedExpiresTTL...),
						},
					},
					"caching_behavior": schema.StringAttribute{
						Description: "Cache behavior. Allowed values: `default`, `ignore_query_string`.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(allowedCachingBehavior...),
						},
					},
					"cache_mode": schema.StringAttribute{
						Description: "Cache type (off, static content only, or full cache). Allowed values: `off`, `default`, `full`.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(allowedCacheMode...),
						},
					},
					"cache_ttl": schema.Int64Attribute{
						Description: "Cache expiration time, in seconds. Allowed values: `300`, `600`, `900`, `1800`, `3600`, `7200`, `14400`, `86400`, `172800`, `604800`, `1296000`.",
						Optional:    true,
						Validators: []validator.Int64{
							int64validator.OneOf(allowedCacheTTL...),
						},
					},
					"cors": schema.StringAttribute{
						Description: "Sets Cross-Origin. Example: `http://www.gocache.com.br`.",
						Optional:    true,
					},
					"ssl_mode": schema.StringAttribute{
						Description: "SSL mode. The value `partial` is equivalent to Edge mode. Allowed values: `partial`, `full`.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(allowedSSLMode...),
						},
					},
					"connection": schema.StringAttribute{
						Description: "Sets the request Connection header. Allowed values: `close`, `keep-alive`.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(allowedConnection...),
						},
					},
					"hide": schema.StringAttribute{
						Description: "Removes a response header. The value must be the header name.",
						Optional:    true,
					},
					"set": schema.StringAttribute{
						Description: "Sets a response header. Example: `Strict-Transport-Security: max-age=86400`.",
						Optional:    true,
					},
					"set_req_header": schema.ListAttribute{
						Description: "Sets request headers. Example: `X-GoCache-Custom-Header:true`. At most 3 request headers can be specified.",
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtMost(3),
						},
					},
					"signed_url_key": schema.StringAttribute{
						Description: "Sets the signed URL key.",
						Optional:    true,
					},
					"signed_url_type": schema.StringAttribute{
						Description: "Sets the signed URL type. Allowed values: `s3qs`, `off`.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(allowedSignedURLType...),
						},
					},
					"cache_301": schema.BoolAttribute{
						Description: "Enable/disable caching of requests with status 301.",
						Optional:    true,
					},
					"cache_302": schema.BoolAttribute{
						Description: "Enable/disable caching of requests with status 302.",
						Optional:    true,
					},
					"cache_404": schema.BoolAttribute{
						Description: "Enable/disable caching of requests with status 404.",
						Optional:    true,
					},
					"gzip_status": schema.BoolAttribute{
						Description: "Enable/disable Gzip compression.",
						Optional:    true,
					},
					"ignore_cache_control": schema.BoolAttribute{
						Description: "Enable/disable ignoring the Cache-Control header.",
						Optional:    true,
					},
					"ignore_expires": schema.BoolAttribute{
						Description: "Enable/disable ignoring the Expires header.",
						Optional:    true,
					},
					"ignore_vary": schema.BoolAttribute{
						Description: "Enable/disable ignoring the Vary header.",
						Optional:    true,
					},
					"waf_status": schema.BoolAttribute{
						Description: "Enable/disable the WAF. Available on the GROWTH plan.",
						Optional:    true,
					},
					"waf_mode": schema.StringAttribute{
						Description: "WAF operation mode. Allowed values: `block`, `simulate`, `challenge`. Available on the GROWTH plan.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(allowedWAFMode...),
						},
					},
					"waf_level": schema.StringAttribute{
						Description: "WAF protection level. Allowed values: `low`, `medium`, `high`. Available on the GROWTH plan.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf(allowedWAFLevel...),
						},
					},
					"rate_limit_status": schema.BoolAttribute{
						Description: "Enable/disable the Rate Limit counter. Available on the CUSTOM plan.",
						Optional:    true,
					},
					"image_optimize": schema.BoolAttribute{
						Description: "Enable/disable image optimization. Available on the GROWTH plan.",
						Optional:    true,
					},
					"image_optimize_webp": schema.BoolAttribute{
						Description: "Convert images to WEBP. Available on the GROWTH plan.",
						Optional:    true,
					},
					"image_optimize_progressive": schema.BoolAttribute{
						Description: "Convert JPEG images to the progressive format. Available on the GROWTH plan.",
						Optional:    true,
					},
					"image_optimize_metadata": schema.BoolAttribute{
						Description: "Remove image metadata. Available on the GROWTH plan.",
						Optional:    true,
					},
					"image_optimize_level": schema.Int64Attribute{
						Description: "Image optimization level. Use `0` to keep the original level, `90` for low, `75` for medium and `65` for high. Available on the GROWTH plan.",
						Optional:    true,
						Validators: []validator.Int64{
							int64validator.OneOf(allowedImageLevel...),
						},
					},
				},
			},
		},
	}
}

func (r *generalSmartRuleResource) buildRequest(ctx context.Context, plan generalSmartRuleModel, diags *diag.Diagnostics) gocache.GeneralSmartRuleRequest {
	action := plan.Action

	return gocache.GeneralSmartRuleRequest{
		Name:   stringPtr(plan.Name),
		Status: boolPtr(plan.Status),
		Notes:  stringPtr(plan.Notes),
		Match:  generalMatchToClient(ctx, plan.Match, diags),
		Action: gocache.GeneralSmartRuleAction{
			SetHost:                  stringPtr(action.SetHost),
			Backend:                  stringPtr(action.Backend),
			SetURI:                   stringPtr(action.SetURI),
			CustomCacheKey:           stringPtr(action.CustomCacheKey),
			ExpiresTTL:               int64Ptr(action.ExpiresTTL),
			CachingBehavior:          stringPtr(action.CachingBehavior),
			CacheMode:                stringPtr(action.CacheMode),
			CacheTTL:                 int64Ptr(action.CacheTTL),
			CORS:                     stringPtr(action.CORS),
			SSLMode:                  stringPtr(action.SSLMode),
			Connection:               stringPtr(action.Connection),
			Hide:                     stringPtr(action.Hide),
			Set:                      stringPtr(action.Set),
			SetReqHeader:             stringListPtr(ctx, action.SetReqHeader, diags),
			SignedURLKey:             stringPtr(action.SignedURLKey),
			SignedURLType:            stringPtr(action.SignedURLType),
			Cache301:                 boolPtr(action.Cache301),
			Cache302:                 boolPtr(action.Cache302),
			Cache404:                 boolPtr(action.Cache404),
			GzipStatus:               boolPtr(action.GzipStatus),
			IgnoreCacheControl:       boolPtr(action.IgnoreCacheControl),
			IgnoreExpires:            boolPtr(action.IgnoreExpires),
			IgnoreVary:               boolPtr(action.IgnoreVary),
			WAFStatus:                boolPtr(action.WAFStatus),
			WAFMode:                  stringPtr(action.WAFMode),
			WAFLevel:                 stringPtr(action.WAFLevel),
			RateLimitStatus:          boolPtr(action.RateLimitStatus),
			ImageOptimize:            boolPtr(action.ImageOptimize),
			ImageOptimizeWebp:        boolPtr(action.ImageOptimizeWebp),
			ImageOptimizeProgressive: boolPtr(action.ImageOptimizeProgressive),
			ImageOptimizeMetadata:    boolPtr(action.ImageOptimizeMetadata),
			ImageOptimizeLevel:       int64Ptr(action.ImageOptimizeLevel),
		},
	}
}

func (r *generalSmartRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan generalSmartRuleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request := r.buildRequest(ctx, plan, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.client.CreateGeneralSmartRule(plan.Domain.ValueString(), request)

	if err != nil {
		resp.Diagnostics.AddError("Error creating GoCache general Smart Rule", err.Error())

		return
	}

	plan.ID = types.StringValue(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *generalSmartRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state generalSmartRuleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	rules, err := r.client.ListGeneralSmartRules(state.Domain.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Error reading GoCache general Smart Rules", err.Error())

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

		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *generalSmartRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan generalSmartRuleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request := r.buildRequest(ctx, plan, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateGeneralSmartRule(plan.Domain.ValueString(), plan.ID.ValueString(), request)

	if err != nil {
		resp.Diagnostics.AddError("Error updating GoCache general Smart Rule", err.Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *generalSmartRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state generalSmartRuleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteGeneralSmartRule(state.Domain.ValueString(), state.ID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Error deleting GoCache general Smart Rule", err.Error())

		return
	}
}

func (r *generalSmartRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importSmartRuleState(ctx, req, resp)
}

func (r *generalSmartRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func importSmartRuleState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := splitImportID(req.ID)

	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier in the format \"domain/id\". Got: %q", req.ID),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
