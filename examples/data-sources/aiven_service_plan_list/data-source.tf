# Use the `aiven_plans` data source to query available service plans and their cloud availability.

data "aiven_service_plan_list" "kafka_plans" {
  project      = "example-project-name"
  service_type = "kafka"
}

# List all available plan names
output "plan_names" {
  value = [for plan in data.aiven_service_plan_list.kafka_plans.service_plans : plan.service_plan]
}

## Find a specific plan
locals {
  business_plan = one([
    for plan in data.aiven_service_plan_list.kafka_plans.service_plans :
      plan if plan.service_plan == "business-4"
  ])
}

output "business_4_cloud_names" {
  value = local.business_plan.cloud_names
}
