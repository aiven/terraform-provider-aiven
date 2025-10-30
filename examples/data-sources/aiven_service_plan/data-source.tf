# Get detailed service plan information including specifications and pricing
data "aiven_service_plan" "kafka_business_plan" {
  project      = "example-project-name"
  service_type = "kafka"
  service_plan = "business-4"
  cloud_name   = "aws-us-east-1"
}

# Output plan specifications
output "plan_specs" {
  value = {
    node_count      = data.aiven_service_plan.kafka_business_plan.node_count
    disk_space_mb   = data.aiven_service_plan.kafka_business_plan.disk_space_mb
    disk_space_cap  = data.aiven_service_plan.kafka_business_plan.disk_space_cap_mb
  }
}

# Output pricing information
output "plan_pricing" {
  value = {
    base_price_usd              = data.aiven_service_plan.kafka_business_plan.base_price_usd
    object_storage_gb_price_usd = data.aiven_service_plan.kafka_business_plan.object_storage_gb_price_usd
  }
}

# Calculate monthly cost estimate (assuming 730 hours per month)
output "estimated_monthly_cost" {
  value = tonumber(data.aiven_service_plan.kafka_business_plan.base_price_usd) * 730
}
