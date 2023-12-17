package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vydrazde/kypo-go-client/pkg/kypo"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &sandboxRequestOutputDataSource{}
	_ datasource.DataSourceWithConfigure = &sandboxRequestOutputDataSource{}
)

const int64MaxValue = 9223372036854775807

// NewSandboxRequestOutputDataSource is a helper function to simplify the provider implementation.
func NewSandboxRequestOutputDataSource() datasource.DataSource {
	return &sandboxRequestOutputDataSource{}
}

// sandboxRequestOutputDataSource is the data source implementation.
type sandboxRequestOutputDataSource struct {
	client *kypo.Client
}

type sandboxRequestOutput struct {
	Id     types.Int64  `json:"id" tfsdk:"id"`
	Stage  types.String `json:"stage" tfsdk:"stage"`
	Result types.String `json:"result" tfsdk:"result"`
}

// Metadata returns the data source type name.
func (r *sandboxRequestOutputDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sandbox_request_output"
}

// Schema defines the schema for the data source.
func (r *sandboxRequestOutputDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sandbox allocation request output of one of three stages, which are terraform, networking-ansible or user-ansible.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Id of the sandbox allocation request to read the output from. The sandbox allocation request is always the same as sandbox allocation unit id",
			},
			"stage": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Sandbox request stage to get the output of. Must be one of `user-ansible`, `networking-ansible` or `terraform`. Defaults to `user-ansible`",
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"user-ansible", "networking-ansible", "terraform"}...),
				},
			},
			"result": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The resulting output of the stage, concatenated into a single string",
			},
		},
	}
}

func (r *sandboxRequestOutputDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *sandboxRequestOutputDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var requestOutput sandboxRequestOutput

	resp.Diagnostics.Append(req.Config.Get(ctx, &requestOutput)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if requestOutput.Stage.IsUnknown() || requestOutput.Stage.IsNull() {
		requestOutput.Stage = types.StringValue("user-ansible")
	}

	clientRequestOutput, err := r.client.GetSandboxRequestAnsibleOutputs(ctx, requestOutput.Id.ValueInt64(),
		1, int64MaxValue, requestOutput.Stage.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read sandbox request output, got error: %s", err))
		return
	}

	requestOutput.Result = types.StringValue(clientRequestOutput.Result)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &requestOutput)...)
}
