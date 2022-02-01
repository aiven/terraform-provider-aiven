resource "aiven_pg" "pg" {
  project                 = data.aiven_project.pr1.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "my-pg1"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  static_ips = toset([
    aiven_static_ip.ips[0].static_ip_address_id,
    aiven_static_ip.ips[1].static_ip_address_id,
    aiven_static_ip.ips[2].static_ip_address_id,
    aiven_static_ip.ips[3].static_ip_address_id,
  ])

  pg_user_config {
    pg_version = 11
    static_ips = true

    public_access {
      pg         = true
      prometheus = false
    }

    pg {
      idle_in_transaction_session_timeout = 900
      log_min_duration_statement          = -1
    }
  }

  timeouts {
    create = "20m"
    update = "15m"
  }
}
