package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-microsoftfabric/internal/apiclient"
	"time"
)

// Define the combined resource.
type workspaceGitResource struct {
	client *apiclient.APIClient
}

// Define the schema for the combined resource.
func (r *workspaceGitResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"workspace_id": schema.StringAttribute{
				Required: true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"git_provider_details": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"organization_name": schema.StringAttribute{
						Required: true,
					},
					"project_name": schema.StringAttribute{
						Required: true,
					},
					"git_provider_type": schema.StringAttribute{
						Required: true,
					},
					"repository_name": schema.StringAttribute{
						Required: true,
					},
					"branch_name": schema.StringAttribute{
						Required: true,
					},
					"directory_name": schema.StringAttribute{
						Required: true,
					},
				},
			},
			"initialization_strategy": schema.StringAttribute{
				Required: true,
			},
			"remote_commit_hash": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Define the model for the combined resource.
type workspaceGitResourceModel struct {
	ID                     types.String            `tfsdk:"id"`
	WorkspaceID            types.String            `tfsdk:"workspace_id"`
	GitProviderDetails     GitProviderDetailsModel `tfsdk:"git_provider_details"`
	LastUpdated            types.String            `tfsdk:"last_updated"`
	InitializationStrategy types.String            `tfsdk:"initialization_strategy"`
	RemoteCommitHash       types.String            `tfsdk:"remote_commit_hash"`
}

type GitProviderDetailsModel struct {
	OrganizationName types.String `tfsdk:"organization_name"`
	ProjectName      types.String `tfsdk:"project_name"`
	GitProviderType  types.String `tfsdk:"git_provider_type"`
	RepositoryName   types.String `tfsdk:"repository_name"`
	BranchName       types.String `tfsdk:"branch_name"`
	DirectoryName    types.String `tfsdk:"directory_name"`
}

// Implement Metadata method.
func (r *workspaceGitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "microsoftfabric_workspace_git"
}

// Define the provider.
func NewWorkspaceGitResource(client *apiclient.APIClient) resource.Resource {
	return &workspaceGitResource{client: client}
}

// Implement CRUD operations.
func (r *workspaceGitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan workspaceGitResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Connect workspace to Git.
	err := r.connectWorkspaceToGit(plan.WorkspaceID.ValueString(), plan.GitProviderDetails)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error connecting workspace to Git",
			"Could not connect workspace to Git: "+err.Error(),
		)
		return
	}

	// Create Git connection.
	remoteCommitHash, err := r.createGitInit(plan.WorkspaceID.ValueString(), plan.InitializationStrategy.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating git connection",
			"Could not create git connection: "+err.Error(),
		)
		return
	}

	// Commit from Git.
	err = r.commitFromGit(remoteCommitHash, plan.WorkspaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error committing from Git",
			"Could not commit from Git: "+err.Error(),
		)
		return
	}

	// Set fields.
	plan.ID = types.StringValue(plan.WorkspaceID.ValueString() + "-git")
	plan.RemoteCommitHash = types.StringValue(remoteCommitHash)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *workspaceGitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve ID from state.
	var state workspaceGitResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Implement the logic to read the current state if necessary.
}

func (r *workspaceGitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan workspaceGitResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve ID from state.
	var state workspaceGitResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Step 1: Delete the existing Git connection.
	err := r.deleteGitConnection(state.WorkspaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Git connection",
			"Could not delete Git connection: "+err.Error(),
		)
		return
	}

	// Step 2: Reconnect to Git.
	err = r.connectWorkspaceToGit(plan.WorkspaceID.ValueString(), plan.GitProviderDetails)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error connecting workspace to Git",
			"Could not connect workspace to Git: "+err.Error(),
		)
		return
	}

	// Step 3: Create Git initialization.
	remoteCommitHash, err := r.createGitInit(plan.WorkspaceID.ValueString(), plan.InitializationStrategy.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating git connection",
			"Could not create git connection: "+err.Error(),
		)
		return
	}

	// Step 4: Commit from Git if there's a valid remote commit hash.
	if plan.RemoteCommitHash.ValueString() != "" {
		err = r.commitFromGit(plan.RemoteCommitHash.ValueString(), state.WorkspaceID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error committing from Git",
				"Could not commit from Git: "+err.Error(),
			)
			return
		}
	}

	// Step 5: Set LastUpdated field and other state attributes.
	plan.ID = state.ID // Ensure the ID remains unchanged.
	plan.RemoteCommitHash = types.StringValue(remoteCommitHash)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *workspaceGitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve ID from state.
	var state workspaceGitResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete Git connection.
	err := r.deleteGitConnection(state.WorkspaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Git connection",
			"Could not delete Git connection: "+err.Error(),
		)
		return
	}

	// Remove resource from state.
	resp.State.RemoveResource(ctx)
}

// Helper function to connect workspace to Git.
func (r *workspaceGitResource) connectWorkspaceToGit(workspaceID string, details GitProviderDetailsModel) error {
	url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/git/connect", workspaceID)

	// Prepare the request body.
	body := map[string]interface{}{
		"gitProviderDetails": map[string]interface{}{
			"organizationName": details.OrganizationName.ValueString(),
			"projectName":      details.ProjectName.ValueString(),
			"gitProviderType":  details.GitProviderType.ValueString(),
			"repositoryName":   details.RepositoryName.ValueString(),
			"branchName":       details.BranchName.ValueString(),
			"directoryName":    details.DirectoryName.ValueString(),
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	_, err = r.client.PostBytes(url, bodyBytes)
	if err != nil {
		return fmt.Errorf("failed to connect workspace to Git: %v", err)
	}

	return nil
}

// Helper function to create Git initialization.
func (r *workspaceGitResource) createGitInit(workspaceID, initializationStrategy string) (string, error) {
	url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/git/initializeConnection", workspaceID)

	requestBody := map[string]string{
		"initializationStrategy": initializationStrategy,
	}

	respBody, err := r.client.Post(url, requestBody)
	if err != nil {
		return "", err
	}

	// Log the API response for debugging.
	fmt.Printf("API Response: %+v\n", respBody)

	// Check for the expected field in the response.
	remoteCommitHash, exists := respBody["remoteCommitHash"].(string)
	if !exists || remoteCommitHash == "" {
		return "", fmt.Errorf("expected field 'remoteCommitHash' not found in response or is empty")
	}

	return remoteCommitHash, nil
}

// Helper function for updating the Git connection.
func (r *workspaceGitResource) updateGitConnection(workspaceID string, details GitProviderDetailsModel) error {
	// Implement logic to update the current Git connection if needed.
	return r.connectWorkspaceToGit(workspaceID, details) // Reconnect for simplicity.
}

// Helper function to delete Git connection.
func (r *workspaceGitResource) deleteGitConnection(workspaceID string) error {
	url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/git/disconnect", workspaceID)

	r.client.Post(url, nil)
	// Since the delete operation doesn't return a JSON body, we don't need to handle any response body here.
	return nil
}

// Helper function to commit from Git.
func (r *workspaceGitResource) commitFromGit(remoteCommitHash string, workspaceID string) error {
	url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/git/updateFromGit", workspaceID)
	body := map[string]string{
		"remoteCommitHash": remoteCommitHash,
	}

	_, err := r.client.Post(url, body)
	if err != nil {
		return fmt.Errorf("failed to commit from Git: %v", err)
	}

	return nil
}
