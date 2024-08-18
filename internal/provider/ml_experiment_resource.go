package provider

import (
	"context"
	"fmt"
	"terraform-provider-microsoftfabric/internal/apiclient"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Define the resource.
type mlExperimentResource struct {
	client *apiclient.APIClient
}

// Define the schema.
func (r *mlExperimentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"workspace_id": schema.StringAttribute{
				Required: true,
			},
			"display_name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Define the model.
type mlExperimentResourceModel struct {
	ID          types.String `tfsdk:"id"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
	DisplayName types.String `tfsdk:"display_name"`
	Description types.String `tfsdk:"description"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

// Implement Metadata method.
func (r *mlExperimentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "microsoftfabric_ml_experiment"
}

// Define the provider.
func NewMLEexperimentResource(client *apiclient.APIClient) resource.Resource {
	return &mlExperimentResource{client: client}
}

// Implement CRUD operations.
func (r *mlExperimentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mlExperimentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create ML experiment.
	experimentID, err := r.createMLEExperiment(plan.WorkspaceID.ValueString(), plan.DisplayName.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating ML experiment",
			"Could not create ML experiment: "+err.Error(),
		)
		return
	}

	// Set ID and LastUpdated fields.
	plan.ID = types.StringValue(experimentID)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *mlExperimentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state mlExperimentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read ML experiment.
	experiment, err := r.readMLEExperiment(state.WorkspaceID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading ML experiment",
			"Could not read ML experiment: "+err.Error(),
		)
		return
	}

	// Set state.
	state.DisplayName = types.StringValue(experiment.DisplayName.ValueString())
	state.Description = types.StringValue(experiment.Description.ValueString())
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *mlExperimentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan mlExperimentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state mlExperimentResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update ML experiment.
	err := r.updateMLEExperiment(state.WorkspaceID.ValueString(), state.ID.ValueString(), plan.DisplayName.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating ML experiment",
			"Could not update ML experiment: "+err.Error(),
		)
		return
	}

	// Set LastUpdated field.
	plan.ID = state.ID // Ensure the ID remains unchanged.
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *mlExperimentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state mlExperimentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete ML experiment.
	err := r.deleteMLEExperiment(state.WorkspaceID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting ML experiment",
			"Could not delete ML experiment: "+err.Error(),
		)
		return
	}

	// Remove resource from state.
	resp.State.RemoveResource(ctx)
}

// Helper functions for ML experiment operations.

func (r *mlExperimentResource) createMLEExperiment(workspaceID, displayName, description string) (string, error) {
	url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/mlExperiments", workspaceID)
	body := map[string]string{
		"displayName": displayName,
		"description": description,
	}

	responseBody, err := r.client.PostWithOperationCheck(url, body)
	if err != nil {
		return "", err
	}

	experimentID, ok := responseBody["id"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected response format: 'id' key not found")
	}

	return experimentID, nil
}

func (r *mlExperimentResource) readMLEExperiment(workspaceID, experimentID string) (mlExperimentResourceModel, error) {
	url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/mlExperiments/%s", workspaceID, experimentID)

	responseBody, err := r.client.Get(url)
	if err != nil {
		return mlExperimentResourceModel{}, err
	}

	// Assuming the response body contains fields we need.
	experiment := mlExperimentResourceModel{
		ID:          types.StringValue(responseBody["id"].(string)),
		DisplayName: types.StringValue(responseBody["displayName"].(string)),
		Description: types.StringValue(responseBody["description"].(string)), // Assuming this exists.
	}

	return experiment, nil
}

func (r *mlExperimentResource) updateMLEExperiment(workspaceID, experimentID, displayName, description string) error {
	url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/mlExperiments/%s", workspaceID, experimentID)
	body := map[string]string{
		"displayName": displayName,
		"description": description,
	}

	_, err := r.client.Patch(url, body)
	return err
}

func (r *mlExperimentResource) deleteMLEExperiment(workspaceID, experimentID string) error {
	url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/mlExperiments/%s", workspaceID, experimentID)
	return r.client.Delete(url)
}
