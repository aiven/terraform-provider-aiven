package kafka

import (
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	sweep.AddServiceSweeper("kafka")
	sweep.AddServiceSweeper("kafka_mirrormaker")
	sweep.AddServiceSweeper("kafka_connect")
	sweep.AddServiceSweeper("kafka_connector")
}
