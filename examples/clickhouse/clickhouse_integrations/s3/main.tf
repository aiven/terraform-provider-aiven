resource "aiven_clickhouse" "main" {
  project                 = var.aiven_project
  cloud_name              = var.cloud
  plan                    = var.plan
  service_name            = "${var.service_name_prefix}-clickhouse"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  clickhouse_user_config {
    public_access {
      clickhouse       = true
      clickhouse_https = true
      clickhouse_mysql = false
    }
  }
}


resource "aiven_clickhouse_database" "main" {
  project      = var.aiven_project
  service_name = aiven_clickhouse.main.service_name
  name         = "main_db"
}

# ClickHouse users
resource "aiven_clickhouse_user" "app_user" {
  project      = var.aiven_project
  service_name = aiven_clickhouse.main.service_name
  username     = "app_user"
}

resource "aiven_clickhouse_user" "demo_user" {
  project      = var.aiven_project
  service_name = aiven_clickhouse.main.service_name
  username     = "demo_user"
}

# Service integration endpoint for S3
resource "aiven_service_integration_endpoint" "s3_endpoint" {
  project       = var.aiven_project
  endpoint_name = "${var.service_name_prefix}-s3-endpoint"
  endpoint_type = "external_aws_s3"

  external_aws_s3_user_config {
    url               = "https://${aws_s3_bucket.clickhouse_data.bucket}.s3.${var.aws_region}.amazonaws.com/"
    access_key_id     = aws_iam_access_key.clickhouse_user.id
    secret_access_key = aws_iam_access_key.clickhouse_user.secret
  }
}

# Service integration for managed credentials
# This allows specified access S3
resource "aiven_service_integration" "clickhouse_s3" {
  project                  = var.aiven_project
  integration_type         = "clickhouse_credentials"
  source_endpoint_id       = aiven_service_integration_endpoint.s3_endpoint.id
  destination_service_name = aiven_clickhouse.main.service_name

  # IMPORTANT: This grants block gives users access to the named collection
  clickhouse_credentials_user_config {
    grants {
      user = aiven_clickhouse_user.app_user.username
    }
    grants {
      user = aiven_clickhouse_user.demo_user.username
    }
  }

  depends_on = [
    aiven_clickhouse.main,
    aiven_service_integration_endpoint.s3_endpoint,
    aiven_clickhouse_user.app_user,
    aiven_clickhouse_user.demo_user
  ]
}


# Grants for users - S3 access and temporary table creation
resource "aiven_clickhouse_grant" "user_grants" {
  for_each = {
    app_user  = aiven_clickhouse_user.app_user.username
    demo_user = aiven_clickhouse_user.demo_user.username
  }

  project      = var.aiven_project
  service_name = aiven_clickhouse.main.service_name
  user         = each.value

  privilege_grant {
    privilege = "CREATE TEMPORARY TABLE"
    database  = "*"
  }

  privilege_grant {
    privilege = "S3"
    database  = "*"
  }

  depends_on = [
    aiven_clickhouse.main,
    aiven_clickhouse_user.app_user,
    aiven_clickhouse_user.demo_user
  ]
}
