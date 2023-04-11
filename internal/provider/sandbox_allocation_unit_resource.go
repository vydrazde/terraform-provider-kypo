package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strconv"
	"terraform-provider-kypo/internal/KYPOClient"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &sandboxAllocationUnitResource{}
var _ resource.ResourceWithImportState = &sandboxAllocationUnitResource{}
var _ resource.ResourceWithConfigure = &sandboxAllocationUnitResource{}

func NewSandboxAllocationUnitResource() resource.Resource {
	return &sandboxAllocationUnitResource{}
}

// sandboxAllocationUnitResource defines the resource implementation.
type sandboxAllocationUnitResource struct {
	client *KYPOClient.Client
}

func (r *sandboxAllocationUnitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sandbox_allocation_unit"
}

func (r *sandboxAllocationUnitResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Sandbox allocation unit",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Sandbox Allocation Unit Id",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"pool_id": schema.Int64Attribute{
				MarkdownDescription: "Id of associated sandbox pool",
				Required:            true,
			},
			"allocation_request": schema.SingleNestedAttribute{
				MarkdownDescription: "Associated allocation request",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Id of the allocation request",
					},
					"allocation_unit_id": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Id of the associated allocation unit",
					},
					"created": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "TODO",
					},
					"stages": schema.ListAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "TODO",
					},
				},
			},
			"cleanup_request": schema.SingleNestedAttribute{
				MarkdownDescription: "Associated cleanup request",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Id of the cleanup request",
					},
					"allocation_unit_id": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Id of the associated allocation unit",
					},
					"created": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "TODO",
					},
					"stages": schema.ListAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "TODO",
					},
				},
			},
			"created_by": schema.SingleNestedAttribute{
				MarkdownDescription: "Creator of this sandbox pool",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Id of the user",
					},
					"sub": schema.StringAttribute{
						MarkdownDescription: "TODO",
						Computed:            true,
					},
					"full_name": schema.StringAttribute{
						MarkdownDescription: "TODO",
						Computed:            true,
					},
					"given_name": schema.StringAttribute{
						MarkdownDescription: "TODO",
						Computed:            true,
					},
					"family_name": schema.StringAttribute{
						MarkdownDescription: "TODO",
						Computed:            true,
					},
					"mail": schema.StringAttribute{
						MarkdownDescription: "TODO",
						Computed:            true,
					},
				},
			},
			"locked": schema.BoolAttribute{
				MarkdownDescription: "Revision hash of the Git repository of the sandbox definition",
				Computed:            true,
			},
		},
	}
}

func (r *sandboxAllocationUnitResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*KYPOClient.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected KYPOClient.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}
	r.client = client
}

func (r *sandboxAllocationUnitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var poolId int64

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("pool_id"), &poolId)...)

	if resp.Diagnostics.HasError() {
		return
	}

	allocationUnits, err := r.client.CreateSandboxAllocationUnits(poolId, 1)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create sandbox allocation unit, got error: %s", err))
		return
	}
	allocationUnit := allocationUnits[0]
	resp.Diagnostics.Append(resp.State.Set(ctx, allocationUnit)...)
	if resp.Diagnostics.HasError() {
		return
	}

	allocationRequest, err := r.client.PollRequestFinished(allocationUnit.AllocationRequest.Id, 5*time.Second, "allocation")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("awaiting allocation request failed, got error: %s", err))
		return
	}
	allocationUnit.AllocationRequest = *allocationRequest
	resp.Diagnostics.Append(resp.State.Set(ctx, allocationUnit)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if allocationUnit.AllocationRequest.Stages[0] != "FINISHED" {
		resp.Diagnostics.AddError("Sandbox Creation Error", fmt.Sprintf("Creation of sandbox allocation unit %d finished with error in Terraform stage", poolId))
		return
	}
	if allocationUnit.AllocationRequest.Stages[1] != "FINISHED" {
		resp.Diagnostics.AddError("Sandbox Creation Error", fmt.Sprintf("Creation of sandbox allocation unit %d finished with error in Networking Ansible stage", poolId))
		return
	}
	if allocationUnit.AllocationRequest.Stages[2] != "FINISHED" {
		resp.Diagnostics.AddError("Sandbox Creation Error", fmt.Sprintf("Creation of sandbox allocation unit %d finished with error in User Ansible stage", poolId))
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, fmt.Sprintf("created sandbox allocation unit %d", allocationUnit.Id))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, allocationUnit)...)
}

func (r *sandboxAllocationUnitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var id int64

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	allocationUnit, err := r.client.GetSandboxAllocationUnit(id)
	if _, ok := err.(*KYPOClient.ErrNotFound); ok {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read sandbox allocation unit, got error: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &allocationUnit)...)
}

func (r *sandboxAllocationUnitResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

func (r *sandboxAllocationUnitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var id int64

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	err := r.client.CreateSandboxCleanupRequestAwait(id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete sandbox allocation unit, got error: %s", err))
		return
	}
}

func (r *sandboxAllocationUnitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import sandbox allocation unit, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
