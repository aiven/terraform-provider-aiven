//go:build sweep

package mysql

import (
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/sweep"
)

func init() {
	sweep.AddServiceSweeper("mysql")
}
