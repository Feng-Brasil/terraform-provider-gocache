// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package smart_rule

import (
	"context"
	"strings"

	"github.com/Feng-Brasil/terraform-provider-gocache/internal/client/gocache"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Allowed values shared across Smart Rules, as documented by the GoCache API.
var (
	allowedSchemes     = []string{"http", "https", "http*"}
	allowedHTTPVersion = []string{"HTTP/1.0", "HTTP/1.1", "HTTP/2.0"}
	allowedDeviceType  = []string{"desktop", "mobile", "bot", "na"}
	allowedMethods     = []string{"GET", "POST", "PUT", "DELETE"}
	allowedContinents  = []string{"SA", "NA", "OC", "EU", "AF", "AS"}
)

type generalMatchModel struct {
	Host       types.String `tfsdk:"host"`
	Scheme     types.String `tfsdk:"scheme"`
	RequestURI types.String `tfsdk:"request_uri"`
	sharedMatchModel
	OriginASN types.Int64 `tfsdk:"origin_asn"`
}

type redirectMatchModel struct {
	Request types.String `tfsdk:"request"`
	sharedMatchModel
}

// sharedMatchModel maps criteria that are present in both Smart Rule match
// blocks. Terraform Plugin Framework requires exact struct/schema matches, so
// resource-specific fields live in generalMatchModel and redirectMatchModel.
type sharedMatchModel struct {
	HTTPVersion     types.List   `tfsdk:"http_version"`
	DeviceType      types.List   `tfsdk:"device_type"`
	HTTPUserAgent   types.String `tfsdk:"http_user_agent"`
	Cookie          types.String `tfsdk:"cookie"`
	CookieContent   types.String `tfsdk:"cookie_content"`
	RequestMethod   types.List   `tfsdk:"request_method"`
	HTTPReferer     types.String `tfsdk:"http_referer"`
	RemoteAddress   types.String `tfsdk:"remote_address"`
	Header          types.String `tfsdk:"header"`
	OriginCountry   types.String `tfsdk:"origin_country"`
	OriginContinent types.List   `tfsdk:"origin_continent"`
}

// matchSchemaAttributes returns the schema attributes for the `match` block.
// includeURI controls whether the general-rule attributes (host/scheme/
// request_uri/origin_asn) or the redirect-rule attribute (request) are exposed.
func matchSchemaAttributes(general bool) map[string]schema.Attribute {
	attrs := map[string]schema.Attribute{
		"http_version": schema.ListAttribute{
			Description: "HTTP version used in the request. Allowed values: `HTTP/1.0`, `HTTP/1.1`, `HTTP/2.0`.",
			Optional:    true,
			ElementType: types.StringType,
			Validators: []validator.List{
				listvalidator.ValueStringsAre(stringvalidator.OneOf(allowedHTTPVersion...)),
			},
		},
		"device_type": schema.ListAttribute{
			Description: "Type of the device performing the request. Allowed values: `desktop`, `mobile`, `bot`, `na`.",
			Optional:    true,
			ElementType: types.StringType,
			Validators: []validator.List{
				listvalidator.ValueStringsAre(stringvalidator.OneOf(allowedDeviceType...)),
			},
		},
		"http_user_agent": schema.StringAttribute{
			Description: "User agent of the request. Supports the wildcard (`*`). Example: `*Mozilla*`.",
			Optional:    true,
		},
		"cookie": schema.StringAttribute{
			Description: "Match when one or more cookies from a list are present. Separate multiple cookies with a comma. Supports the wildcard (`*`). Example: `wp_login,wp_test`.",
			Optional:    true,
		},
		"cookie_content": schema.StringAttribute{
			Description: "Match when a cookie has a specific value. Name and value are separated by an equals sign (`=`). Supports the wildcard (`*`) in the cookie value. Example: `wp_test=yes`.",
			Optional:    true,
		},
		"request_method": schema.ListAttribute{
			Description: "HTTP request method. Allowed values: `GET`, `POST`, `PUT`, `DELETE`.",
			Optional:    true,
			ElementType: types.StringType,
			Validators: []validator.List{
				listvalidator.ValueStringsAre(stringvalidator.OneOf(allowedMethods...)),
			},
		},
		"http_referer": schema.StringAttribute{
			Description: "Referer of the request.",
			Optional:    true,
		},
		"remote_address": schema.StringAttribute{
			Description: "Source IP or IP range of the request. For ranges, only the masks `/32`, `/24` and `/16` are available.",
			Optional:    true,
		},
		"header": schema.StringAttribute{
			Description: "Match when the request has a specific header. Name and value are separated by a colon (`:`). Supports the wildcard (`*`) in the header value. Example: `Accept:*`.",
			Optional:    true,
		},
		"origin_country": schema.StringAttribute{
			Description: "Country of origin (ISO 3166-1 alpha-2 code). Multiple values can be combined with a pipe (`|`). Example: `US|UK`.",
			Optional:    true,
		},
		"origin_continent": schema.ListAttribute{
			Description: "Continent of origin. Allowed values: `SA`, `NA`, `OC`, `EU`, `AF`, `AS`.",
			Optional:    true,
			ElementType: types.StringType,
			Validators: []validator.List{
				listvalidator.ValueStringsAre(stringvalidator.OneOf(allowedContinents...)),
			},
		},
	}

	if general {
		attrs["host"] = schema.StringAttribute{
			Description: "Domain/subdomain used to access the site. Example: `www.gocache.com.br`.",
			Optional:    true,
		}
		attrs["scheme"] = schema.StringAttribute{
			Description: "Access protocol. Allowed values: `http`, `https`, `http*`.",
			Optional:    true,
			Validators: []validator.String{
				stringvalidator.OneOf(allowedSchemes...),
			},
		}
		attrs["request_uri"] = schema.StringAttribute{
			Description: "Request URI. Supports the wildcard (`*`). Example: `/admin/*`.",
			Optional:    true,
		}
		attrs["origin_asn"] = schema.Int64Attribute{
			Description: "ASN of the requester, represented only by numbers.",
			Optional:    true,
		}
	} else {
		attrs["request"] = schema.StringAttribute{
			Description: "Full request URL, composed of the HTTP protocol, a host and the URI. Supports the wildcard (`*`). Example: `http://www.gocache.com.br/blog/*`.",
			Optional:    true,
		}
	}

	return attrs
}

func matchSchema(general bool) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Criteria that a request must meet for the rule to be applied. At least one attribute should be set.",
		Required:    true,
		Attributes:  matchSchemaAttributes(general),
	}
}

func stringPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	value := v.ValueString()

	return &value
}

func int64Ptr(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	value := v.ValueInt64()

	return &value
}

func boolPtr(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	value := v.ValueBool()

	return &value
}

func stringListPtr(ctx context.Context, v types.List, diags *diag.Diagnostics) []string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	var values []string

	diags.Append(v.ElementsAs(ctx, &values, false)...)

	return values
}

func sharedMatchToClient(ctx context.Context, m sharedMatchModel, diags *diag.Diagnostics) gocache.SmartRuleMatch {
	return gocache.SmartRuleMatch{
		HTTPVersion:     stringListPtr(ctx, m.HTTPVersion, diags),
		DeviceType:      stringListPtr(ctx, m.DeviceType, diags),
		HTTPUserAgent:   stringPtr(m.HTTPUserAgent),
		Cookie:          stringPtr(m.Cookie),
		CookieContent:   stringPtr(m.CookieContent),
		RequestMethod:   stringListPtr(ctx, m.RequestMethod, diags),
		HTTPReferer:     stringPtr(m.HTTPReferer),
		RemoteAddress:   stringPtr(m.RemoteAddress),
		Header:          stringPtr(m.Header),
		OriginCountry:   stringPtr(m.OriginCountry),
		OriginContinent: stringListPtr(ctx, m.OriginContinent, diags),
	}
}

func generalMatchToClient(ctx context.Context, m *generalMatchModel, diags *diag.Diagnostics) gocache.SmartRuleMatch {
	if m == nil {
		return gocache.SmartRuleMatch{}
	}

	match := sharedMatchToClient(ctx, m.sharedMatchModel, diags)
	match.Host = stringPtr(m.Host)
	match.Scheme = stringPtr(m.Scheme)
	match.RequestURI = stringPtr(m.RequestURI)
	match.OriginASN = int64Ptr(m.OriginASN)

	return match
}

func redirectMatchToClient(ctx context.Context, m *redirectMatchModel, diags *diag.Diagnostics) gocache.SmartRuleMatch {
	if m == nil {
		return gocache.SmartRuleMatch{}
	}

	match := sharedMatchToClient(ctx, m.sharedMatchModel, diags)
	match.Request = stringPtr(m.Request)

	return match
}

// splitImportID splits a "domain/id" import identifier into its two parts,
// splitting on the first separator only.
func splitImportID(id string) []string {
	return strings.SplitN(id, "/", 2)
}
