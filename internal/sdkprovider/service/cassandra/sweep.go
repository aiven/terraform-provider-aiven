//go:build sweep

package cassandra

import (
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/sweep"
)

func init() {
	sweep.AddServiceSweeper("cassandra")
}
