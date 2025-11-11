# Use the `aiven_plans` data source to query available service plans and their cloud availability.

data "aiven_service_plan_list" "kafka_plans" {
  project      = "example-project-name"
  service_type = "kafka"
}

# List all available plan names
output "plan_names" {
  value = [for plan in data.aiven_service_plan_list.kafka_plans.service_plans : plan.service_plan]
}

# Find a specific plan
locals {
  business_plan = one([
    for plan in data.aiven_service_plan_list.kafka_plans.service_plans :
      plan if plan.service_plan == "business-4"
  ])
}

# All available region names (cloud names) for the "business-4" plan
output "business_4_region_names" {
  value = keys(local.business_plan.regions)
}

# CPU count for the "business-4" plan in "aws-eu-west-1"
output "business_4_aws_eu_west_1_cpu" {
  value = local.business_plan.regions["aws-eu-west-1"].node_cpu_count
}

# Memory amount for the "business-4" plan in "aws-eu-west-1"
output "business_4_aws_eu_west_1_memory" {
  value = local.business_plan.regions["aws-eu-west-1"].node_memory_mb
}
