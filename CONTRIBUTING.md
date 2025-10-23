# Welcome!

Contributions are very welcome on terraform-provider-aiven. When contributing please keep this in mind:

- Open an issue to discuss new bigger features.
- Write code consistent with the project style and make sure the tests are passing.
- Stay in touch with us if we have follow up questions or requests for further changes.

## Development

### Local environment

[Go](https://go.dev/doc/install) >=1.25 \
[Terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli) >=1.2 \
[Task](https://taskfile.dev/#/installation) >=3.20.0

#### Alternative command to build/test inside a container

In case you would like to avoid Go dependency on your local machine or CI, you can use the following build command:

```bash
podman run -v .:/terraform-provider-aiven:z --workdir /terraform-provider-aiven golang:latest <command>
```

> _**N.B.** You need to be in the root of terraform-provider-aiven repository and replace the _&lt;command&gt;_ placeholder with your desired task command, e.g. `task ci:docs`._

### Tests

Run the tests with the command below:

```bash
task test
```

Acceptance tests interact with the real Aiven API and require an API token and existing project and organization on Aiven to create resources in. A non-default API URL can be used by setting the `AIVEN_WEB_URL` environment variable to e.g. `https://your.custom.api:443`.

Run the acceptance tests with the commands below:

```bash
export AIVEN_TOKEN="your-token"
export AIVEN_PROJECT_NAME="your-project-name"
export AIVEN_ORGANIZATION_NAME="your-existing-org-name"
export AIVEN_ACCOUNT_NAME="$AIVEN_ORGANIZATION_NAME"

# run all acceptance tests
task test-acc

# or run a specific acceptance test
task test-acc -- -run=TestAccAivenOrganizationUserDataSource_using_email
```

> _**N.B.** Acceptance tests create real resources, and often cost money to run._

For information about writing acceptance tests, see the main [Terraform contributing guide](https://www.terraform.io/docs/extend/testing/acceptance-tests/index.html).

#### Testing examples

Run example tests with the following commands:

```bash
export AIVEN_TOKEN="your-token"
export AIVEN_PROJECT_NAME="your-project-name"

task test-examples
```

### Linting and formatting

Before pushing your changes, run the following commands to ensure code quality:

```bash
task lint
task fmt
```

### Using a dev build

To test local changes during development, you can create and use a development build of the provider.

1. Build the provider and show the installation path by running:

```bash
task build-dev
task: [build-dev] mkdir -p /Users/<USERNAME>/.terraform.d/plugins/registry.terraform.io/aiven-dev/aiven/0.0.0+dev/darwin_arm64
```

2. Create the `~/.terraformrc` file with the path using the output from the previous step:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/aiven-dev/aiven" = "/Users/<USERNAME>/.terraform.d/plugins/registry.terraform.io/aiven-dev/aiven/0.0.0+dev/darwin_arm64/"
  }

  direct {}
}
```

3. Update your Terraform configuration to use the development provider by changing the source to `aiven-dev/aiven`:

```hcl
terraform {
  required_providers {
    aiven = {
      source = "aiven-dev/aiven"
    }
  }
}

resource "aiven_pg" "example" {
  // ...
}
```

4. To test your changes directly, run the following command. You can skip the `terraform init` step.

```bash
terraform plan
```

After making code changes, rebuild by running `task build-dev`.

For more information, see [Development Overrides for Provider Developers](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers).


### Using a custom Aiven API endpoint

To use a custom Aiven API endpoint, set the `AIVEN_WEB_URL` environment variable:

```bash
export AIVEN_WEB_URL="your-custom-url"
```

### Debugging

To start debugging the provider we use the `dlv debug` command and pass the `-debug` flag to the compiled binary:

```bash
dlv debug -- -debug
Type 'help' for list of commands.
(dlv)
```

Next, we set a breakpoint at a function that we want to look into and continue to start the plugin:

```log
(dlv) break resourceServiceCreate
Breakpoint 1 set at 0x129667b for github.com/aiven/terraform-provider-aiven/aiven.resourceServiceCreate() ./aiven/resource_service.go:806
(dlv) c
{"@level":"debug","@message":"plugin address","@timestamp":"2021-10-12T16:36:45.528158+02:00","address":"/tmp/plugin252726151","network":"unix"}
Provider started, to attach Terraform set the TF_REATTACH_PROVIDERS env var:
    TF_REATTACH_PROVIDERS='{"registry.terraform.io/aiven/aiven":{"Protocol":"grpc","ProtocolVersion":5,"Pid":3153652,"Test":true,"Addr":{"Network":"unix","String":"/tmp/plugin252726151"}}}'
```

Next, in a different terminal session we start terraform using the `TF_REATTACH_PROVIDERS` environment variable and run some commands:

```bash
export TF_REATTACH_PROVIDERS='{"registry.terraform.io/aiven/aiven":{"Protocol":"grpc","ProtocolVersion":5,"Pid":3153652,"Test":true,"Addr":{"Network":"unix","String":"/tmp/plugin252726151"}}}'
terraform init -upgrade
terraform plan
terraform apply -parallelism=1
```

Now we can see that the debugged process did hit the breakpoint we specified earlier:

```log
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

## Opening a PR

- Commit messages should describe the changes, not the filenames. Win our admiration by following the [excellent advice from Chris Beams](https://chris.beams.io/posts/git-commit/) when composing commit messages.
- Choose a meaningful title for your pull request.
- The pull request description should focus on what changed and why.
- Check that the tests pass (and add test coverage for your changes if appropriate).

### Commit Messages

This project adheres to the [Conventional Commits](https://conventionalcommits.org/en/v1.0.0/) specification.
Please, make sure that your commit messages follow that specification.
