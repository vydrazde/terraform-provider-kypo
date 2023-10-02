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
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Token    types.String `tfsdk:"token"`
	ClientID types.String `tfsdk:"client_id"`
}

func (p *KypoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kypo"
	resp.Version = p.version
}

func (p *KypoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "URI of the homepage of the KYPO instance, like `https://my.kypo.instance.ex`. Can be set with `KYPO_ENDPOINT` environmental variable.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "`username` of the user to login as with `password`. Use either `username` and `password` or just `token`. Can be set with `KYPO_USERNAME` environmental variable.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "`password` of the user to login as with `username`. Use either `username` and `password` or just `token`. Can be set with `KYPO_PASSWORD` environmental variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Bearer token to be used. Takes precedence before `username` and `password`. Bearer tokens usually have limited lifespan. Can be set with `KYPO_TOKEN` environmental variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "KYPO local OIDC client ID. Will be ignored when `token` is set. Defaults to `KYPO-Client`. Can be set with `KYPO_CLIENT_ID` environmental variable. See [how to get KYPO client_id](https://github.com/vydrazde/terraform-provider-kypo/wiki/How-to-get-KYPO-CRP-client_id).",
				Optional:            true,
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
	if data.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown KYPO API Username",
			"The provider cannot create the KYPO API client as there is an unknown configuration value for the KYPO API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the KYPO_USERNAME environment variable.",
		)
	}
	if data.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown KYPO API Password",
			"The provider cannot create the KYPO API client as there is an unknown configuration value for the KYPO API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the KYPO_PASSWORD environment variable.",
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
	if data.ClientID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Unknown KYPO API Client ID",
			"The provider cannot create the KYPO API client as there is an unknown configuration value for the KYPO API client ID. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the KYPO_CLIENT_ID environment variable.",
		)
	}

	endpoint := os.Getenv("KYPO_ENDPOINT")
	username := os.Getenv("KYPO_USERNAME")
	password := os.Getenv("KYPO_PASSWORD")
	token := os.Getenv("KYPO_TOKEN")
	clientId := os.Getenv("KYPO_CLIENT_ID")

	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	}
	if !data.Username.IsNull() {
		username = data.Username.ValueString()
	}
	if !data.Password.IsNull() {
		password = data.Password.ValueString()
	}
	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}
	if !data.ClientID.IsNull() {
		clientId = data.ClientID.ValueString()
	}

	if clientId == "" {
		clientId = "KYPO-Client"
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
	if token == "" && (username == "" || password == "") {
		resp.Diagnostics.AddError(
			"Missing KYPO API Token or Username and Password",
			"The provider cannot create the KYPO API client as there is a missing or empty value for the KYPO API token or username and password. "+
				"Set the host value in the configuration or use the KYPO_TOKEN, KYPO_USERNAME and KYPO_PASSWORD environment variables. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "kypo_endpoint", endpoint)
	ctx = tflog.SetField(ctx, "kypo_username", username)
	ctx = tflog.SetField(ctx, "kypo_password", password)
	ctx = tflog.SetField(ctx, "kypo_token", token)
	ctx = tflog.SetField(ctx, "client_id", clientId)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "kypo_password", "kypo_token")

	tflog.Debug(ctx, "Creating KYPO client")
	var client *KYPOClient.Client

	var err error
	if token != "" {
		client, err = KYPOClient.NewClientWithToken(endpoint, clientId, token)
	} else {
		client, err = KYPOClient.NewClient(endpoint, clientId, username, password)
	}
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
		NewSandboxDefinitionResource,
		NewSandboxPoolResource,
		NewSandboxAllocationUnitResource,
		NewTrainingDefinitionResource,
		NewTrainingDefinitionAdaptiveResource,
	}
}

func (p *KypoProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSandboxRequestOutputDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &KypoProvider{
			version: version,
		}
	}
}
