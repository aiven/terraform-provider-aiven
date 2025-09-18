resource "aiven_project" "project" {
  project   = var.aiven_project
  parent_id = var.aiven_organization
}

resource "aiven_static_ip" "ips" {
  count = 6

  project    = aiven_project.project.project
  cloud_name = "google-europe-west1"
}

resource "aiven_pg" "pg" {
  project      = aiven_project.project.project
  cloud_name   = "google-europe-west1"
  plan         = "startup-4"
  service_name = "pg-with-static-ips"

  static_ips = toset([
    aiven_static_ip.ips[0].static_ip_address_id,
    aiven_static_ip.ips[1].static_ip_address_id,
    aiven_static_ip.ips[2].static_ip_address_id,
    aiven_static_ip.ips[3].static_ip_address_id,
    aiven_static_ip.ips[4].static_ip_address_id,
    aiven_static_ip.ips[5].static_ip_address_id,
  ])

  pg_user_config {
    static_ips = true
  }
}
