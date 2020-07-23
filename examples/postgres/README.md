# Setup

## Variables

Rename `./secrets.tfvars.tmp` to `./secrets.tfvars` and fill in the appropriate values.

## Initialize Terraform

Ensure that you have Terraform v0.12.\* installed and initialize the project.

```sh
$ bin/setup

Terraform v0.12.24
+ provider.aiven (unversioned)

Your version of Terraform is out of date! The latest version
is 0.12.28. You can update by downloading from https://www.terraform.io/downloads.html

Initializing the backend...

Initializing provider plugins...

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.
```

# Apply

Deploy your changes

```sh
$ bin/apply
...
Plan: 1 to add, 0 to change, 0 to destroy.

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes
...
aiven_service.avn-pg: Still creating... [6m1s elapsed]
aiven_service.avn-pg: Creation complete after 6m2s [id=david-tf-demo/postgres-eu]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
```

# Cleanup

```sh
$ bin/destroy
...
Plan: 0 to add, 0 to change, 1 to destroy.

Do you really want to destroy all resources?
  Terraform will destroy all your managed infrastructure, as shown above.
  There is no undo. Only 'yes' will be accepted to confirm.

  Enter a value: yes

aiven_service.avn-pg: Destroying... [id=david-tf-demo/postgres-eu]
aiven_service.avn-pg: Destruction complete after 0s

Destroy complete! Resources: 1 destroyed.
```
