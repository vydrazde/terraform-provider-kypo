package plan_modifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/exp/slices"
)

var _ planmodifier.List = AllocationRequestStatePlanModifier{}

type AllocationRequestStatePlanModifier struct{}

func (r AllocationRequestStatePlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	var sandboxUnitAllocationStages []string
	req.State.GetAttribute(ctx, path.Root("allocation_request").AtName("stages"), &sandboxUnitAllocationStages)
	resp.RequiresReplace = slices.Contains(sandboxUnitAllocationStages, "FAILED")
	resp.PlanValue, _ = types.ListValueFrom(ctx, types.StringType, []string{"FINISHED", "FINISHED", "FINISHED"})
}

func (r AllocationRequestStatePlanModifier) Description(ctx context.Context) string {
	return r.MarkdownDescription(ctx)
}

func (r AllocationRequestStatePlanModifier) MarkdownDescription(_ context.Context) string {
	return "Replace is required when one of the stages is `FAILED`, update - which only waits for completion, " +
		"is required when all stages are not `FINISHED`"
}
