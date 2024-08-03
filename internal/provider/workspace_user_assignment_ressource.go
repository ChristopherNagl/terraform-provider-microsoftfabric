package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-microsoft-fabric/internal/apiclient"
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
					},
				},
			},
		},
	}
}

// Define the model for users.
type userModel struct {
	Email types.String `tfsdk:"email"`
	Role  types.String `tfsdk:"role"`
}

// Define the model for user assignments.
type workspaceUserAssignmentResourceModel struct {
	WorkspaceID types.String `tfsdk:"workspace_id"`
	Users       []userModel  `tfsdk:"users"`
}

// Implement Metadata method.
func (r *workspaceUserAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "microsoftfabric_workspace_user_assignment"
}

// Define the provider.
func NewWorkspaceUserAssignmentResource(client *apiclient.APIClient) resource.Resource {
	return &workspaceUserAssignmentResource{client: client}
}

// Implement Create operation.
func (r *workspaceUserAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan workspaceUserAssignmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for duplicate emails.
	if err := checkDuplicateEmails(plan.Users); err != nil {
		resp.Diagnostics.AddError("Duplicate email found", err.Error())
		return
	}

	// Assign users to workspace.
	for _, user := range plan.Users {
		err := r.assignUserToWorkspace(plan.WorkspaceID.ValueString(), user.Email.ValueString(), user.Role.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error assigning user to workspace",
				fmt.Sprintf("Could not assign user %s to workspace: %v", user.Email.ValueString(), err),
			)
			return
		}
	}

	// Set state.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Implement Read operation.
func (r *workspaceUserAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve values from state.
	var state workspaceUserAssignmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Logic to read resource from API and update state
	// Since this is a read function, typically you would query the API for current state.
	// and update the Terraform state with the fetched data.
}

// Implement Update operation.
func (r *workspaceUserAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from state.
	var state workspaceUserAssignmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from plan.
	var plan workspaceUserAssignmentResourceModel
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for duplicate emails.
	if err := checkDuplicateEmails(plan.Users); err != nil {
		resp.Diagnostics.AddError("Duplicate email found", err.Error())
		return
	}

	// Determine users to be added, updated, and removed.
	toAdd := difference(plan.Users, state.Users)
	toUpdate := intersection(plan.Users, state.Users)
	toRemove := difference(state.Users, plan.Users)

	// Add new users to workspace.
	for _, user := range toAdd {
		err := r.assignUserToWorkspace(plan.WorkspaceID.ValueString(), user.Email.ValueString(), user.Role.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error assigning user to workspace",
				fmt.Sprintf("Could not assign user %s to workspace: %v", user.Email.ValueString(), err),
			)
			return
		}
	}

	// Update existing users in workspace.
	for _, user := range toUpdate {
		err := r.updateUserInWorkspace(plan.WorkspaceID.ValueString(), user.Email.ValueString(), user.Role.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating user in workspace",
				fmt.Sprintf("Could not update user %s in workspace: %v", user.Email.ValueString(), err),
			)
			return
		}
	}

	// Remove users from workspace.
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

	// Set state.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Implement Delete operation.
func (r *workspaceUserAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state workspaceUserAssignmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Remove all users from workspace.
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
func (r *workspaceUserAssignmentResource) assignUserToWorkspace(workspaceID, userEmail, userRole string) error {
	body := map[string]string{
		"identifier":           userEmail,
		"groupUserAccessRight": userRole,
		"principalType":        "User",
	}

	url := fmt.Sprintf("https://api.powerbi.com/v1.0/myorg/groups/%s/users", workspaceID)
	_, err := r.client.Post(url, body)
	if err != nil {
		// Return the error received from the Post method.
		return err
	}

	// If no errors occurred, return nil.
	return nil
}

// Implement user update function.
func (r *workspaceUserAssignmentResource) updateUserInWorkspace(workspaceID, userEmail, userRole string) error {
	body := map[string]string{
		"identifier":           userEmail,
		"groupUserAccessRight": userRole,
		"principalType":        "User",
	}

	url := fmt.Sprintf("https://api.powerbi.com/v1.0/myorg/groups/%s/users", workspaceID)
	_, err := r.client.Put(url, body)
	if err != nil {
		// Return the error received from the Put method.
		return err
	}

	// If no errors occurred, return nil.
	return nil
}

// Implement user removal function.
func (r *workspaceUserAssignmentResource) removeUserFromWorkspace(workspaceID, userEmail string) error {
	url := fmt.Sprintf("https://api.powerbi.com/v1.0/myorg/groups/%s/users/%s", workspaceID, userEmail)
	err := r.client.Delete(url)
	if err != nil {
		// Return the error received from the Delete method.
		return err
	}

	// If no errors occurred, return nil.
	return nil
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
	for _, x := range b {
		mb[x.Email.ValueString()] = true
	}
	var diff []userModel
	for _, x := range a {
		if !mb[x.Email.ValueString()] {
			diff = append(diff, x)
		}
	}
	return diff
}

func intersection(a, b []userModel) []userModel {
	mb := make(map[string]bool, len(b))
	for _, x := range b {
		mb[x.Email.ValueString()] = true
	}
	var inter []userModel
	for _, x := range a {
		if mb[x.Email.ValueString()] {
			inter = append(inter, x)
		}
	}
	return inter
}
