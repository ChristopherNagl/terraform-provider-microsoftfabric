package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-microsoftfabric/internal/apiclient"
	"time"
)

type domainResource struct {
	client *apiclient.APIClient
}

func (r *domainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"parent_domain_id": schema.StringAttribute{
				Optional: true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

type domainResourceModel struct {
	ID             types.String `tfsdk:"id"`
	DisplayName    types.String `tfsdk:"display_name"`
	Description    types.String `tfsdk:"description"`
	ParentDomainID types.String `tfsdk:"parent_domain_id"`
	LastUpdated    types.String `tfsdk:"last_updated"`
}

func (r *domainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "microsoftfabric_domain"
}

func NewDomainResource(client *apiclient.APIClient) resource.Resource {
	return &domainResource{client: client}
}

func (r *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID, err := r.createDomain(plan.DisplayName.ValueString(), plan.Description.ValueString(), plan.ParentDomainID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating domain", "Could not create domain: "+err.Error())
		return
	}

	plan.ID = types.StringValue(domainID)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *domainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state domainResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.readDomain(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading domain", "Could not read domain: "+err.Error())
		return
	}

	// Handle potential nil values safely
	if domain["displayName"] != "" {
		state.DisplayName = types.StringValue(domain["displayName"].(string))
	}
	if domain["description"] != "" {
		state.Description = types.StringValue(domain["description"].(string))
	}
	if domain["parentDomainId"] != "" {
		state.ParentDomainID = types.StringValue(domain["parentDomainId"].(string))
	}
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}


func (r *domainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan domainResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state domainResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.updateDomain(state.ID.ValueString(), plan.DisplayName.ValueString(), plan.Description.ValueString(), plan.ParentDomainID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating domain", "Could not update domain: "+err.Error())
		return
	}

	plan.ID = state.ID
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.deleteDomain(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting domain", "Could not delete domain: "+err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *domainResource) createDomain(displayName, description, parentDomainID string) (string, error) {
	url := "https://api.fabric.microsoft.com/v1/admin/domains"
	body := map[string]string{
		"displayName":    displayName,
		"description":    description,
		"parentDomainId": parentDomainID,
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

func (r *domainResource) readDomain(id string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/admin/domains/%s", id)

	respBody, err := r.client.Get(url)
	if err != nil {
		return nil, err
	}

	// Check for nil values and safely handle type assertions
	domain := map[string]interface{}{
		"displayName":    "",
		"description":    "",
		"parentDomainId": "",
	}

	if displayName, ok := respBody["displayName"].(string); ok {
		domain["displayName"] = displayName
	}
	if description, ok := respBody["description"].(string); ok {
		domain["description"] = description
	}
	if parentDomainId, ok := respBody["parentDomainId"].(string); ok {
		domain["parentDomainId"] = parentDomainId
	}

	return domain, nil
}


func (r *domainResource) updateDomain(id, displayName, description, parentDomainID string) error {
	url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/admin/domains/%s", id)
	body := map[string]string{
		"displayName":    displayName,
		"description":    description,
		"parentDomainId": parentDomainID,
	}

	_, err := r.client.Patch(url, body)
	if err != nil {
		return err
	}

	return nil
}

func (r *domainResource) deleteDomain(id string) error {
	url := fmt.Sprintf("https://api.fabric.microsoft.com/v1/admin/domains/%s", id)

	err := r.client.Delete(url)
	if err != nil {
		return err
	}

	return nil
}
