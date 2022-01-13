//go:build tools
// +build tools

package tools

import (
	_ "github.com/bflad/tfproviderlint/cmd/tfproviderlintx"
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
