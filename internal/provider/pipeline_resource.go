package provider

import (
	"context"
	"fmt"
	"time"

	"terraform-provider-microsoftfabric/internal/apiclient"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Define the resource.
type pipelineResource struct {
	client *apiclient.APIClient
}

// Define the schema.
func (r *pipelineResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"display_name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"workspaces": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"workspace_id": schema.StringAttribute{
							Required: true,
						},
						"stage_order": schema.Int64Attribute{
							Required: true,
						},
					},
				},
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type pipelineWorkspaceModel struct {
	WorkspaceID types.String `tfsdk:"workspace_id"`
	StageOrder  types.Int64  `tfsdk:"stage_order"`
}

type pipelineResourceModel struct {
	ID          types.String             `tfsdk:"id"`
	DisplayName types.String             `tfsdk:"display_name"`
	Description types.String             `tfsdk:"description"`
	Workspaces  []pipelineWorkspaceModel `tfsdk:"workspaces"`
	LastUpdated types.String             `tfsdk:"last_updated"`
}

// Implement Metadata method.
func (r *pipelineResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "microsoftfabric_pipeline"
}

// Define the provider.
func NewPipelineResource(client *apiclient.APIClient) resource.Resource {
	return &pipelineResource{client: client}
}

// Implement CRUD operations.
func (r *pipelineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pipelineResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create pipeline.
	pipelineID, err := r.createPipeline(plan.DisplayName.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating pipeline",
			"Could not create pipeline: "+err.Error(),
		)
		return
	}

	// Assign workspaces if provided.
	for _, workspace := range plan.Workspaces {
		err = r.assignWorkspace(pipelineID, int(workspace.StageOrder.ValueInt64()), workspace.WorkspaceID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error assigning workspace",
				"Could not assign workspace: "+err.Error(),
			)
			return
		}
	}

	// Set ID and LastUpdated fields.
	plan.ID = types.StringValue(pipelineID)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *pipelineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state pipelineResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read pipeline.
	pipeline, err := r.readPipeline(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading pipeline",
			"Could not read pipeline: "+err.Error(),
		)
		return
	}

	displayName, ok := pipeline["displayName"].(string)
	if !ok {
		resp.Diagnostics.AddError(
			"Error reading pipeline",
			"Unexpected response format: 'displayName' key not found or not a string",
		)
		return
	}

	description, _ := pipeline["description"].(string) // description is optional

	// Set state
	state.DisplayName = types.StringValue(displayName)
	state.Description = types.StringValue(description)
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
func (r *pipelineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan pipelineResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state pipelineResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update pipeline.
	err := r.updatePipeline(state.ID.ValueString(), plan.DisplayName.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating pipeline",
			"Could not update pipeline: "+err.Error(),
		)
		return
	}

	// Prepare to manage workspace assignments.
	currentWorkspaceAssignments := make(map[string]int) // Track current workspace assignments.
	for _, workspace := range state.Workspaces {
		currentWorkspaceAssignments[workspace.WorkspaceID.ValueString()] = int(workspace.StageOrder.ValueInt64())
	}

	for _, workspace := range plan.Workspaces {
		workspaceID := workspace.WorkspaceID.ValueString()
		newStageOrder := int(workspace.StageOrder.ValueInt64())

		// Check the current stage order for this workspace.
		if currentStageOrder, exists := currentWorkspaceAssignments[workspaceID]; exists {
			if currentStageOrder != newStageOrder {
				// If the stage order has changed, unassign from the current stage.
				err = r.unassignWorkspace(state.ID.ValueString(), currentStageOrder, workspaceID)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error unassigning workspace",
						"Could not unassign workspace: "+err.Error(),
					)
					return
				}
			}

			// Remove from the current assignments map to avoid reassigning it.
			delete(currentWorkspaceAssignments, workspaceID)
		}

		// Assign the workspace to the new stage.
		err = r.assignWorkspace(state.ID.ValueString(), newStageOrder, workspaceID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error assigning workspace",
				"Could not assign workspace: "+err.Error(),
			)
			return
		}
	}

	// Unassign any remaining workspaces in current assignments that are not in the plan.
	for workspaceID, stageOrder := range currentWorkspaceAssignments {
		err = r.unassignWorkspace(state.ID.ValueString(), stageOrder, workspaceID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error unassigning workspace",
				"Could not unassign workspace: "+err.Error(),
			)
			return
		}
	}

	// Set LastUpdated field.
	plan.ID = state.ID // Ensure the ID remains unchanged.
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *pipelineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state pipelineResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Unassign all workspaces before deletion.
	for _, workspace := range state.Workspaces {
		err := r.unassignWorkspace(state.ID.ValueString(), int(workspace.StageOrder.ValueInt64()), workspace.WorkspaceID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error unassigning workspace during delete",
				"Could not unassign workspace: "+err.Error(),
			)
			return
		}
	}

	// Delete pipeline.
	err := r.deletePipeline(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting pipeline",
			"Could not delete pipeline: "+err.Error(),
		)
		return
	}

	// Remove resource from state.
	resp.State.RemoveResource(ctx)
}

// Implement pipeline creation function.
func (r *pipelineResource) createPipeline(displayName, description string) (string, error) {
	url := "https://api.powerbi.com/v1.0/myorg/pipelines"
	body := map[string]string{
		"displayName": displayName,
		"description": description,
	}

	respBody, err := r.client.Post(url, body)
	if err != nil {
		return "", err
	}

	if id, ok := respBody["id"].(string); ok {
		return id, nil
	}

	return "", fmt.Errorf("unexpected response: %v", respBody)
}

// Implement pipeline read function.
func (r *pipelineResource) readPipeline(id string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.powerbi.com/v1.0/myorg/pipelines/%s", id)

	respBody, err := r.client.Get(url)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

// Implement pipeline update function.
func (r *pipelineResource) updatePipeline(id, displayName, description string) error {
	url := fmt.Sprintf("https://api.powerbi.com/v1.0/myorg/pipelines/%s", id)
	body := map[string]string{
		"displayName": displayName,
		"description": description,
	}

	_, err := r.client.Patch(url, body)
	if err != nil {
		return err
	}

	// No response body is expected for a successful update.
	return nil
}

// Implement pipeline deletion function.
func (r *pipelineResource) deletePipeline(id string) error {
	url := fmt.Sprintf("https://api.powerbi.com/v1.0/myorg/pipelines/%s", id)

	err := r.client.Delete(url)
	if err != nil {
		return err
	}

	// No response body is expected for a successful delete.
	return nil
}

// Implement workspace assignment function.
func (r *pipelineResource) assignWorkspace(pipelineID string, stageOrder int, workspaceID string) error {
	url := fmt.Sprintf("https://api.powerbi.com/v1.0/myorg/pipelines/%s/stages/%d/assignWorkspace", pipelineID, stageOrder)
	body := map[string]string{
		"workspaceId": workspaceID,
	}

	respBody, err := r.client.Post(url, body)
	if err != nil {
		return err
	}

	if len(respBody) > 0 {
		return nil
	}

	return nil
}

// Implement workspace unassignment function.
func (r *pipelineResource) unassignWorkspace(pipelineID string, stageOrder int, workspaceID string) error {
	url := fmt.Sprintf("https://api.powerbi.com/v1.0/myorg/pipelines/%s/stages/%d/unassignWorkspace", pipelineID, stageOrder)
	body := map[string]string{
		"workspaceId": workspaceID,
	}

	respBody, err := r.client.Post(url, body)
	if err != nil {
		return err
	}

	if len(respBody) > 0 {
		return nil
	}

	return nil
}
