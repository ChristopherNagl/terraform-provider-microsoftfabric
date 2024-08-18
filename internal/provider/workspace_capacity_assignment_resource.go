package provider

import (
	"context"
	"fmt"
	"terraform-provider-microsoftfabric/internal/apiclient"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Define the resource.
type workspaceCapacityAssignmentResource struct {
    client *apiclient.APIClient
}

// Define the schema.
func (r *workspaceCapacityAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "workspace_id": schema.StringAttribute{
                Required: true,
            },
            "capacity_id": schema.StringAttribute{
                Required: true,
            },
        },
    }
}

// Define the model.
type workspaceCapacityAssignmentResourceModel struct {
    WorkspaceID types.String `tfsdk:"workspace_id"`
    CapacityID  types.String `tfsdk:"capacity_id"`
}

// Implement Metadata method.
func (r *workspaceCapacityAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = "microsoftfabric_workspace_capacity_assignment"
}

// Define the provider.
func NewWorkspaceCapacityAssignmentResource(client *apiclient.APIClient) resource.Resource {
    return &workspaceCapacityAssignmentResource{client: client}
}

// Implement CRUD operations.
func (r *workspaceCapacityAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // Retrieve values from plan.
    var plan workspaceCapacityAssignmentResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Assign capacity to workspace.
    err := r.assignCapacityToWorkspace(plan.WorkspaceID.ValueString(), plan.CapacityID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error assigning capacity to workspace",
            "Could not assign capacity to workspace: "+err.Error(),
        )
        return
    }

    // Set state.
    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *workspaceCapacityAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    // Implement read operation if needed.
}

func (r *workspaceCapacityAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // Retrieve current state and updated plan values.
    var state, plan workspaceCapacityAssignmentResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    diags = req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Unassign capacity from current workspace.
    err := r.unassignCapacityFromWorkspace(state.WorkspaceID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error unassigning capacity from workspace",
            "Could not unassign capacity from workspace: "+err.Error(),
        )
        return
    }

    // Assign new capacity to workspace.
    err = r.assignCapacityToWorkspace(plan.WorkspaceID.ValueString(), plan.CapacityID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error assigning new capacity to workspace",
            "Could not assign new capacity to workspace: "+err.Error(),
        )
        return
    }

    // Update state to reflect new capacity assignment.
    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *workspaceCapacityAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    // Retrieve values from state.
    var state workspaceCapacityAssignmentResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Unassign capacity from workspace.
    err := r.unassignCapacityFromWorkspace(state.WorkspaceID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error unassigning capacity from workspace",
            "Could not unassign capacity from workspace: "+err.Error(),
        )
        return
    }

    // Remove the state since the resource is deleted.
    resp.State.RemoveResource(ctx)
}

// Implement the function to assign capacity to workspace using the Fabric API.
func (r *workspaceCapacityAssignmentResource) assignCapacityToWorkspace(workspaceID, capacityID string) error {
    url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/assignToCapacity", workspaceID)
    body := map[string]string{"capacityId": capacityID}

    // Sending the POST request to the Fabric API
    _, err := r.client.Post(url, body)
    if err != nil {
        return err
    }

    return nil
}

// Implement the function to unassign capacity from workspace.
func (r *workspaceCapacityAssignmentResource) unassignCapacityFromWorkspace(workspaceID string) error {
    url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/unassignFromCapacity", workspaceID)

    _, err := r.client.Post(url, nil)
    if err != nil {
        return err
    }

    return nil
}