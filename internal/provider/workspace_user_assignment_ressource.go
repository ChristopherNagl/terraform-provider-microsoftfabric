package provider

import (
    "context"
    "fmt"
    "sort"
    
    "terraform-provider-microsoftfabric/internal/apiclient"

    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

// Define the resource.
type workspaceUserAssignmentResource struct {
    client *apiclient.APIClient
}

// Define the schema.
func (r *workspaceUserAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "workspace_id": schema.StringAttribute{
                Required: true,
            },
            "users": schema.ListNestedAttribute{
                Required: true,
                NestedObject: schema.NestedAttributeObject{
                    Attributes: map[string]schema.Attribute{
                        "email": schema.StringAttribute{
                            Required: true,
                        },
                        "role": schema.StringAttribute{
                            Required: true,
                        },
                        "principal_type": schema.StringAttribute{
                            Required: true,
                        },
                    },
                },
            },
        },
    }
}

// Define the model for users.
type userModel struct {
    Email         types.String `tfsdk:"email"`
    Role          types.String `tfsdk:"role"`
    PrincipalType types.String `tfsdk:"principal_type"`
}

// Define the model for user assignments.
type workspaceUserAssignmentResourceModel struct {
    WorkspaceID types.String `tfsdk:"workspace_id"`
    Users       []userModel  `tfsdk:"users"`
}

// Implement the Metadata method.
func (r *workspaceUserAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = "microsoftfabric_workspace_user_assignment"
}

// Define the provider.
func NewWorkspaceUserAssignmentResource(client *apiclient.APIClient) resource.Resource {
    return &workspaceUserAssignmentResource{client: client}
}

// Implement Create operation.
func (r *workspaceUserAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan workspaceUserAssignmentResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    if err := checkDuplicateEmails(plan.Users); err != nil {
        resp.Diagnostics.AddError("Duplicate email found", err.Error())
        return
    }

    SortUsers(&plan.Users)

    for _, user := range plan.Users {
        principalType := user.PrincipalType.ValueString()
        if principalType == "" {
            resp.Diagnostics.AddError("Missing principal_type", "The principal_type field must be provided.")
            return
        }

        err := r.assignUserToWorkspace(plan.WorkspaceID.ValueString(), user.Email.ValueString(), user.Role.ValueString(), principalType)
        if err != nil {
            resp.Diagnostics.AddError(
                "Error assigning user to workspace",
                fmt.Sprintf("Could not assign user %s to workspace: %v", user.Email.ValueString(), err),
            )
            return
        }
    }

    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
}

// Implement Read operation.
func (r *workspaceUserAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    // Implementation goes here...
}

// Implement Update operation.
func (r *workspaceUserAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var state, plan workspaceUserAssignmentResourceModel
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

    if err := checkDuplicateEmails(plan.Users); err != nil {
        resp.Diagnostics.AddError("Duplicate email found", err.Error())
        return
    }

    SortUsers(&plan.Users)
    SortUsers(&state.Users)

    toAdd := difference(plan.Users, state.Users)
    toUpdate, _ := intersection(plan.Users, state.Users)
    toRemove := difference(state.Users, plan.Users)

    for _, user := range toAdd {
        principalType := user.PrincipalType.ValueString()
        if principalType == "" {
            resp.Diagnostics.AddError("Missing principal_type", "The principal_type field must be provided.")
            return
        }

        err := r.assignUserToWorkspace(plan.WorkspaceID.ValueString(), user.Email.ValueString(), user.Role.ValueString(), principalType)
        if err != nil {
            resp.Diagnostics.AddError(
                "Error assigning user to workspace",
                fmt.Sprintf("Could not assign user %s to workspace: %v", user.Email.ValueString(), err),
            )
            return
        }
    }

    for _, user := range toUpdate {
        principalType := user.PrincipalType.ValueString()
        err := r.updateUserInWorkspace(plan.WorkspaceID.ValueString(), user.Email.ValueString(), user.Role.ValueString(), principalType)
        if err != nil {
            resp.Diagnostics.AddError(
                "Error updating user in workspace",
                fmt.Sprintf("Could not update user %s in workspace: %v", user.Email.ValueString(), err),
            )
            return
        }
    }

    for _, user := range toRemove {
        err := r.removeUserFromWorkspace(state.WorkspaceID.ValueString(), user.Email.ValueString())
        if err != nil {
            resp.Diagnostics.AddError(
                "Error removing user from workspace",
                fmt.Sprintf("Could not remove user %s from workspace: %v", user.Email.ValueString(), err),
            )
            return
        }
    }

    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
}

// Implement Delete operation.
func (r *workspaceUserAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state workspaceUserAssignmentResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    for _, user := range state.Users {
        err := r.removeUserFromWorkspace(state.WorkspaceID.ValueString(), user.Email.ValueString())
        if err != nil {
            resp.Diagnostics.AddError(
                "Error removing user from workspace",
                fmt.Sprintf("Could not remove user %s from workspace: %v", user.Email.ValueString(), err),
            )
            return
        }
    }
}

// Implement user assignment function.
func (r *workspaceUserAssignmentResource) assignUserToWorkspace(workspaceID, userEmail, userRole, principalType string) error {
    body := map[string]string{
        "identifier":           userEmail,
        "groupUserAccessRight": userRole,
        "principalType":        principalType,
    }

    url := fmt.Sprintf("https://api.powerbi.com/v1.0/myorg/groups/%s/users", workspaceID)
    _, err := r.client.Post(url, body)
    return err
}

// Implement user update function.
func (r *workspaceUserAssignmentResource) updateUserInWorkspace(workspaceID, userEmail, userRole, principalType string) error {
    body := map[string]string{
        "identifier":           userEmail,
        "groupUserAccessRight": userRole,
        "principalType":        principalType,
    }

    url := fmt.Sprintf("https://api.powerbi.com/v1.0/myorg/groups/%s/users", workspaceID)
    _, err := r.client.Put(url, body)
    return err
}

// Implement user removal function.
func (r *workspaceUserAssignmentResource) removeUserFromWorkspace(workspaceID, userEmail string) error {
    url := fmt.Sprintf("https://api.powerbi.com/v1.0/myorg/groups/%s/users/%s", workspaceID, userEmail)
    return r.client.Delete(url)
}

// Check for duplicate emails.
func checkDuplicateEmails(users []userModel) error {
    emailSet := make(map[string]struct{})
    for _, user := range users {
        if _, exists := emailSet[user.Email.ValueString()]; exists {
            return fmt.Errorf("duplicate email found: %s", user.Email.ValueString())
        }
        emailSet[user.Email.ValueString()] = struct{}{}
    }
    return nil
}

// Utility functions to compute differences and intersections between user slices.
func difference(a, b []userModel) []userModel {
    mb := make(map[string]bool, len(b))
    for _, user := range b {
        mb[user.Email.ValueString()] = true
    }
    var diff []userModel
    for _, user := range a {
        if !mb[user.Email.ValueString()] {
            diff = append(diff, user)
        }
    }
    return diff
}

func intersection(a, b []userModel) ([]userModel, []userModel) {
    mb := make(map[string]userModel, len(b))
    for _, user := range b {
        mb[user.Email.ValueString()] = user
    }
    var inter []userModel
    for _, user := range a {
        if _, exists := mb[user.Email.ValueString()]; exists {
            inter = append(inter, user)
        }
    }
    return inter, []userModel{}
}

// Sort users in-place by email.
func SortUsers(users *[]userModel) {
    sort.Slice(*users, func(i, j int) bool {
        return (*users)[i].Email.ValueString() < (*users)[j].Email.ValueString()
    })
}