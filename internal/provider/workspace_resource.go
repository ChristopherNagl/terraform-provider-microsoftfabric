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
type workspaceResource struct {
    client *apiclient.APIClient
}

// Define the schema.
func (r *workspaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed: true,
            },
            "name": schema.StringAttribute{
                Required: true,
				Description: "The workspace display name. The display name cannot contain more than 256 characters. Only unused workspace names are allowed. 'Admin monitoring' is a reserved workspace name.",
            },
            "description": schema.StringAttribute{
                Optional: true,
				Description: "The workspace description. The description cannot contain more than 4000 characters.",
            },
            "last_updated": schema.StringAttribute{
                Computed: true,
            },
        },
    }
}

// Define the model.
type workspaceResourceModel struct {
    ID          types.String `tfsdk:"id"`
    Name        types.String `tfsdk:"name"`
    Description types.String `tfsdk:"description"`
    LastUpdated types.String `tfsdk:"last_updated"`
}

// Implement Metadata method.
func (r *workspaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = "microsoftfabric_workspace"
}

// Define the provider.
func NewWorkspaceResource(client *apiclient.APIClient) resource.Resource {
    return &workspaceResource{client: client}
}

// Implement CRUD operations.
func (r *workspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // Retrieve values from plan.
    var plan workspaceResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Create workspace.
    workspaceID, err := r.createWorkspace(plan.Name.ValueString(), plan.Description.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error creating workspace",
            "Could not create workspace: "+err.Error(),
        )
        return
    }

    // Set ID and LastUpdated fields.
    plan.ID = types.StringValue(workspaceID)
    plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

    // Set state.
    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *workspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state workspaceResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Read the workspace using the ID from the state
    workspace, err := r.readWorkspace(state.ID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error reading workspace",
            "Could not read workspace: "+err.Error(),
        )
        return
    }

    // Ensure the workspace exists
    if workspace == nil {
        resp.Diagnostics.AddError(
            "Workspace not found",
            "The workspace with ID "+state.ID.ValueString()+" does not exist.",
        )
        resp.State.RemoveResource(ctx)
        return
    }

    // Update the state fields from the API response
    if name, ok := workspace["displayName"].(string); ok {
        state.Name = types.StringValue(name)
    } else {
        resp.Diagnostics.AddError("Error reading workspace", "Display name not found in API response")
        return
    }

    // Handle the description
    if description, ok := workspace["description"].(string); ok && description != "" {
        state.Description = types.StringValue(description)
    } else {
        state.Description = types.StringNull() // Indicate that it is not set
    }

    // Update the state in response
    diags = resp.State.Set(ctx, state)
    resp.Diagnostics.Append(diags...)

}

func (r *workspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // Retrieve values from plan.
    var plan workspaceResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Retrieve ID from state.
    var state workspaceResourceModel
    diags = req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Update workspace.
    err := r.updateWorkspace(state.ID.ValueString(), plan.Name.ValueString(), plan.Description.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error updating workspace",
            "Could not update workspace: "+err.Error(),
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

func (r *workspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    // Retrieve ID from state.
    var state workspaceResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Delete workspace.
    err := r.deleteWorkspace(state.ID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error deleting workspace",
            "Could not delete workspace: "+err.Error(),
        )
        return
    }

    // Remove resource from state.
    resp.State.RemoveResource(ctx)
}

// Implement workspace creation function.
func (r *workspaceResource) createWorkspace(name string, description string) (string, error) {
    url := "https://api.fabric.microsoft.com/v1/workspaces"
    body := map[string]string{
        "displayName": name,
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

// Implement workspace read function.
func (r *workspaceResource) readWorkspace(id string) (map[string]interface{}, error) {
    url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s", id)

    respBody, err := r.client.Get(url)
    if err != nil {
        return nil, err
    }

    return respBody, nil
}

// Implement workspace update function.
func (r *workspaceResource) updateWorkspace(id, name, description string) error {
    url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s", id)
    body := map[string]string{
        "displayName": name,
        "description": description,
    }

    _, err := r.client.Patch(url, body)
    if err != nil {
        return err
    }

    // Since the update operation doesn't return a JSON body, we don't need to handle any response body here.
    return nil
}

// Implement workspace deletion function.
func (r *workspaceResource) deleteWorkspace(id string) error {
    url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s", id)

    err := r.client.Delete(url)
    if err != nil {
        return err
    }

    // Since the delete operation doesn't return a JSON body, we don't need to handle any response body here.
    return nil
}