package provider

import (
	"context"
	"fmt"
	"terraform-provider-gocache/internal/client/gocache"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	DEFAULT_TTL   = 300
	DEFAULT_CLOUD = 0
)

type dnsRecordResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Domain      types.String `tfsdk:"domain"`
	Type        types.String `tfsdk:"type"`
	Name        types.String `tfsdk:"name"`
	Content     types.String `tfsdk:"content"`
	TTL         types.Int64  `tfsdk:"ttl"`
	LastUpdated types.String `tfsdk:"last_updated"`
	Cloud       types.Int64  `tfsdk:"cloud"`
}

var (
	_ resource.Resource                = &dnsRecordResource{}
	_ resource.ResourceWithConfigure   = &dnsRecordResource{}
	_ resource.ResourceWithImportState = &dnsRecordResource{}
)

func NewDNSRecordResource() resource.Resource {
	return &dnsRecordResource{}
}

type dnsRecordResource struct {
	client *gocache.Client
}

func (r *dnsRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *dnsRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a GoCache DNS record.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "DNS record identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Description: "Domain name.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "DNS record type, such as A, CNAME or TXT.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "DNS record name.",
				Required:    true,
			},
			"content": schema.StringAttribute{
				Description: "DNS record content.",
				Required:    true,
			},
			"ttl": schema.Int64Attribute{
				Description: "DNS record TTL.",
				Optional:    true,
				Computed:    true,
			},
			"last_updated": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update.",
				Computed:    true,
			},
			"cloud": schema.Int64Attribute{
				Description: "Whether the DNS record is served by GoCache. Use 1 for enabled and 0 for disabled.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *dnsRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsRecordResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if plan.TTL.IsNull() || plan.TTL.IsUnknown() {
		plan.TTL = types.Int64Value(DEFAULT_TTL)
	}

	if plan.Cloud.IsNull() || plan.Cloud.IsUnknown() {
		plan.Cloud = types.Int64Value(DEFAULT_CLOUD)
	}

	record := gocache.DNSRecordRequest{
		Type:    plan.Type.ValueString(),
		Name:    plan.Name.ValueString(),
		Content: plan.Content.ValueString(),
		TTL:     plan.TTL.ValueInt64(),
		Cloud:   plan.Cloud.ValueInt64(),
	}

	createdRecord, err := r.client.CreateDNSRecord(
		plan.Domain.ValueString(),
		record,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating GoCache DNS record",
			err.Error(),
		)

		return
	}

	plan.ID = types.StringValue(createdRecord.ID.String())

	plan.LastUpdated = types.StringValue(
		time.Now().Format(time.RFC850),
	)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *dnsRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsRecordResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	records, err := r.client.ListDNSRecords(state.Domain.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading GoCache DNS records",
			err.Error(),
		)

		return
	}

	for _, record := range records {
		recordID := record.ID.String()

		if recordID == state.ID.ValueString() {
			state.Type = types.StringValue(record.Type)
			state.Name = types.StringValue(record.Name)
			state.Content = types.StringValue(record.Content)

			diags = resp.State.Set(ctx, &state)
			resp.Diagnostics.Append(diags...)

			return
		}
	}

	resp.State.RemoveResource(ctx)
}

func (r *dnsRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dnsRecordResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if plan.TTL.IsNull() || plan.TTL.IsUnknown() {
		plan.TTL = types.Int64Value(DEFAULT_TTL)
	}

	if plan.Cloud.IsNull() || plan.Cloud.IsUnknown() {
		plan.Cloud = types.Int64Value(DEFAULT_CLOUD)
	}

	record := gocache.DNSRecordRequest{
		Type:    plan.Type.ValueString(),
		Name:    plan.Name.ValueString(),
		Content: plan.Content.ValueString(),
		TTL:     plan.TTL.ValueInt64(),
		Cloud:   plan.Cloud.ValueInt64(),
	}

	err := r.client.UpdateDNSRecord(
		plan.Domain.ValueString(),
		plan.ID.ValueString(),
		record,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating GoCache DNS record",
			err.Error(),
		)

		return
	}

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *dnsRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnsRecordResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDNSRecord(
		state.Domain.ValueString(),
		state.ID.ValueString(),
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting GoCache DNS record",
			err.Error(),
		)

		return
	}
}

func (r *dnsRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *dnsRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
