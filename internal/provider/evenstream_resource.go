package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-microsoft-fabric/internal/apiclient"
	"time"
)

// Define the resource.
type eventstreamResource struct {
	client *apiclient.APIClient
}

// Define the schema.
func (r *eventstreamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"workspace_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Required: true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Define the model.
type eventstreamResourceModel struct {
	ID          types.String `tfsdk:"id"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	LastUpdated types.String `tfsdk:"last_updated"`
}

// Implement Metadata method.
func (r *eventstreamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "microsoftfabric_eventstream"
}

// Define the provider.
func NewEventStreamResource(client *apiclient.APIClient) resource.Resource {
	return &eventstreamResource{client: client}
}

// Implement CRUD operations.
func (r *eventstreamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan eventstreamResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create event stream.
	eventStreamID, err := r.createEventStream(plan.WorkspaceID.ValueString(), plan.Name.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating event stream",
			"Could not create event stream: "+err.Error(),
		)
		return
	}

	// Set ID and LastUpdated fields.
	plan.ID = types.StringValue(eventStreamID)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *eventstreamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Retrieve ID from state.
	var state eventstreamResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read event stream.
	eventStream, err := r.readEventStream(state.WorkspaceID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading event stream",
			"Could not read event stream: "+err.Error(),
		)
		return
	}

	// Check for the presence of the "displayName" key and ensure it's a string.
	name, ok := eventStream["displayName"].(string)
	if !ok {
		resp.Diagnostics.AddError(
			"Error reading event stream",
			"Unexpected response format: 'displayName' key not found or not a string",
		)
		return
	}

	// Set state.
	state.Name = types.StringValue(name)
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *eventstreamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan eventstreamResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve ID from state.
	var state eventstreamResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update event stream.
	err := r.updateEventStream(state.WorkspaceID.ValueString(), state.ID.ValueString(), plan.Name.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating event stream",
			"Could not update event stream: "+err.Error(),
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

func (r *eventstreamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve ID from state.
	var state eventstreamResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete event stream.
	err := r.deleteEventStream(state.WorkspaceID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting event stream",
			"Could not delete event stream: "+err.Error(),
		)
		return
	}

	// Remove resource from state.
	resp.State.RemoveResource(ctx)
}

// Helper functions for event stream operations.

func (r *eventstreamResource) createEventStream(workspaceID, name, description string) (string, error) {
	url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/eventstreams", workspaceID)
	body := map[string]string{
		"displayName": name,
		"description": description,
	}

	responseBody, err := r.client.PostWithOperationCheck(url, body)
	if err != nil {
		return "", err
	}

	// Log the full response body for debugging.
	fmt.Printf("Full Response Body: %+v\n", responseBody)

	eventStreamID, ok := responseBody["id"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected response format: 'id' key not found")
	}

	return eventStreamID, nil
}

func (r *eventstreamResource) readEventStream(workspaceID, eventStreamID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/eventstreams/%s", workspaceID, eventStreamID)
	return r.client.Get(url)
}

func (r *eventstreamResource) updateEventStream(workspaceID, eventStreamID, name, description string) error {
	url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/eventstreams/%s", workspaceID, eventStreamID)
	body := map[string]string{
		"displayName": name,
		"description": description,
	}

	_, err := r.client.Patch(url, body)
	return err
}

func (r *eventstreamResource) deleteEventStream(workspaceID, eventStreamID string) error {
	url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/workspaces/%s/eventstreams/%s", workspaceID, eventStreamID)
	err := r.client.Delete(url)
	if err != nil {
		return err
	}

	// Since the delete operation doesn't return a JSON body, we don't need to handle any response body here.
	return nil
}
