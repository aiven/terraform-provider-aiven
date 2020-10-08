---
title: Home
nav_exclude: true
---

# Terraform Aiven

[Terraform](https://www.terraform.io/) provider for [Aiven.io](https://aiven.io/). This
provider allows you to conveniently manage all resources for Aiven.

**A word of cautions**: While Terraform is an extremely powerful tool that can make
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

## Installation

```hcl-terraform
terraform {
  required_providers {
    aiven = {
      versions = [
        "2.X"
      ]
      source = "aiven.io/provider/aiven"
    }
  }
}

provider "aiven" {
  api_token = var.aiven_api_token
}
```
Then, initialize your Terraform workspace by running `terraform init`.

## Credits

The original version of the Aiven Terraform provider was written and maintained by
Jelmer Snoeck (https://github.com/jelmersnoeck).
