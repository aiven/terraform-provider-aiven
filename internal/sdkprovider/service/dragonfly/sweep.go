package dragonfly

import (
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func init() {
	sweep.AddServiceSweeper("dragonfly")
}
