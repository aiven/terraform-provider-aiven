# Grafana service
resource "aiven_grafana" "grafana-service1" {
  project = aiven_project.project1.project
  cloud_name = "google-europe-west1"
  plan = "startup-4"
  service_name = "samplegrafana"
  grafana_user_config {
    public_access {
      grafana = true
    }
  }
}

locals {
  # A list of components is sorted on API side in a way that the client should pick first entry based on the query
  components_flat = {for component in aiven_grafana.grafana-service1.components :
  "${component.component}_${component.route}" =>  "${component.host}:${component.port}" if component.usage == "primary"
  }
}

output "grafana_public" {
  value = local.components_flat["grafana_public"]
}