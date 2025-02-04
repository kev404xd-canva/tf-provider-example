package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TargetResource struct {
	client *APIClient
}

func NewTargetResource() resource.Resource {
	return &TargetResource{}
}

func (r *TargetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_target"
}

func (r *TargetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"endpoint": schema.StringAttribute{
				Required: true,
			},
			"tags": schema.MapAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
		},
	}
}

func (r *TargetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *APIClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *TargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan TargetModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new target
	target := Target{
		Endpoint: plan.Endpoint.ValueString(),
		Tags:     plan.Tags,
	}

	createdTarget, err := r.client.CreateTarget(target)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Failed to create target: %s", err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(createdTarget.ID)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// The provider uses the Read to retrieve the resource's information and update the Terraform state
// to reflect the resource's current state. The provider invokes this function before every plan to generate
// an accurate diff between the resource's current state and the configuration.
func (r *TargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current item state
	var state TargetModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed target from registry
	target, err := r.client.GetTarget(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Failed to read target: %s", err),
		)
		return
	}

	// Overwrite target with refreshed state (id should not change)
	state.Endpoint = types.StringValue(target.Endpoint)
	state.Tags = target.Tags

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *TargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan TargetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	target := Target{
		ID:       plan.ID.ValueString(),
		Endpoint: plan.Endpoint.ValueString(),
		Tags:     plan.Tags,
	}

	_, err := r.client.UpdateTarget(target)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to update target: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *TargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state TargetModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTarget(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Failed to delete target: %s", err),
		)
		return
	}
}
