package provider

import (
	"context"
	"fmt"

	"terraform-provider-microsoftfabric/internal/apiclient"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Define the resource for semantic model user assignment.
type semanticModelUserAssignmentResource struct {
	client *apiclient.APIClient
}

// Define the schema for the semantic model user assignment resource.
func (r *semanticModelUserAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"workspace_id": schema.StringAttribute{
				Required: true,
			},
			"semantic_model_id": schema.StringAttribute{
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
							Required:    true,
							Description: "The principal type (App, Group, None, User)",
						},
					},
				},
			},
		},
	}
}

// Define the model for users.
type userModelSemanticModel struct {
	Email         types.String `tfsdk:"email"`
	Role          types.String `tfsdk:"role"`
	PrincipalType types.String `tfsdk:"principal_type"`
}

// Define the model for semantic model user assignments.
type semanticModelUserAssignmentResourceModel struct {
	WorkspaceID     types.String             `tfsdk:"workspace_id"`
	SemanticModelID types.String             `tfsdk:"semantic_model_id"`
	Users           []userModelSemanticModel `tfsdk:"users"`
}

// Implement the Metadata method.
func (r *semanticModelUserAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "microsoftfabric_semantic_model_user_assignment"
}

// Define the provider function.
func NewSemanticModelUserAssignmentResource(client *apiclient.APIClient) resource.Resource {
	return &semanticModelUserAssignmentResource{client: client}
}

// Implement the Create operation.
func (r *semanticModelUserAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan semanticModelUserAssignmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for duplicate emails.
	if err := checkDuplicateSemanticEmails(plan.Users); err != nil {
		resp.Diagnostics.AddError("Duplicate email found", err.Error())
		return
	}

	// Assign users to semantic model.
	for _, user := range plan.Users {
		err := r.semanticAssignUserToSemanticModel(
			plan.WorkspaceID.ValueString(),
			plan.SemanticModelID.ValueString(),
			user.Email.ValueString(),
			user.Role.ValueString(),
			user.PrincipalType.ValueString(), // Add principal type here
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error assigning user to semantic model",
				fmt.Sprintf("Could not assign user %s to semantic model: %v", user.Email.ValueString(), err),
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

// Implement the Read operation.
func (r *semanticModelUserAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state semanticModelUserAssignmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Logic to read resource from API and update state (often specific to the implementation).
}

// Implement the Update operation.
func (r *semanticModelUserAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from state.
	var state semanticModelUserAssignmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from plan.
	var plan semanticModelUserAssignmentResourceModel
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for duplicate emails.
	if err := checkDuplicateSemanticEmails(plan.Users); err != nil {
		resp.Diagnostics.AddError("Duplicate email found", err.Error())
		return
	}

	// Determine users to be added, updated, and removed.
	toAdd := semanticDifference(plan.Users, state.Users)
	toUpdate := semanticIntersection(plan.Users, state.Users)
	toRemove := semanticDifference(state.Users, plan.Users)

	// Add new users to semantic model.
	for _, user := range toAdd {
		err := r.semanticAssignUserToSemanticModel(
			plan.WorkspaceID.ValueString(),
			plan.SemanticModelID.ValueString(),
			user.Email.ValueString(),
			user.Role.ValueString(),
			user.PrincipalType.ValueString(), // Add principal type here
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error assigning user to semantic model",
				fmt.Sprintf("Could not assign user %s to semantic model: %v", user.Email.ValueString(), err),
			)
			return
		}
	}

	// Update existing users in semantic model.
	for _, user := range toUpdate {
		// Check if the user already has the same role
		existingUser := findUserByEmail(state.Users, user.Email.ValueString())
		if existingUser != nil && existingUser.Role.ValueString() == user.Role.ValueString() &&
			existingUser.PrincipalType.ValueString() == user.PrincipalType.ValueString() {
			continue // Skip if there's no change in user role or principal type
		}

		err := r.semanticUpdateUserInSemanticModel(
			plan.WorkspaceID.ValueString(),
			plan.SemanticModelID.ValueString(),
			user.Email.ValueString(),
			user.Role.ValueString(),
			user.PrincipalType.ValueString(), // Add principal type here
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating user in semantic model",
				fmt.Sprintf("Could not update user %s in semantic model: %v", user.Email.ValueString(), err),
			)
			return
		}
	}

	// Remove users from semantic model.
	for _, user := range toRemove {
		err := r.semanticRemoveUserFromSemanticModel(
			state.WorkspaceID.ValueString(),
			state.SemanticModelID.ValueString(),
			user.Email.ValueString(),
			user.PrincipalType.ValueString(), // Add principal type here
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error removing user from semantic model",
				fmt.Sprintf("Could not remove user %s from semantic model: %v", user.Email.ValueString(), err),
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

// Implement the Delete operation.
func (r *semanticModelUserAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state semanticModelUserAssignmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Remove all users from semantic model.
	for _, user := range state.Users {
		err := r.semanticRemoveUserFromSemanticModel(
			state.WorkspaceID.ValueString(),
			state.SemanticModelID.ValueString(),
			user.Email.ValueString(),
			user.PrincipalType.ValueString(), // Add principal type here
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error removing user from semantic model",
				fmt.Sprintf("Could not remove user %s from semantic model: %v", user.Email.ValueString(), err),
			)
			return
		}
	}
}

// Assign user to semantic model.
func (r *semanticModelUserAssignmentResource) semanticAssignUserToSemanticModel(workspaceID, semanticModelID, userEmail, userRole, principalType string) error {
	body := map[string]interface{}{
		"identifier":                   userEmail,
		"principalType":                principalType,
		"semanticModelUserAccessRight": userRole, // Adjusted to semantic model terminology
	}

	url := fmt.Sprintf("https://api.powerbi.com/v1.0/myorg/workspaces/%s/semanticModels/%s/users", workspaceID, semanticModelID)
	_, err := r.client.Post(url, body) // Updated to use POST for assignment
	if err != nil {
		return err
	}

	return nil
}

// Update user in semantic model.
func (r *semanticModelUserAssignmentResource) semanticUpdateUserInSemanticModel(workspaceID, semanticModelID, userEmail, userRole, principalType string) error {
	body := map[string]interface{}{
		"identifier":                   userEmail,
		"semanticModelUserAccessRight": userRole, // Adjusted to semantic model terminology
		"principalType":                principalType,
	}

	url := fmt.Sprintf("https://api.powerbi.com/v1.0/myorg/workspaces/%s/semanticModels/%s/users", workspaceID, semanticModelID)
	_, err := r.client.Put(url, body)
	if err != nil {
		return fmt.Errorf("failed to update user %s in semantic model: %w", userEmail, err)
	}

	return nil
}

// Remove user from semantic model.
func (r *semanticModelUserAssignmentResource) semanticRemoveUserFromSemanticModel(workspaceID, semanticModelID, userEmail, principalType string) error {
	body := map[string]interface{}{
		"identifier":                   userEmail,
		"semanticModelUserAccessRight": "None", // Setting the access right to None removes the permissions
		"principalType":                principalType,
	}

	url := fmt.Sprintf("https://api.powerbi.com/v1.0/myorg/workspaces/%s/semanticModels/%s/users", workspaceID, semanticModelID)
	_, err := r.client.Put(url, body) // Use PUT to update the user's permissions to None
	if err != nil {
		return err
	}

	return nil
}

// Check for duplicate email addresses in user list.
func checkDuplicateSemanticEmails(users []userModelSemanticModel) error {
	emailSet := make(map[string]struct{})
	for _, user := range users {
		email := user.Email.ValueString()
		if _, exists := emailSet[email]; exists {
			return fmt.Errorf("duplicate email found: %s", email)
		}
		emailSet[email] = struct{}{}
	}
	return nil
}

// Find user by email in the user list.
func findUserByEmail(users []userModelSemanticModel, email string) *userModelSemanticModel {
	for _, user := range users {
		if user.Email.ValueString() == email {
			return &user
		}
	}
	return nil
}

// Determine the difference between two user lists.
func semanticDifference(a, b []userModelSemanticModel) []userModelSemanticModel {
	diff := []userModelSemanticModel{}
	for _, userA := range a {
		found := false
		for _, userB := range b {
			if userA.Email.ValueString() == userB.Email.ValueString() {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, userA)
		}
	}
	return diff
}

// Determine the intersection between two user lists.
func semanticIntersection(a, b []userModelSemanticModel) []userModelSemanticModel {
	inter := []userModelSemanticModel{}
	for _, userA := range a {
		for _, userB := range b {
			if userA.Email.ValueString() == userB.Email.ValueString() {
				inter = append(inter, userA)
				break
			}
		}
	}
	return inter
}
