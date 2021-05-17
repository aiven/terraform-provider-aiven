# Aiven Terraform Provider
The Terraform provider for [Aiven.io](https://aiven.io/), an open source data platform as a service. 

**See the [official documentation](https://registry.terraform.io/providers/aiven/aiven/latest/docs) to learn about all the possible services and resources.**

## 🚨 A word of caution 🚨
Recreating stateful services with Terraform will possibly **delete** the service and all its data before creating it again. Whenever the Terraform plan indicates that a service will be **deleted** or **replaced**, a catastrophic action is possibly about to happen.

Some properties, like **project** and the **resource name**, cannot be changed and it will trigger a resource replacement.

To avoid any issues, **please set the `termination_protection` property to `true` on all production services**, it will prevent Terraform to remove the service until the flag is set back to `false` again. While it prevents a service to be deleted, any logical databases, topics or other configurations may be removed **even when this section is enabled**. Be very careful! 

## Quick Start
- [Signup for Aiven](https://console.aiven.io/signup)
- [Get your authentication token and project name](https://help.aiven.io/en/articles/2059201-authentication-tokens)
- Create a file named `main.tf` with the content below:
```hcl
terraform {
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = "2.1.12" # check out the latest version in the release section
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

- Run in your terminal:
```bash
$ terraform init
$ terraform plan
$ terraform apply
$ psql "$(terraform output -raw postgresql_service_uri)"
```

Voilà, a PostgreSQL database.

## Developing
### Requirements
- [Terraform](https://www.terraform.io/downloads.html) v0.12.X or greater
- [Go](https://golang.org/doc/install) 1.16.X or greater
### Cloning
```bash
$ git clone https://github.com/aiven/terraform-provider-aiven.git
```
### Building
Run the command below. It will generate the binaries under the `plugins/$OS_$ARCH` folder.
```bash
$ make bins
```

### Testing
Run the tests with the command below:
```bash
$ make test
```

Run the acceptance tests with the commands below:
```bash
$ export AIVEN_TOKEN="your-token"
$ export AIVEN_PROJECT_NAME="your-project-name"

$ make testacc

# or run a specific acceptance test
$ make testacc TESTARGS='-run=TestAccAiven_kafka'
```

> Acceptance tests create real resources, and often cost money to run.

For information about writing acceptance tests, see the main [Terraform contributing guide](https://www.terraform.io/docs/extend/testing/acceptance-tests/index.html).

### Adding Dependencies
This provider uses [Go modules](https://blog.golang.org/using-go-modules).

To add a new dependency to your Terraform provider:

```bash
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## License
[MIT License](LICENSE).

## Credits

The original version of the Aiven Terraform provider was written and maintained by [Jelmer Snoeck](https://github.com/jelmersnoeck).