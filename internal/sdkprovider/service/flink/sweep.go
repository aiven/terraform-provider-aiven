//go:build sweep

package flink

import (
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/sweep"
)

func init() {
	sweep.AddServiceSweeper("flink")
}
