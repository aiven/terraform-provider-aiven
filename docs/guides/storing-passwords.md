---
page_title: "Manage passwords without storing them in state"
---

# Manage passwords without storing them in state

By default, all passwords are stored in your Terraform state file. To avoid storing passwords in state, use write-only password fields with ephemeral resources.

Write-only password fields are available for:
- `aiven_clickhouse_user`

~> **Requirement**
Write-only arguments support requires Terraform 1.11 or later.

## Create a ClickHouse user with a write-only password

Use the `password_wo` and `password_wo_version` fields to set a custom password that won't be stored in state:

```hcl
resource "aiven_clickhouse" "example" {
  project      = var.aiven_project_name
  service_name = "example-clickhouse"
  cloud_name   = "google-europe-west1"
  plan         = "startup-16"
}

ephemeral "random_password" "clickhouse_password" {
  length  = 24
  special = true
}

resource "aiven_clickhouse_user" "example" {
  project             = var.aiven_project_name
  service_name        = aiven_clickhouse.example.service_name
  username            = "app_user"
  password_wo         = ephemeral.random_password.clickhouse_password.result
  password_wo_version = 1
}
```

The `password_wo` field is write-only and never appears in state. Only the `password_wo_version` number is stored.

## Store passwords in a secret manager

To use the password in your application, store it in a secret manager during the same Terraform run:

```hcl
# AWS Secrets Manager example
resource "aws_secretsmanager_secret" "clickhouse_password" {
  name = "clickhouse-app-user-password"
}

resource "aws_secretsmanager_secret_version" "clickhouse_password" {
  secret_id     = aws_secretsmanager_secret.clickhouse_password.id
  secret_string = ephemeral.random_password.clickhouse_password.result
}
```

Your application can then retrieve the password from the secret manager without it ever being stored in Terraform state.

## Rotate passwords

To rotate a password, increment the `password_wo_version` number:

```hcl
resource "aiven_clickhouse_user" "example" {
  project             = var.aiven_project_name
  service_name        = aiven_clickhouse.example.service_name
  username            = "app_user"
  password_wo         = ephemeral.random_password.clickhouse_password.result
  password_wo_version = 2  # Incremented from 1 to trigger rotation
}
```

When you increment the version number and apply the changes, the password is updated without being stored in state.

## Migrate from generated passwords

If you have existing users with generated passwords, you can migrate to write-only passwords:

1. Add the `password_wo` and `password_wo_version` fields to your existing user resource:

   ```hcl
   resource "aiven_clickhouse_user" "example" {
     project             = var.aiven_project_name
     service_name        = aiven_clickhouse.example.service_name
     username            = "app_user"
     password_wo         = ephemeral.random_password.clickhouse_password.result
     password_wo_version = 1
   }
   ```

2. Apply the changes:

   ```bash
   terraform apply
   ```

The old password will be removed from state and replaced with only the version number.
