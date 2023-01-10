//go:build sweep

package opensearch

import (
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	sweep.AddServiceSweeper("opensearch")
}
