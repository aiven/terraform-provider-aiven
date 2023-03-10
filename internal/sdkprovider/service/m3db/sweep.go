//go:build sweep

package m3db

import (
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/sweep"
)

func init() {
	sweep.AddServiceSweeper("m3db")
	sweep.AddServiceSweeper("m3aggregator")
}
