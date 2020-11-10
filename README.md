---
title: Home
nav_exclude: true
---

# Terraform Aiven

[Terraform](https://www.terraform.io/) provider for [Aiven.io](https://aiven.io/). This
provider allows you to conveniently manage all resources for Aiven.

**A word of caution**: While Terraform is an extremely powerful tool that can make
managing your infrastructure a breeze, great care must be taken when comparing the
changes that are about to applied to your infrastructure. When it comes to stateful
services, you cannot just re-create a resource and have it in the original state;
deleting a service deletes all data associated with it and there's often no way to
recover the data later. Whenever the Terraform plan indicates that a service, database,
topic or other such central construct is about to be deleted, something catastrophic is
quite possibly about to happen unless you're dealing with some throwaway test
environments or are deliberately retiring the service/database/topic.

There are many properties that cannot be changed after a resource is created and changing
those values later is handled by deleting the original resource and creating a new one.
These properties include such as the project a service is associated with, the name of a
service, etc. Unless the system contains no relevant data, such changes must not be
performed.

To allow mitigating this problem, the service resource supports
`termination_protection` property. It is recommended to set this property to `true`
for all production services to avoid them being accidentally deleted. With this setting
enabled service deletion, both intentional and unintentional, will fail until an explicit
update is done to change the setting to `false`. Note that while this does prevent the
service itself from being deleted, any databases, topics or such that have been configured
with Terraform can still be deleted and they will be deleted before the service itself is
attempted to be deleted so even with this setting enabled you need to be very careful
with the changes that are to be applied.

## Requirements
- [Terraform](https://www.terraform.io/downloads.html) v0.12.X or greater
- [Go](https://golang.org/doc/install) 1.14.X or greater

## Using the provider
See the [Aiven Provider documentation](https://registry.terraform.io/providers/aiven/aiven/latest/docs) to get started using the Aiven provider.

## Developing the Provider
If you wish to work on the provider, you'll first need Go installed on your machine (version 1.14+ is required).

To compile the provider, run `make release`. This will build the provider and put the provider binary in the `plugins/$OS_$ARCH` directory.

In order to test the provider, you can simply run make test.

```shell script
$ make test
```
In order to run the full suite of acceptance tests, run `make testacc`.

**Required environment variables for acceptance tests:**
- `AIVEN_TOKEN` - Aiven Token.
- `AIVEN_PROJECT_NAME` - project with enough credits where all tests will be executed.

*Note: Acceptance tests create real resources, and often cost money to run.*

```shell script
$ make testacc
```

In order to run a specific acceptance test, use the TESTARGS environment variable. For example, 
the following command will run `TestAccAiven_kafka` acceptance test only:

```shell script
$ make testacc TESTARGS='-run=TestAccAiven_kafka'
```

For information about writing acceptance tests, see the main [Terraform contributing guide](https://github.com/hashicorp/terraform/blob/master/.github/CONTRIBUTING.md#writing-acceptance-tests).

## Credits

The original version of the Aiven Terraform provider was written and maintained by
Jelmer Snoeck (https://github.com/jelmersnoeck).
