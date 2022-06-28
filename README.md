# Aiven Terraform Provider

The Terraform provider for [Aiven.io](https://aiven.io/), an open source data platform as a service.

**See the [official documentation](https://registry.terraform.io/providers/aiven/aiven/latest/docs) to learn about all the possible services and resources.**

## Quick start

- [Signup for Aiven](https://console.aiven.io/signup?utm_source=github&utm_medium=organic&utm_campaign=terraform&utm_content=signup)
- [Get your authentication token and project name](https://help.aiven.io/en/articles/2059201-authentication-tokens)
- Create a file named `main.tf` with the content below:

```hcl
terraform {
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = "x.y.z" # check out the latest version in the release section
    }
  }
}

provider "aiven" {
  api_token = "your-api-token"
}

resource "aiven_pg" "postgresql" {
  project                = "your-project-name"
  service_name           = "postgresql"
  cloud_name             = "google-europe-west3"
  plan                   = "startup-4"

  termination_protection = true
}

output "postgresql_service_uri" {
  value     = aiven_pg.postgresql.service_uri
  sensitive = true
}
```

- Run these commands in your terminal:

```bash
terraform init
terraform plan
terraform apply
psql "$(terraform output -raw postgresql_service_uri)"
```

Voilà, a PostgreSQL database.

## A word of caution

Recreating stateful services with Terraform will possibly **delete** the service and all its data before creating it again. Whenever the Terraform plan indicates that a service will be **deleted** or **replaced**, a catastrophic action is possibly about to happen.

Some properties, like **project** and the **resource name**, cannot be changed and it will trigger a resource replacement.

To avoid any issues, **please set the `termination_protection` property to `true` on all production services**, it will prevent Terraform to remove the service until the flag is set back to `false` again. While it prevents a service to be deleted, any logical databases, topics or other configurations may be removed **even when this section is enabled**. Be very careful!

## Contributing

Bug reports and patches are very welcome, please post them as GitHub issues and pull requests at https://github.com/aiven/terraform-provider-aiven. Please review the guides below.

- [Contributing guidelines](CONTRIBUTING.md)
- [Code of conduct](CODE_OF_CONDUCT.md)

Please see our [security](SECURITY.md) policy to report any possible vulnerabilities or serious issues.

## License

terraform-provider-aiven is licensed under the MIT license. Full license text is available in the [LICENSE](LICENSE) file. Please note that the project explicitly does not require a CLA (Contributor License Agreement) from its contributors.

## Credits

The original version of the Aiven Terraform provider was written and maintained by [Jelmer Snoeck](https://github.com/jelmersnoeck).
