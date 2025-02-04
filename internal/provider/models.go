package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type TargetModel struct {
	ID       types.String      `tfsdk:"id"`
	Endpoint types.String      `tfsdk:"endpoint"`
	Tags     map[string]string `tfsdk:"tags"`
}

type Target struct {
	ID       string
	Endpoint string
	Tags     map[string]string
}
