package util

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// UnpackCompoundID is a helper function that splits the ID by the separator and sets the attributes in the response.
// It is used in the ImportState method of the resource structs.
func UnpackCompoundID(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
	paths ...string,
) {
	splitResourceID := strings.Split(req.ID, "/")

	for idx, v := range splitResourceID {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(paths[idx]), v)...)
	}
}
