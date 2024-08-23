package provider

import (
	"context"

	"terraform-provider-microsoftfabric/internal/apiclient"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
				Description: "The Client ID for Fabric API access.",
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The Client Secret for Fabric API access.",
			},
			"tenant_id": schema.StringAttribute{
				Required:    true,
				Description: "The Tenant ID for Fabric API access.",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "The username for Fabric API access.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The password for Fabric API access.",
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
		ClientID      string       `tfsdk:"client_id"`
		ClientSecret  types.String `tfsdk:"client_secret"`
		TenantID      string       `tfsdk:"tenant_id"`
		Username      types.String `tfsdk:"username"`
		Password      types.String `tfsdk:"password"`
		TokenFilePath types.String `tfsdk:"token_file_path"` // Use types.String for optional value
	}

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tokenFilePath := ""
	if config.TokenFilePath.IsNull() {
		// If token_file_path is optional and null, set it to an empty string or handle as necessary
		tokenFilePath = ""
	} else {
		tokenFilePath = config.TokenFilePath.ValueString() // Get the string value from the types package
	}

	// Initialize the API client with all required parameters
	p.client = apiclient.NewAPIClient(config.ClientID, config.ClientSecret.ValueString(), config.TenantID, config.Username.ValueString(), config.Password.ValueString(), tokenFilePath)
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
		func() resource.Resource { return NewLakehouseResource(p.client) },
		func() resource.Resource { return NewSparkPoolResource(p.client) },
		func() resource.Resource { return NewDomainWorkspaceAssignResource(p.client) },
		func() resource.Resource { return NewShortcutResource(p.client) },
		func() resource.Resource { return NewLakehouseTableResource(p.client) },
		func() resource.Resource { return NewKqlDatabaseResource(p.client) },
	}
}
