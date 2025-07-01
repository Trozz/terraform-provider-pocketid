package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/Trozz/terraform-provider-pocketid/internal/client"
	"github.com/Trozz/terraform-provider-pocketid/internal/datasources"
	"github.com/Trozz/terraform-provider-pocketid/internal/resources"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &pocketIDProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &pocketIDProvider{
			version: version,
		}
	}
}

// pocketIDProvider is the provider implementation.
type pocketIDProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// pocketIDProviderModel maps provider schema data to a Go type.
type pocketIDProviderModel struct {
	BaseURL       types.String `tfsdk:"base_url"`
	APIToken      types.String `tfsdk:"api_token"`
	SkipTLSVerify types.Bool   `tfsdk:"skip_tls_verify"`
	Timeout       types.Int64  `tfsdk:"timeout"`
}

// Metadata returns the provider type name.
func (p *pocketIDProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pocketid"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *pocketIDProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Pocket-ID OIDC provider. Pocket-ID is a simple and easy-to-use OIDC provider that allows users to authenticate with their passkeys to your services.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Description: "Base URL of the Pocket-ID instance. Can also be set via POCKETID_BASE_URL environment variable.",
				Optional:    true,
			},
			"api_token": schema.StringAttribute{
				Description: "API token for authentication. Can also be set via POCKETID_API_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"skip_tls_verify": schema.BoolAttribute{
				Description: "Skip TLS certificate verification. Default is false. Only use this for development/testing.",
				Optional:    true,
			},
			"timeout": schema.Int64Attribute{
				Description: "HTTP client timeout in seconds. Default is 30.",
				Optional:    true,
			},
		},
	}
}

// Configure prepares a Pocket-ID API client for data sources and resources.
func (p *pocketIDProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Pocket-ID provider")

	// Retrieve provider data from configuration
	var config pocketIDProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.
	if config.BaseURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Unknown Pocket-ID Base URL",
			"The provider cannot create the Pocket-ID client as there is an unknown configuration value for the Pocket-ID base URL. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the POCKETID_BASE_URL environment variable.",
		)
	}

	if config.APIToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Unknown Pocket-ID API Token",
			"The provider cannot create the Pocket-ID client as there is an unknown configuration value for the Pocket-ID API token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the POCKETID_API_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	baseURL := os.Getenv("POCKETID_BASE_URL")
	apiToken := os.Getenv("POCKETID_API_TOKEN")

	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	if !config.APIToken.IsNull() {
		apiToken = config.APIToken.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if baseURL == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Missing Pocket-ID Base URL",
			"The provider cannot create the Pocket-ID client as there is a missing or empty value for the Pocket-ID base URL. "+
				"Set the base_url value in the configuration or use the POCKETID_BASE_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if apiToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Missing Pocket-ID API Token",
			"The provider cannot create the Pocket-ID client as there is a missing or empty value for the Pocket-ID API token. "+
				"Set the api_token value in the configuration or use the POCKETID_API_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Extract optional configuration
	skipTLSVerify := false
	if !config.SkipTLSVerify.IsNull() {
		skipTLSVerify = config.SkipTLSVerify.ValueBool()
	}

	timeout := int64(30)
	if !config.Timeout.IsNull() {
		timeout = config.Timeout.ValueInt64()
	}

	ctx = tflog.SetField(ctx, "pocketid_base_url", baseURL)
	ctx = tflog.SetField(ctx, "pocketid_skip_tls_verify", skipTLSVerify)
	ctx = tflog.SetField(ctx, "pocketid_timeout", timeout)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "pocketid_api_token")

	tflog.Debug(ctx, "Creating Pocket-ID client")

	// Create a new Pocket-ID client using the configuration values
	client, err := client.NewClient(baseURL, apiToken, skipTLSVerify, timeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Pocket-ID Client",
			"An unexpected error occurred when creating the Pocket-ID client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Pocket-ID Client Error: "+err.Error(),
		)
		return
	}

	// Make the Pocket-ID client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Pocket-ID client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *pocketIDProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewClientDataSource,
		datasources.NewClientsDataSource,
		datasources.NewUserDataSource,
		datasources.NewUsersDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *pocketIDProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewClientResource,
		resources.NewUserResource,
		resources.NewGroupResource,
	}
}
