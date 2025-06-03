---
page_title: "Upgrade Aiven Provider for Terraform from v1 to v2"
---

# Upgrade Aiven Provider for Terraform from v1 to v2

Version 2 of the Aiven Terraform Provider was released in [October of 2020](https://aiven.io/blog/aiven-terraform-provider-v2-release).

## Major changes in v2

The main changes in v2 are:

-   Billing Groups have been introduced instead of needing to provide
    Card ID
-   Work is being done to deprecate `aiven_service` in order to support
    individual service configuration better, using `aiven_kafka` for
    example
-   New services are available in the updated Provider, such as
    `aiven_flink` and `aiven_opensearch`.

More information is available in the [detailed changelog](https://github.com/aiven/terraform-provider-aiven/blob/master/CHANGELOG.md).

## Upgrade to v2

It's recommended to [upgrade Terraform](https://developer.hashicorp.com/terraform/language/upgrade-guides) to the latest version.

Update the Aiven Terraform Provider by editing the version in the providers block:

```hcl
terraform {
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">=3.0.0, < 4.0.0"
    }
  }
}
```

## Upgrade Terraform 0.12 to 0.13

Between v0.12 and v0.13, the syntax of Terraform files changed. If you
have the older syntax, update it:

1.  Upgrade your modules first by installing Terraform v0.13.x.

2.  In the `terraform` block, update the `required_version` from `>= 0.12` to `>= 0.13`.

3.  To replace old refernces to the new format, update your state file by running:

    ```bash
    terraform state replace-provider registry.terraform.io/-/aiven registry.terraform.io/aiven/aiven
    ```

4.  To see other fixes recommended by HashiCorp, run `terraform 0.13upgrade` .

5.  Run `terraform init -upgrade`.

6.  Remove the old Terraform folder by running `rm -rf ~/.terraform.d`.

7.  Check your changes by running `terraform plan`

## Upgrade Terraform from 0.13 or later

Any version above 0.13 can be upgraded to latest without any special
steps.

-> **Note**
If you are using Aiven Terraform Provider v1 with Terraform 0.14
[`dev_overrides`](https://www.terraform.io/cli/config/config-file), add Aiven Provider to the `exclude` block or remove
`dev_overrides` completely.

1.  [Upgrade to the latest version of Terraform](https://www.terraform.io/upgrade-guides).
2.  Run `terraform init -upgrade`.
3.  Run `terraform plan`.

## Update to service specific resource syntax

V2 of the Aiven Terraform Provider moves away from using `aiven_service`
as a resource, and instead provides specific service resources such as
`aiven_kafka`. To avoid destroying existing resources, use the following steps
to migrate to the new resources.

~> **Warning**
Running `terraform state mv <a> <b>` is not
recommended because these are different resource types.

1.  Optional: Back up your Terraform state file.
2.  Replace the `aiven_service` resources with the new service resource.
3.  Remove the old resources from the state file.
4.  Import existing services.
5.  To preview the changes, run `terraform plan`.
6.  To apply the new configuration, run `terraform apply`.

For example, to change from the old `aiven_service` to the new `aiven_kafka`
esource, replace the resource and the old remove the `service_type` field.
Update any references to `aiven_service.kafka.*` with `aiven_kafka.kafka.*`.

The output is similar to the following:

    ```bash
    - resource "aiven_service" "kafka" {
    -    service_type            = "kafka"
    + resource "aiven_kafka" "kafka" {
        ...
    }
    resource "aiven_service_user" "kafka_user" {
      project      = var.aiven_project_name
    -  service_name = aiven_service.kafka.service_name
    +  service_name = aiven_kafka.kafka.service_name
      username     = var.kafka_user_name
    }
    ```
