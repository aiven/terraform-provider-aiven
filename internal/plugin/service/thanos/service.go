package thanos

import "github.com/aiven/terraform-provider-aiven/internal/plugin/util/serviceutil"

var svc = serviceutil.Service{Type: "thanos"}

var (
	expandModifier  = svc.ExpandModifier
	flattenModifier = svc.FlattenModifier
	modifyPlan      = svc.ModifyPlan
)

func init() {
	svc.Wrap(&ResourceOptions, &DataSourceOptions)
}
