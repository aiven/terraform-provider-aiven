//go:build sweep

package kafka

import (
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/sweep"
)

func init() {
	sweep.AddServiceSweeper("kafka")
	sweep.AddServiceSweeper("kafka_mirrormaker")
	sweep.AddServiceSweeper("kafka_connect")
}
