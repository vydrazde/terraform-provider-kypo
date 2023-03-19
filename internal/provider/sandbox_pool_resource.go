package provider

import (
	"context"
	"errors"
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
var _ resource.Resource = &sandboxPoolResource{}
var _ resource.ResourceWithImportState = &sandboxPoolResource{}
var _ resource.ResourceWithConfigure = &sandboxPoolResource{}

func NewSandboxPoolResource() resource.Resource {
	return &sandboxPoolResource{}
}

// sandboxPoolResource defines the resource implementation.
type sandboxPoolResource struct {
	client *KYPOClient.Client
}

func (r *sandboxPoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sandbox_pool"
}

func (r *sandboxPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Sandbox pool",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Sandbox Pool Id",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"definition_id": schema.Int64Attribute{
				MarkdownDescription: "Id of associated sandbox definition",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "Current number of sandboxes",
				Computed:            true,
			},
			"max_size": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of sandboxes",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"lock_id": schema.Int64Attribute{
				MarkdownDescription: "Id of associated lock",
				Computed:            true,
			},

			"rev": schema.StringAttribute{
				MarkdownDescription: "Revision of the associated Git repository of the sandbox definition",
				Computed:            true,
			},
			"rev_sha": schema.StringAttribute{
				MarkdownDescription: "Revision hash of the Git repository of the sandbox definition",
				Computed:            true,
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
			"hardware_usage": schema.SingleNestedAttribute{
				MarkdownDescription: "Current resource usage",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"vcpu": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Id of the user",
					},
					"ram": schema.StringAttribute{
						MarkdownDescription: "TODO",
						Computed:            true,
					},
					"instances": schema.StringAttribute{
						MarkdownDescription: "TODO",
						Computed:            true,
					},
					"network": schema.StringAttribute{
						MarkdownDescription: "TODO",
						Computed:            true,
					},
					"subnet": schema.StringAttribute{
						MarkdownDescription: "TODO",
						Computed:            true,
					},
					"port": schema.StringAttribute{
						MarkdownDescription: "TODO",
						Computed:            true,
					},
				},
			},
			"definition": schema.SingleNestedAttribute{
				MarkdownDescription: "The associated sandbox definition",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"id": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "Example identifier",
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "Name of the sandbox definition",
						Computed:            true,
					},
					"url": schema.StringAttribute{
						MarkdownDescription: "Url to the Git repository of the sandbox definition",
						Computed:            true,
					},
					"rev": schema.StringAttribute{
						MarkdownDescription: "Revision of the Git repository of the sandbox definition",
						Computed:            true,
					},
					"created_by": schema.SingleNestedAttribute{
						MarkdownDescription: "Creator of this sandbox definition",
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
				},
			},
		},
	}
}

func (r *sandboxPoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sandboxPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var definitionId, maxSize int64

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("definition_id"), &definitionId)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("max_size"), &maxSize)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	pool, err := r.client.CreateSandboxPool(definitionId, maxSize)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}
	pool.DefinitionId = pool.Definition.Id
	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &pool)...)
}

func (r *sandboxPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var id int64

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	pool, err := r.client.GetSandboxPool(id)
	if errors.Is(err, KYPOClient.ErrNotFound) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}
	pool.DefinitionId = pool.Definition.Id

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &pool)...)
}

func (r *sandboxPoolResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

func (r *sandboxPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var id int64

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("id"), &id)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	err := r.client.DeleteSandboxPool(id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}
}

func (r *sandboxPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import pool, got error: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
