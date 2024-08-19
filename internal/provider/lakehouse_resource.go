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
type lakehouseResource struct {
    client *apiclient.APIClient
}

// Define the schema.
func (r *lakehouseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed: true,
                Description: "The unique identifier for the lakehouse resource.",
            },
            "workspace_id": schema.StringAttribute{
                Required: true,
                Description: "The ID of the workspace where the lakehouse will be created.",
            },
            "display_name": schema.StringAttribute{
                Required: true,
                Description: "The display name for the lakehouse resource. This name is used in the user interface.",
            },
            "description": schema.StringAttribute{
                Optional: true,
                Description: "An optional description of the lakehouse resource, providing more context about its purpose.",
            },
            "last_updated": schema.StringAttribute{
                Computed: true,
                Description: "The timestamp of the last update made to the lakehouse resource.",
            },
        },
    }
}

// Define the model.
type lakehouseResourceModel struct {
    ID          types.String `tfsdk:"id"`
    WorkspaceID types.String `tfsdk:"workspace_id"`
    DisplayName types.String `tfsdk:"display_name"`
    Description types.String `tfsdk:"description"`
    LastUpdated types.String `tfsdk:"last_updated"`
}

// Implement Metadata method.
func (r *lakehouseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = "microsoftfabric_lakehouse"
}

// Define the provider.
func NewLakehouseResource(client *apiclient.APIClient) resource.Resource {
    return &lakehouseResource{client: client}
}

// Implement CRUD operations.
func (r *lakehouseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // Retrieve values from plan.
    var plan lakehouseResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Create lakehouse.
    lakehouseID, err := r.createLakehouse(plan.WorkspaceID.ValueString(), plan.DisplayName.ValueString(), plan.Description.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error creating lakehouse",
            "Could not create lakehouse: "+err.Error(),
        )
        return
    }

    // Set ID and LastUpdated fields.
    plan.ID = types.StringValue(lakehouseID)
    plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

    // Set state.
    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *lakehouseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    // Retrieve ID from state.
    var state lakehouseResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Read lakehouse.
    lakehouse, err := r.readLakehouse(state.WorkspaceID.ValueString(), state.ID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error reading lakehouse",
            "Could not read lakehouse: "+err.Error(),
        )
        return
    }

    // Check for required keys.
    displayName, ok := lakehouse["displayName"].(string)
    if !ok {
        resp.Diagnostics.AddError(
            "Error reading lakehouse",
            "Unexpected response format: 'displayName' key not found or not a string",
        )
        return
    }

    // Set state.
    state.DisplayName = types.StringValue(displayName)
    state.Description = types.StringValue(lakehouse["description"].(string)) // Assuming description is included in response
    state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

    diags = resp.State.Set(ctx, state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (r *lakehouseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // Retrieve values from plan.
    var plan lakehouseResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Retrieve ID from state.
    var state lakehouseResourceModel
    diags = req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Update lakehouse.
    err := r.updateLakehouse(state.WorkspaceID.ValueString(), state.ID.ValueString(), plan.DisplayName.ValueString(), plan.Description.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error updating lakehouse",
            "Could not update lakehouse: "+err.Error(),
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

func (r *lakehouseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    // Retrieve ID from state.
    var state lakehouseResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Delete lakehouse (implement delete if needed; API details not provided).
    err := r.deleteLakehouse(state.WorkspaceID.ValueString(), state.ID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error deleting lakehouse",
            "Could not delete lakehouse: "+err.Error(),
        )
        return
    }

    // Remove resource from state.
    resp.State.RemoveResource(ctx)
}

// Helper functions for lakehouse operations.
func (r *lakehouseResource) createLakehouse(workspaceID, displayName, description string) (string, error) {
    url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/lakehouses", workspaceID)
    body := map[string]string{
        "displayName": displayName,
        "description": description,
    }

    // Send the POST request using the existing Post method from the API client.
    responseBody, err := r.client.Post(url, body)
    if err != nil {
        return "", fmt.Errorf("error during POST request: %w", err)
    }

    // Log the full response body for debugging.
    fmt.Printf("Full Response Body: %+v\n", responseBody)

    // Extract the lakehouse ID from the response
    lakehouseID, ok := responseBody["id"].(string)
    if !ok {
        return "", fmt.Errorf("unexpected response format: 'id' key not found")
    }

    return lakehouseID, nil
}



func (r *lakehouseResource) readLakehouse(workspaceID, lakehouseID string) (map[string]interface{}, error) {
    url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/lakehouses/%s", workspaceID, lakehouseID)
    return r.client.Get(url)
}

func (r *lakehouseResource) updateLakehouse(workspaceID, lakehouseID, displayName, description string) error {
    url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/lakehouses/%s", workspaceID, lakehouseID)
    body := map[string]string{
        "displayName": displayName,
        "description": description,
    }

    _, err := r.client.Patch(url, body)
    return err
}

func (r *lakehouseResource) deleteLakehouse(workspaceID, lakehouseID string) error {
    url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/lakehouses/%s", workspaceID, lakehouseID)
    err := r.client.Delete(url)
    if err != nil {
        return err
    }

    // No response body to handle for delete.
    return nil
}