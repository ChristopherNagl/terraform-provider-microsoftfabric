package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"terraform-microsoft-fabric/internal/apiclient"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &microsoftFabricProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &microsoftFabricProvider{

			version: version,
		}
	}
}

// microsoftFabricProvider is the provider implementation.
type microsoftFabricProvider struct {
	version string
	client  *apiclient.APIClient
}

// Metadata returns the provider type name.
func (p *microsoftFabricProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "microsoftFabric"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *microsoftFabricProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Required:    true,
				Description: "The Client ID for Power BI API access.",
			},
			"client_secret": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The Client Secret for Power BI API access.",
			},
			"tenant_id": schema.StringAttribute{
				Required:    true,
				Description: "The Tenant ID for Power BI API access.",
			},
			"token_file_path": schema.StringAttribute{
				Optional:    true,
				Description: "The path to the token file.",
			},
		},
	}
}

// Configure prepares a HashiCups API client for data sources and resources.
func (p *microsoftFabricProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config struct {
		ClientID      string `tfsdk:"client_id"`
		ClientSecret  string `tfsdk:"client_secret"`
		TenantID      string `tfsdk:"tenant_id"`
		TokenFilePath string `tfsdk:"token_file_path"`
	}

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Initialize the API client with all four parameters
	p.client = apiclient.NewAPIClient(config.ClientID, config.ClientSecret, config.TenantID, config.TokenFilePath)
}

// DataSources defines the data sources implemented in the provider.
func (p *microsoftFabricProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// Resources defines the resources implemented in the provider.
func (p *microsoftFabricProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return NewWorkspaceResource(p.client) },
		func() resource.Resource { return NewWorkspaceUserAssignmentResource(p.client) },
		func() resource.Resource { return NewWorkspaceCapacityAssignmentResource(p.client) },
		func() resource.Resource { return NewEventStreamResource(p.client) },
		func() resource.Resource { return NewWorkspaceGitResource(p.client) },
		func() resource.Resource { return NewMLEexperimentResource(p.client) },
		func() resource.Resource { return NewEventhouseResource(p.client) },
		func() resource.Resource { return NewPipelineResource(p.client) },
		func() resource.Resource { return NewSemanticModelUserAssignmentResource(p.client) }, 
		func() resource.Resource { return NewDomainResource(p.client) },
	}
}
