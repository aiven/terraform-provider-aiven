package byocawsprovision

import "github.com/aiven/terraform-provider-aiven/internal/plugin/byoc"

var modifyPlan = byoc.RejectIdentityChange("aws_iam_role_arn")
