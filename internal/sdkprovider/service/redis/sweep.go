//go:build sweep

package redis

import (
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/sweep"
)

func init() {
	sweep.AddServiceSweeper("redis")
}
