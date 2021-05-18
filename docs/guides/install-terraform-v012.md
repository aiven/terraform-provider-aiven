---
page_title: "Using Terraform v0.12"
---

# Using Terraform v0.12
_If you can please upgrade your Terraform client to version 0.13 or above, in this case, there is no need to install the provider manually._

Download the latest Aiven provider for your platform from the [release page](https://github.com/aiven/terraform-provider-aiven/releases).

Third-party provider plugins — locally installed providers, not on the registry — need to be assigned a source and placed in the appropriate subdirectory for Terraform to find and use them. Create the appropriate subdirectory within the user plugins directory for the Aiven provider and move the downloaded binary there.

```bash
export AIVEN_PROVIDER_VERSION=2.X.X
export OS_ARCH=linux_adm64

mkdir -p ~/.terraform.d/plugins/aiven.io/provider/aiven/$AIVEN_PROVIDER_VERSION/$OS_ARCH
mv terraform-provider-aiven ~/.terraform.d/plugins/aiven.io/provider/aiven/$AIVEN_PROVIDER_VERSION/$OS_ARCH
```

Now you can use the provider in your Terraform configuration:
```hcl
terraform {
  required_providers {
    aiven = {
      versions = [
        "2.X.X"
      ]
      source = "aiven.io/provider/aiven"
    }
  }
}

provider "aiven" {
  api_token = var.aiven_api_token
}
```

Then, initialize your Terraform workspace by running `terraform init`. If your Aiven provider is located in the correct directory, it should successfully initialize.