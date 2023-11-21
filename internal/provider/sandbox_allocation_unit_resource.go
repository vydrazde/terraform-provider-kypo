package provider

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vydrazde/kypo-go-client/pkg/kypo"
	"golang.org/x/exp/slices"
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
	client *kypo.Client
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

func checkAllocationRequestResult(allocationUnit *kypo.SandboxAllocationUnit, diagnostics *diag.Diagnostics, warningOnAllocationFailureBool bool, id int64) {
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

func setTimeout(diags *diag.Diagnostics, ctx context.Context, timeoutsValue timeouts.Value, timeoutName string) (context.Context, context.CancelFunc) {
	value, ok := timeoutsValue.Object.Attributes()[timeoutName]
	if !ok || value.IsNull() || value.IsUnknown() {
		tflog.Info(ctx, timeoutName+" timeout configuration not found, null or unknown, no timeout will be set")
		return ctx, func() {}
	}

	timeout, err := time.ParseDuration(value.(types.String).ValueString())
	if err != nil {
		diags.AddError("Timeout Cannot Be Parsed",
			fmt.Sprintf("timeout for %q cannot be parsed, %s", timeoutName, err),
		)

		return ctx, func() {}
	}

	return context.WithTimeout(ctx, timeout)
}

func (r *sandboxAllocationUnitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sandbox_allocation_unit"
}

func (r *sandboxAllocationUnitResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Sandbox allocation unit",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Id of the sandbox allocation unit",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"pool_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the associated sandbox pool",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"allocation_request": schema.SingleNestedAttribute{
				MarkdownDescription: "Allocation request of the allocation unit",
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
						MarkdownDescription: "Date and time when the allocation request was created",
					},
					"stages": schema.ListAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Statuses of the allocation stages. List of three strings, where each is one of `IN_QUEUE`, `FINISHED`, `FAILED` or `RUNNING`",
						PlanModifiers: []planmodifier.List{
							allocationUnitStatePlanModifier{},
						},
					},
				},
			},
			"cleanup_request": schema.SingleNestedAttribute{
				MarkdownDescription: "Cleanup request of the allocation unit",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Id of the cleanup request",
					},
					"allocation_unit_id": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Id of the allocation unit",
					},
					"created": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Date and time when the allocation request was created",
					},
					"stages": schema.ListAttribute{
						Computed:            true,
						ElementType:         types.StringType,
						MarkdownDescription: "Statuses of cleanup stages. List of three strings, where each is one of `IN_QUEUE`, `FINISHED`, `FAILED` or `RUNNING`",
					},
				},
			},
			"created_by": schema.SingleNestedAttribute{
				MarkdownDescription: "Who created the sandbox allocation unit",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Id of the user",
					},
					"sub": schema.StringAttribute{
						MarkdownDescription: "Sub of the user as given by an OIDC provider",
						Computed:            true,
					},
					"full_name": schema.StringAttribute{
						MarkdownDescription: "Full name of the user",
						Computed:            true,
					},
					"given_name": schema.StringAttribute{
						MarkdownDescription: "Given name of the user",
						Computed:            true,
					},
					"family_name": schema.StringAttribute{
						MarkdownDescription: "Family name of the user",
						Computed:            true,
					},
					"mail": schema.StringAttribute{
						MarkdownDescription: "Email of the user",
						Computed:            true,
					},
				},
			},
			"locked": schema.BoolAttribute{
				MarkdownDescription: "Whether the allocation unit is locked. The allocation unit is locked when it is claimed by a Trainee and has an associated training run",
				Computed:            true,
			},
			"warning_on_allocation_failure": schema.BoolAttribute{
				MarkdownDescription: "Whether to emit a warning instead of error when one of the allocation request stages fails",
				Optional:            true,
			},
			"timeouts": timeouts.AttributesAll(ctx),
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

	client, ok := req.ProviderData.(*kypo.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected kypo.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}
	r.client = client
}

func (r *sandboxAllocationUnitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var poolId int64
	var timeoutsValue timeouts.Value

	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("pool_id"), &poolId)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("timeouts"), &timeoutsValue)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("timeouts"), timeoutsValue)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := setTimeout(&resp.Diagnostics, ctx, timeoutsValue, "create")
	defer cancel()

	if resp.Diagnostics.HasError() {
		return
	}

	allocationUnits, err := r.client.CreateSandboxAllocationUnits(ctx, poolId, 1)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create sandbox allocation unit, got error: %s", err))
		return
	}
	allocationUnit := allocationUnits[0]
	setState(ctx, allocationUnit, response{State: &resp.State, Diagnostics: &resp.Diagnostics})
	if resp.Diagnostics.HasError() {
		return
	}

	allocationRequest, err := r.client.PollRequestFinished(ctx, allocationUnit.AllocationRequest.Id, 5*time.Second, "allocation")
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
	var timeoutsValue timeouts.Value

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("timeouts"), &timeoutsValue)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := setTimeout(&resp.Diagnostics, ctx, timeoutsValue, "read")
	defer cancel()

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	allocationUnit, err := r.client.GetSandboxAllocationUnit(ctx, id)
	if errors.Is(err, kypo.ErrNotFound) {
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
	var timeoutsValue timeouts.Value

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("warning_on_allocation_failure"), &stateWarningOnAllocationFailure)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("warning_on_allocation_failure"), &planWarningOnAllocationFailure)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("allocation_request"), &planAllocationRequest)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("timeouts"), &timeoutsValue)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("timeouts"), timeoutsValue)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := setTimeout(&resp.Diagnostics, ctx, timeoutsValue, "update")
	defer cancel()

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

	allocationUnit, err := r.client.GetSandboxAllocationUnit(ctx, id.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read sandbox allocation unit, got error: %s", err))
		return
	}

	allocationRequest, err := r.client.PollRequestFinished(ctx, allocationUnit.AllocationRequest.Id, 5*time.Second, "allocation")
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
	var allocationRequest *kypo.SandboxRequest
	var id int64
	var timeoutsValue timeouts.Value

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("allocation_request"), &allocationRequest)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("timeouts"), &timeoutsValue)...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := setTimeout(&resp.Diagnostics, ctx, timeoutsValue, "delete")
	defer cancel()

	if resp.Diagnostics.HasError() {
		return
	}

	if slices.Contains(allocationRequest.Stages, "RUNNING") {
		err := r.client.CancelSandboxAllocationRequest(ctx, allocationRequest.Id)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to cancel sandbox allocation unit allocation request, got error: %s", err))
			return
		}
	}

	err := r.client.CreateSandboxCleanupRequestAwait(ctx, id, 5*time.Second)
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
