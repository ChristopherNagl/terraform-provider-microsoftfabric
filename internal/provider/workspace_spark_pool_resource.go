package provider

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    "terraform-provider-microsoftfabric/internal/apiclient"

    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

// Define the Spark pool resource.
type sparkPoolResource struct {
    client *apiclient.APIClient
}

// Define the schema for the Spark pool resource.
func (r *sparkPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed:    true,
                Description: "Custom pool ID.",
            },
            "workspace_id": schema.StringAttribute{
                Required:    true,
                Description: "ID of the workspace to which the custom pool belongs.",
            },
            "name": schema.StringAttribute{
                Required:    true,
                Description: "Custom pool name. The name must be between 1 and 64 characters long and must contain only letters, numbers, dashes, underscores, and spaces. Custom pool names must be unique within the workspace. 'Starter Pool' is a reserved custom pool name.",
            },
            "node_family": schema.StringAttribute{
                Required:    true,
                Description: "Node family. Available options include 'MemoryOptimized'. Additional NodeFamily types may be added over time.",
            },
            "node_size": schema.StringAttribute{
                Required:    true,
                Description: "Node size. Available options include 'Small', 'Medium', 'Large', 'XLarge', 'XXLarge'. Additional NodeSize types may be added over time.",
            },
            "auto_scale": schema.SingleNestedAttribute{
                Required:    true,
                Description: "Autoscale properties.",
                Attributes: map[string]schema.Attribute{
                    "enabled": schema.BoolAttribute{
                        Required:    true,
                        Description: "The status of the autoscale. False - Disabled, true - Enabled.",
                    },
                    "min_node_count": schema.Int64Attribute{
                        Required:    true,
                        Description: "The minimum node count.",
                    },
                    "max_node_count": schema.Int64Attribute{
                        Required:    true,
                        Description: "The maximum node count.",
                    },
                },
            },
            "dynamic_executor_allocation": schema.SingleNestedAttribute{
                Required:    true,
                Description: "Dynamic executor allocation properties.",
                Attributes: map[string]schema.Attribute{
                    "enabled": schema.BoolAttribute{
                        Required:    true,
                        Description: "The status of the dynamic executor allocation. False - Disabled, true - Enabled.",
                    },
                    "min_executors": schema.Int64Attribute{
                        Required:    true,
                        Description: "The minimum number of executors.",
                    },
                    "max_executors": schema.Int64Attribute{
                        Required:    true,
                        Description: "The maximum number of executors.",
                    },
                },
            },
            "last_updated": schema.StringAttribute{
                Computed:    true,
                Description: "The timestamp of the last update.",
            },
        },
    }
}


// Define the model for the Spark pool resource.
type sparkPoolResourceModel struct {
    ID                        types.String                `tfsdk:"id"`
    WorkspaceID               types.String                `tfsdk:"workspace_id"`
    Name                      types.String                `tfsdk:"name"`
    NodeFamily                types.String                `tfsdk:"node_family"`
    NodeSize                  types.String                `tfsdk:"node_size"`
    AutoScale                 AutoScalePropertiesModel    `tfsdk:"auto_scale"`
    DynamicExecutorAllocation  DynamicExecutorAllocationModel `tfsdk:"dynamic_executor_allocation"`
    LastUpdated               types.String                `tfsdk:"last_updated"`
}

type AutoScalePropertiesModel struct {
    Enabled     types.Bool  `tfsdk:"enabled"`
    MinNodeCount types.Int64 `tfsdk:"min_node_count"`
    MaxNodeCount types.Int64 `tfsdk:"max_node_count"`
}

type DynamicExecutorAllocationModel struct {
    Enabled     types.Bool  `tfsdk:"enabled"`
    MinExecutors types.Int64 `tfsdk:"min_executors"`
    MaxExecutors types.Int64 `tfsdk:"max_executors"`
}

// Implement Metadata method.
func (r *sparkPoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = "microsoftfabric_spark_pool"
}

// Define the provider.
func NewSparkPoolResource(client *apiclient.APIClient) resource.Resource {
    return &sparkPoolResource{client: client}
}

// Implement CRUD operations.
func (r *sparkPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan sparkPoolResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    poolID, err := r.createSparkPool(plan.WorkspaceID.ValueString(), plan)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error creating Spark pool",
            "Could not create Spark pool: "+err.Error(),
        )
        return
    }

    plan.ID = types.StringValue(poolID)
    plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
}

func (r *sparkPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state sparkPoolResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Implement logic to read the current state if necessary.
}

func (r *sparkPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan sparkPoolResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    var state sparkPoolResourceModel
    diags = req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Update the Spark pool configuration
    err := r.updateSparkPool(plan.WorkspaceID.ValueString(), state.ID.ValueString(), plan)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error updating Spark pool",
            "Could not update Spark pool: "+err.Error(),
        )
        return
    }

    // Preserve the ID from the state
    plan.ID = state.ID // Preserve the existing ID
    plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
}
func (r *sparkPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state sparkPoolResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    err := r.deleteSparkPool(state.WorkspaceID.ValueString(), state.ID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error deleting Spark pool",
            "Could not delete Spark pool: "+err.Error(),
        )
        return
    }

    resp.State.RemoveResource(ctx)
}

// Helper function to create the Spark pool.
func (r *sparkPoolResource) createSparkPool(workspaceID string, plan sparkPoolResourceModel) (string, error) {
    url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/spark/pools", workspaceID)

    body := map[string]interface{}{
        "name":        plan.Name.ValueString(),
        "nodeFamily":  plan.NodeFamily.ValueString(),
        "nodeSize":    plan.NodeSize.ValueString(),
        "autoScale": map[string]interface{}{
            "enabled":      plan.AutoScale.Enabled.ValueBool(),
            "minNodeCount": plan.AutoScale.MinNodeCount.ValueInt64(),
            "maxNodeCount": plan.AutoScale.MaxNodeCount.ValueInt64(),
        },
        "dynamicExecutorAllocation": map[string]interface{}{
            "enabled":     plan.DynamicExecutorAllocation.Enabled.ValueBool(),
            "minExecutors": plan.DynamicExecutorAllocation.MinExecutors.ValueInt64(),
            "maxExecutors": plan.DynamicExecutorAllocation.MaxExecutors.ValueInt64(),
        },
    }

    bodyBytes, err := json.Marshal(body)
    if err != nil {
        return "", fmt.Errorf("failed to marshal request body: %v", err)
    }

    respBody, err := r.client.PostBytes(url, bodyBytes)
    if err != nil {
        return "", fmt.Errorf("failed to create Spark pool: %v", err)
    }

    // Read the created pool ID from the response.
    poolID, exists := respBody["id"].(string)
    if !exists || poolID == "" {
        return "", fmt.Errorf("expected field 'id' not found in response or is empty")
    }

    return poolID, nil
}

// Helper function to update the Spark pool.
func (r *sparkPoolResource) updateSparkPool(workspaceID, poolID string, plan sparkPoolResourceModel) error {
    url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/spark/pools/%s", workspaceID, poolID)

    body := map[string]interface{}{
        "name":        plan.Name.ValueString(),
        "nodeFamily":  plan.NodeFamily.ValueString(),
        "nodeSize":    plan.NodeSize.ValueString(),
        "autoScale": map[string]interface{}{
            "enabled":      plan.AutoScale.Enabled.ValueBool(),
            "minNodeCount": plan.AutoScale.MinNodeCount.ValueInt64(),
            "maxNodeCount": plan.AutoScale.MaxNodeCount.ValueInt64(),
        },
        "dynamicExecutorAllocation": map[string]interface{}{
            "enabled":     plan.DynamicExecutorAllocation.Enabled.ValueBool(),
            "minExecutors": plan.DynamicExecutorAllocation.MinExecutors.ValueInt64(),
            "maxExecutors": plan.DynamicExecutorAllocation.MaxExecutors.ValueInt64(),
        },
    }

    _, err := r.client.PatchBytes(url, body) // Use the new Patch method
    if err != nil {
        return fmt.Errorf("failed to update Spark pool: %v", err)
    }

    // Optionally handle the response as needed

    return nil
}

// Helper function to delete the Spark pool.
func (r *sparkPoolResource) deleteSparkPool(workspaceID, poolID string) error {
    url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/spark/pools/%s", workspaceID, poolID)

    err := r.client.Delete(url) // Assuming your APIClient has a Delete method
    if err != nil {
        return fmt.Errorf("failed to delete Spark pool: %v", err)
    }

    return nil
}