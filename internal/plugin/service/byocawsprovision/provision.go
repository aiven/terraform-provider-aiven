package byocawsprovision

import "github.com/aiven/terraform-provider-aiven/internal/plugin/byoc"

func init() {
	ResourceOptions.Delete = byoc.DeleteProvisioned(deleteView)
}
