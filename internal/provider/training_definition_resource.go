package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strconv"
	"terraform-provider-kypo/internal/KYPOClient"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &trainingDefinitionResource{}
var _ resource.ResourceWithImportState = &trainingDefinitionResource{}
var _ resource.ResourceWithConfigure = &trainingDefinitionResource{}

func NewTrainingDefinitionResource() resource.Resource {
	return &trainingDefinitionResource{}
}

// trainingDefinitionResource defines the resource implementation.
type trainingDefinitionResource struct {
	client *KYPOClient.Client
}

func (r *trainingDefinitionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_training_definition"
}

func (r *trainingDefinitionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Training definition",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Id of the training definition",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "JSON with exported training definition",
				Required:            true,
			},
		},
	}
}

func (r *trainingDefinitionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *trainingDefinitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var url, rev string

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("url"), &url)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("rev"), &rev)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	definition, err := r.client.CreateSandboxDefinition(url, rev)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create sandbox definition, got error: %s", err))
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, fmt.Sprintf("created sandbox definition %d", definition.Id))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &definition)...)
}

func (r *trainingDefinitionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var id int64

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	definition, err := r.client.GetSandboxDefinition(id)
	if _, ok := err.(*KYPOClient.ErrNotFound); ok {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read sandbox definition, got error: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &definition)...)
}

func (r *trainingDefinitionResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

func (r *trainingDefinitionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var id int64

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	err := r.client.DeleteSandboxDefinition(id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete sandbox definition, got error: %s", err))
		return
	}
}

func (r *trainingDefinitionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import sandbox definition, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
