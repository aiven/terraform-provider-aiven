package byocazureprovision

import "github.com/aiven/terraform-provider-aiven/internal/plugin/byoc"

var modifyPlan = byoc.RejectIdentityChange("azure_subscription_id", "azure_tenant_id")
