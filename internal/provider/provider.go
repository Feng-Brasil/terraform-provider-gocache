// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/Feng-Brasil/terraform-provider-gocache/internal/client/gocache"
	"github.com/Feng-Brasil/terraform-provider-gocache/internal/services/dns_record"
	"github.com/Feng-Brasil/terraform-provider-gocache/internal/services/smart_rule"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type gocacheProviderModel struct {
	Token types.String `tfsdk:"token"`
}

var (
	_ provider.Provider = &gocacheProvider{}
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &gocacheProvider{
			version: version,
		}
	}
}

type gocacheProvider struct {
	version string
}

func (p *gocacheProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "gocache"
	resp.Version = p.version
}

func (p *gocacheProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with GoCache API.",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Description: "GoCache API token. May also be provided via GOCACHE_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *gocacheProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring GoCache client")

	var config gocacheProviderModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown GoCache Token",
			"The provider cannot create the GoCache API client because the token value is unknown.",
		)

		return
	}

	token := os.Getenv("GOCACHE_TOKEN")

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing GoCache Token",
			"Set token in the provider configuration or use the GOCACHE_TOKEN environment variable.",
		)

		return
	}

	ctx = tflog.SetField(ctx, "gocache_token", token)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "gocache_token")

	client := gocache.NewClient(token)

	err := client.ValidateToken()

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to authenticate in GoCache API",
			err.Error(),
		)

		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured GoCache client", map[string]any{"success": true})
}

func (p *gocacheProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		dns_record.NewDNSRecordsDataSource,
	}
}

func (p *gocacheProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		dns_record.NewDNSRecordResource,
		smart_rule.NewGeneralSmartRuleResource,
		smart_rule.NewRedirectSmartRuleResource,
	}
}
