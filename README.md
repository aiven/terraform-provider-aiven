# Aiven Terraform Provider
The Terraform provider for [Aiven.io](https://aiven.io/), an open source data platform as a service. 

**See the [official documentation](https://registry.terraform.io/providers/aiven/aiven/latest/docs) to learn about all the possible services and resources.**

## Quick Start
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

- Run in your terminal:
```bash
$ terraform init
$ terraform plan
$ terraform apply
$ psql "$(terraform output -raw postgresql_service_uri)"
```

Voilà, a PostgreSQL database.

## A word of caution
Recreating stateful services with Terraform will possibly **delete** the service and all its data before creating it again. Whenever the Terraform plan indicates that a service will be **deleted** or **replaced**, a catastrophic action is possibly about to happen.

Some properties, like **project** and the **resource name**, cannot be changed and it will trigger a resource replacement.

To avoid any issues, **please set the `termination_protection` property to `true` on all production services**, it will prevent Terraform to remove the service until the flag is set back to `false` again. While it prevents a service to be deleted, any logical databases, topics or other configurations may be removed **even when this section is enabled**. Be very careful! 

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

### Alternate command to build/test inside a container

In case you'd like to avoid Go-dependency on your local machine or CI, you can run the following build command:

```
podman run -v .:/terraform-provider-aiven:z --workdir /terraform-provider-aiven golang:latest make <COMMAND>
```

Note that you'll need to be in the root of **terraform-provider-aiven** repository and replace the <COMMAND> name with your choice of command (for example, ``make docs``).

## Debugging
### Requirements
- [Terraform](https://www.terraform.io/downloads.html) v0.12.26 or greater
- [Delve](https://github.com/go-delve/delve/tree/master/Documentation/installation) 1.7.X or greater

### Starting the Provider Plugin with `dlv`

To start debugging the provider we use the `dlv debug` command and pass the `-debug` flag to the compiled binary:

```bash
$ dlv debug -- -debug
Type 'help' for list of commands.
(dlv)
```

Next, we set a breakpoint at a function that we want to look into and continue to start the plugin:

```bash
(dlv) break resourceServiceCreate
Breakpoint 1 set at 0x129667b for github.com/aiven/terraform-provider-aiven/aiven.resourceServiceCreate() ./aiven/resource_service.go:806
(dlv) c
{"@level":"debug","@message":"plugin address","@timestamp":"2021-10-12T16:36:45.528158+02:00","address":"/tmp/plugin252726151","network":"unix"}
Provider started, to attach Terraform set the TF_REATTACH_PROVIDERS env var:
    TF_REATTACH_PROVIDERS='{"registry.terraform.io/aiven/aiven":{"Protocol":"grpc","ProtocolVersion":5,"Pid":3153652,"Test":true,"Addr":{"Network":"unix","String":"/tmp/plugin252726151"}}}'
```

Next, in a different terminal session we start terraform using the `TF_REATTACH_PROVIDERS` environment variable:

```bash
$ export TF_REATTACH_PROVIDERS='{"registry.terraform.io/aiven/aiven":{"Protocol":"grpc","ProtocolVersion":5,"Pid":3153652,"Test":true,"Addr":{"Network":"unix","String":"/tmp/plugin252726151"}}}'

$ terraform init -upgrade
<...>
$ terraform plan
<...>
$ terraform apply -parallelism=1
<...>
aiven_pg.testing: Creating...
```

Now we can see that the debugged process did hit the breakpoint we specified earlier:

```bash
> github.com/aiven/terraform-provider-aiven/aiven.resourceServiceCreate() ./aiven/resource_service.go:806 (hits goroutine(326):1 total:1) (PC: 0x129667b)
   801:			return resourceServiceCreate(ctx, d, m)
   802:		}
   803:	
   804:	}
   805:	
=> 806:	func resourceServiceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
   807:		client := m.(*aiven.Client)
   808:		serviceType := d.Get("service_type").(string)
   809:		userConfig := ConvertTerraformUserConfigToAPICompatibleFormat("service", serviceType, true, d)
   810:		vpcID := d.Get("project_vpc_id").(string)
   811:		var apiServiceIntegrations []aiven.NewServiceIntegration
```


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
