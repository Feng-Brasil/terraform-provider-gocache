package provider

import (
	"context"
	"fmt"

	"terraform-provider-hashicups/internal/gocache"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type dnsRecordsDataSourceModel struct {
	Domain  types.String     `tfsdk:"domain"`
	Records []dnsRecordState `tfsdk:"records"`
}

type dnsRecordState struct {
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Content types.String `tfsdk:"content"`
	TTL     types.String `tfsdk:"ttl"`
	Cloud   types.String `tfsdk:"cloud"`
}

var (
	_ datasource.DataSource              = &dnsRecordsDataSource{}
	_ datasource.DataSourceWithConfigure = &dnsRecordsDataSource{}
)

func NewDNSRecordsDataSource() datasource.DataSource {
	return &dnsRecordsDataSource{}
}

type dnsRecordsDataSource struct {
	client *gocache.Client
}

func (d *dnsRecordsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_records"
}

func (d *dnsRecordsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists GoCache DNS records for a domain.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "Domain name.",
				Required:    true,
			},
			"records": schema.ListNestedAttribute{
				Description: "DNS records returned by GoCache.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "DNS record ID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "DNS record name.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "DNS record type.",
							Computed:    true,
						},
						"content": schema.StringAttribute{
							Description: "DNS record content.",
							Computed:    true,
						},
						"ttl": schema.StringAttribute{
							Description: "DNS record TTL.",
							Computed:    true,
						},
						"cloud": schema.StringAttribute{
							Description: "GoCache cloud/proxy status.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *dnsRecordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dnsRecordsDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	records, err := d.client.ListDNSRecords(state.Domain.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing GoCache DNS records",
			err.Error(),
		)

		return
	}

	state.Records = make([]dnsRecordState, 0, len(records))

	for _, record := range records {
		state.Records = append(state.Records, dnsRecordState{
			ID:      types.StringValue(record.ID.String()),
			Name:    types.StringValue(record.Name),
			Type:    types.StringValue(record.Type),
			Content: types.StringValue(record.Content),
			TTL:     types.StringValue(record.TTL),
			Cloud:   types.StringValue(record.Cloud),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (d *dnsRecordsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*gocache.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *gocache.Client, got: %T.", req.ProviderData),
		)

		return
	}

	d.client = client
}
