package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/slices"
	"reflect"
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

type response struct {
	State       *tfsdk.State
	Diagnostics *diag.Diagnostics
}

func setState(ctx context.Context, stateValue any, resp response) {
	valueOf := reflect.ValueOf(stateValue)
	typeOf := reflect.TypeOf(stateValue)

	for i := 0; i < valueOf.NumField(); i++ {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(typeOf.Field(i).Tag.Get("tfsdk")), valueOf.Field(i).Interface())...)
	}
}

func checkAllocationRequestResult(allocationUnit *KYPOClient.SandboxAllocationUnit, diagnostics *diag.Diagnostics, warningOnAllocationFailureBool bool, id int64) {
	if allocationUnit.AllocationRequest.Stages[0] != "FINISHED" {
		warningOrError(diagnostics, warningOnAllocationFailureBool, "Sandbox Creation Error - Terraform Stage Failed",
			fmt.Sprintf("Creation of sandbox allocation unit %d finished with error in Terraform stage", id))
		return
	}
	if allocationUnit.AllocationRequest.Stages[1] != "FINISHED" {
		warningOrError(diagnostics, warningOnAllocationFailureBool, "Sandbox Creation Error - Ansible Stage Failed",
			fmt.Sprintf("Creation of sandbox allocation unit %d finished with error in Networking Ansible stage", id))
		return
	}
	if allocationUnit.AllocationRequest.Stages[2] != "FINISHED" {
		warningOrError(diagnostics, warningOnAllocationFailureBool, "Sandbox Creation Error - User Stage Failed",
			fmt.Sprintf("Creation of sandbox allocation unit %d finished with error in User Ansible stage", id))
		return
	}
}

func warningOrError(diagnostics *diag.Diagnostics, warning bool, summary, errorString string) {
	if warning {
		diagnostics.AddWarning(summary, errorString)
	} else {
		diagnostics.AddError(summary, errorString)
	}
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
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
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
						PlanModifiers: []planmodifier.List{
							allocationUnitStatePlanModifier{},
						},
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
				MarkdownDescription: "TODO",
				Computed:            true,
			},
			"warning_on_allocation_failure": schema.BoolAttribute{
				MarkdownDescription: "If `true`, will emit a warning instead of error when one of the allocation " +
					"request stages fails.",
				Optional: true,
			},
		},
	}
}

type allocationUnitStatePlanModifier struct{}

func (r allocationUnitStatePlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	var sandboxUnitAllocationStages []string
	req.State.GetAttribute(ctx, path.Root("allocation_request").AtName("stages"), &sandboxUnitAllocationStages)
	resp.RequiresReplace = slices.Contains(sandboxUnitAllocationStages, "FAILED")
	resp.PlanValue, _ = types.ListValueFrom(ctx, types.StringType, []string{"FINISHED", "FINISHED", "FINISHED"})
}

func (r allocationUnitStatePlanModifier) Description(ctx context.Context) string {
	return r.MarkdownDescription(ctx)
}

func (r allocationUnitStatePlanModifier) MarkdownDescription(_ context.Context) string {
	return "Replace is required when one of the stages is `FAILED`, update - which only waits for completion, " +
		"is required when all stages are not `FINISHED`"
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
	setState(ctx, allocationUnit, response{State: &resp.State, Diagnostics: &resp.Diagnostics})
	if resp.Diagnostics.HasError() {
		return
	}

	allocationRequest, err := r.client.PollRequestFinished(allocationUnit.AllocationRequest.Id, 5*time.Second, "allocation")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("awaiting allocation request failed, got error: %s", err))
		return
	}
	allocationUnit.AllocationRequest = *allocationRequest
	setState(ctx, allocationUnit, response{State: &resp.State, Diagnostics: &resp.Diagnostics})
	if resp.Diagnostics.HasError() {
		return
	}

	var warningOnAllocationFailure types.Bool
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("warning_on_allocation_failure"), &warningOnAllocationFailure)...)
	if resp.Diagnostics.HasError() {
		return
	}
	warningOnAllocationFailureBool := warningOnAllocationFailure.Equal(types.BoolValue(true))

	if warningOnAllocationFailureBool {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("warning_on_allocation_failure"), warningOnAllocationFailureBool)...)
	}

	checkAllocationRequestResult(&allocationUnit, &resp.Diagnostics, warningOnAllocationFailureBool, allocationUnit.Id)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, fmt.Sprintf("created sandbox allocation unit %d", allocationUnit.Id))
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

	setState(ctx, *allocationUnit, response{State: &resp.State, Diagnostics: &resp.Diagnostics})
}

func (r *sandboxAllocationUnitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var id types.Int64
	var stateWarningOnAllocationFailure, planWarningOnAllocationFailure types.Bool
	var planAllocationRequest types.Object

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("warning_on_allocation_failure"), &stateWarningOnAllocationFailure)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("warning_on_allocation_failure"), &planWarningOnAllocationFailure)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("allocation_request"), &planAllocationRequest)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !stateWarningOnAllocationFailure.Equal(planWarningOnAllocationFailure) {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("warning_on_allocation_failure"), planWarningOnAllocationFailure)...)
		if resp.Diagnostics.HasError() {
			return
		}

	}

	if planAllocationRequest.IsNull() || planAllocationRequest.IsUnknown() {
		return
	}

	allocationUnit, err := r.client.GetSandboxAllocationUnit(id.ValueInt64())
	if _, ok := err.(*KYPOClient.ErrNotFound); ok {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read sandbox allocation unit, got error: %s", err))
		return
	}

	allocationRequest, err := r.client.PollRequestFinished(allocationUnit.AllocationRequest.Id, 5*time.Second, "allocation")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("awaiting allocation request failed, got error: %s", err))
		return
	}
	allocationUnit.AllocationRequest = *allocationRequest
	setState(ctx, *allocationUnit, response{State: &resp.State, Diagnostics: &resp.Diagnostics})
	if resp.Diagnostics.HasError() {
		return
	}

	warningOnAllocationFailureBool := planWarningOnAllocationFailure.Equal(types.BoolValue(true))

	checkAllocationRequestResult(allocationUnit, &resp.Diagnostics, warningOnAllocationFailureBool, id.ValueInt64())
}

func (r *sandboxAllocationUnitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var allocationRequest *KYPOClient.SandboxRequest
	var id int64

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("allocation_request"), &allocationRequest)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if slices.Contains(allocationRequest.Stages, "RUNNING") {
		err := r.client.CancelSandboxAllocationRequest(allocationRequest.Id)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to cancel sandbox allocation unit allocation request, got error: %s", err))
			return
		}
	}

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
