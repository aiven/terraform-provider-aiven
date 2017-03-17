# Terraform Aiven

[Terraform](https://www.terraform.io/) provider resources for [Aiven.io](https://aiven.io/).

## Sample project

There is a [sample project](sample.tf) which sets up a new project, adds a new service,
creates a new database and sets up a new user for the service.

Make sure you have a look at the [variables](terraform.tfvars.sample) and copy
it over to `terraform.tfvars` with your own credentials.

## Options

### Provider

```
provider "aiven" {
    email = "<AIVEN_EMAIL>"
    password = "<AIVEN_PASSWORD>"
}
```

### Resource Project

**Note**: it is currently not possible to automatically set up a new card for a project. If you already have a card linked to your account, you can get the card id and use it to set up this project.

```
resource "aiven_service" "my-project" {
    project = "<PROJECT_NAME>"
    card_id = "<CARD_ID>"
    cloud = "google-europe-west1"
}
```

### Resource Service

**Note**: to make a new service, you'll have to have set up a project with a valid billing plan.

```
resource "aiven_service" "my-service" {
    project = "<PROJECT_NAME>"
    group_name = "my-group"
    cloud = "google-europe-west1"
    plan = "business-16"
    service_name = "<SERVICE_NAME>"
    service_type = "pg"
}
```

### Resource Database

```
resource "aiven_database" "my-database" {
    project = "<PROJECT_NAME>"
    service_name = "<SERVICE_NAME>"
    database = "<DATABASE_NAME>"
}
```

### Resource ServiceUser

```
resource "aiven_service_user" "my-service-user" {
    project = "<PROJECT_NAME>"
    service_name = "<SERVICE_NAME>"
    username = "<USERNAME>"
}
```
