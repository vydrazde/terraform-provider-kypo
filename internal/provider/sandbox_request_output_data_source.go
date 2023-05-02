package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-kypo/internal/KYPOClient"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &sandboxRequestOutputDataSource{}
	_ datasource.DataSourceWithConfigure = &sandboxRequestOutputDataSource{}
)

// NewSandboxRequestOutputDataSource is a helper function to simplify the provider implementation.
func NewSandboxRequestOutputDataSource() datasource.DataSource {
	return &sandboxRequestOutputDataSource{}
}

// sandboxRequestOutputDataSource is the data source implementation.
type sandboxRequestOutputDataSource struct {
	client *KYPOClient.Client
}

// Metadata returns the data source type name.
func (r *sandboxRequestOutputDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sandbox_request_output"
}

// Schema defines the schema for the data source.
func (r *sandboxRequestOutputDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sandbox allocation request output of one of three stages. Terraform, Networking Ansible or User Ansible.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Sandbox Request Id",
			},
			"stage": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "Sandbox Request stage to get the output of. Must be one of `user-ansible`, " +
					"`networking-ansible` or `terraform`. Defaults to `user-ansible`.",
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"user-ansible", "networking-ansible", "terraform"}...),
				},
			},
			"page": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Page number",
			},
			"page_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Number of lines in page",
			},
			"page_count": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Number of pages",
			},
			"count": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Number of lines in results",
			},
			"total_count": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Total number of lines",
			},
			"results": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Array of single lines of output",
			},
		},
	}
}

func (r *sandboxRequestOutputDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *sandboxRequestOutputDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

}
