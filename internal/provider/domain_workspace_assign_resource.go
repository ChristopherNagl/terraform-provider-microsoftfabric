package provider

import (
    "context"
    "encoding/json"
    "fmt"
    "terraform-provider-microsoftfabric/internal/apiclient"

    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

// Define the domain_workspace_assign resource.
type domainWorkspaceAssignResource struct {
    client *apiclient.APIClient
}

// Define the schema for the domain_workspace_assign resource.
func (r *domainWorkspaceAssignResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "domain_id": schema.StringAttribute{
                Required:    true,
                Description: "The unique identifier of the domain",
            },
            "workspace_ids": schema.ListAttribute{
                Required:    true,
                ElementType: types.StringType,
                Description: "A list of workspace IDs that are to be assigned to the specified domain.",
            },
        },
    }
}

// Define the model for the domain_workspace_assign resource.
type domainWorkspaceAssignResourceModel struct {
    DomainID    types.String `tfsdk:"domain_id"`
    WorkspaceIDs []types.String `tfsdk:"workspace_ids"`
}

// Implement Metadata method.
func (r *domainWorkspaceAssignResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = "microsoftfabric_domain_workspace_assign"
}

// Define the provider.
func NewDomainWorkspaceAssignResource(client *apiclient.APIClient) resource.Resource {
    return &domainWorkspaceAssignResource{client: client}
}

// Implement CRUD operations.
func (r *domainWorkspaceAssignResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // Retrieve values from the plan.
    var plan domainWorkspaceAssignResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Assign workspaces to the domain.
    err := r.assignWorkspaces(plan.DomainID.ValueString(), plan.WorkspaceIDs)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error assigning workspaces to domain",
            "Could not assign workspaces to the domain: "+err.Error(),
        )
        return
    }

    // Set the state with the original workspace IDs.
    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
}

func (r *domainWorkspaceAssignResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    // Implement Read logic if necessary.
}

func (r *domainWorkspaceAssignResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // Retrieve the current state.
    var state domainWorkspaceAssignResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Retrieve the planned state.
    var plan domainWorkspaceAssignResourceModel
    diags = req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Check for duplicate workspace IDs.
    if err := checkDuplicateWorkspaceIDs(plan.WorkspaceIDs); err != nil {
        resp.Diagnostics.AddError("Duplicate workspace IDs found", err.Error())
        return
    }

    // Compute the differences between the current and planned state.
    toAdd := differenceStrings(plan.WorkspaceIDs, state.WorkspaceIDs)
    toRemove := differenceStrings(state.WorkspaceIDs, plan.WorkspaceIDs)

    // Assign new workspaces.
    if len(toAdd) > 0 {
        err := r.assignWorkspaces(plan.DomainID.ValueString(), toAdd)
        if err != nil {
            resp.Diagnostics.AddError(
                "Error assigning workspaces to domain",
                fmt.Sprintf("Could not assign workspaces to the domain: %v", err),
            )
            return
        }
    }

    // Unassign removed workspaces.
    if len(toRemove) > 0 {
        err := r.unassignWorkspaces(plan.DomainID.ValueString(), toRemove)
        if err != nil {
            resp.Diagnostics.AddError(
                "Error unassigning workspaces from domain",
                fmt.Sprintf("Could not unassign workspaces from the domain: %v", err),
            )
            return
        }
    }

    // Update the state with the new assignments.
    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}



func (r *domainWorkspaceAssignResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    // Retrieve values from the state.
    var state domainWorkspaceAssignResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Unassign the workspaces from the domain.
    err := r.unassignWorkspaces(state.DomainID.ValueString(), state.WorkspaceIDs)
    if err != nil {
        resp.Diagnostics.AddError(
            "Error unassigning workspaces from domain",
            "Could not unassign workspaces from the domain: "+err.Error(),
        )
        return
    }

    // No need to set the state as it will be removed.
}

func (r *domainWorkspaceAssignResource) assignWorkspaces(domainID string, workspaceIDs []types.String) error {
    url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/admin/domains/%s/assignWorkspaces", domainID)

    // Initialize the slice for workspaces IDs.
    workspacesIds := make([]string, len(workspaceIDs))

    // Populate the slice with workspace IDs.
    for i, id := range workspaceIDs {
        workspacesIds[i] = id.ValueString()
    }

    body := map[string]interface{}{
        "workspacesIds": workspacesIds,
    }

    bodyBytes, err := json.Marshal(body)
    if err != nil {
        return fmt.Errorf("failed to marshal request body: %v", err)
    }

    _, err = r.client.PostBytes(url, bodyBytes)
    if err != nil {
        return fmt.Errorf("failed to assign workspaces to domain: %v", err)
    }

    return nil
}

func (r *domainWorkspaceAssignResource) unassignWorkspaces(domainID string, workspaceIDs []types.String) error {
    url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/admin/domains/%s/unassignWorkspaces", domainID)

    // Initialize the slice for workspaces IDs.
    workspacesIds := make([]string, len(workspaceIDs))

    // Populate the slice with workspace IDs.
    for i, id := range workspaceIDs {
        workspacesIds[i] = id.ValueString()
    }

    body := map[string]interface{}{
        "workspacesIds": workspacesIds,
    }

    bodyBytes, err := json.Marshal(body)
    if err != nil {
        return fmt.Errorf("failed to marshal request body: %v", err)
    }

    _, err = r.client.PostBytes(url, bodyBytes)
    if err != nil {
        return fmt.Errorf("failed to unassign workspaces from domain: %v", err)
    }

    return nil
}


func differenceStrings(a, b []types.String) []types.String {
    mb := make(map[string]struct{}, len(b))
    for _, x := range b {
        mb[x.ValueString()] = struct{}{}
    }
    var diff []types.String
    for _, x := range a {
        if _, found := mb[x.ValueString()]; !found {
            diff = append(diff, x)
        }
    }
    return diff
}

func checkDuplicateWorkspaceIDs(workspaceIDs []types.String) error {
    idMap := make(map[string]struct{})
    for _, id := range workspaceIDs {
        if _, exists := idMap[id.ValueString()]; exists {
            return fmt.Errorf("duplicate workspace ID found: %s", id.ValueString())
        }
        idMap[id.ValueString()] = struct{}{}
    }
    return nil
}