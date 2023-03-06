package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"os"
	"terraform-provider-kypo/internal/KYPOClient"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure KypoProvider satisfies various provider interfaces.
var _ provider.Provider = &KypoProvider{}

// KypoProvider defines the provider implementation.
type KypoProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// KypoProviderModel describes the provider data model.
type KypoProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
}

func (p *KypoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kypo"
	resp.Version = p.version
}

func (p *KypoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "URI",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "JSON token",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *KypoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data KypoProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.
	if data.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown KYPO API Endpoint",
			"The provider cannot create the KYPO API client as there is an unknown configuration value for the KYPO API endpoint. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the KYPO_ENDPOINT environment variable.",
		)
	}
	if data.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown KYPO API Token",
			"The provider cannot create the KYPO API client as there is an unknown configuration value for the KYPO API token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the KYPO_TOKEN environment variable.",
		)
	}

	endpoint := os.Getenv("KYPO_ENDPOINT")
	token := os.Getenv("KYPO_TOKEN")

	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	}
	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing KYPO API Endpoint",
			"The provider cannot create the KYPO API client as there is a missing or empty value for the KYPO API endpoint. "+
				"Set the host value in the configuration or use the KYPO_ENDPOINT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing KYPO API Token",
			"The provider cannot create the KYPO API client as there is a missing or empty value for the KYPO API token. "+
				"Set the host value in the configuration or use the KYPO_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "kypo_endpoint", endpoint)
	ctx = tflog.SetField(ctx, "kypo_token", token)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "kypo_token")

	tflog.Debug(ctx, "Creating KYPO client")

	client, err := KYPOClient.NewClient(endpoint, token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create KYPO API Client",
			"An unexpected error occurred when creating the KYPO API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"KYPO Client Error: "+err.Error(),
		)
		return
	}
	resp.DataSourceData = client
	resp.ResourceData = client
	tflog.Info(ctx, "Configured KYPO client", map[string]any{"success": true})
}

func (p *KypoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDefinitionsResource,
	}
}

func (p *KypoProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		//NewExampleDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &KypoProvider{
			version: version,
		}
	}
}
