---
page_title: "Troubleshooting"
---

# Troubleshooting

## Log the API requests and responses

To see what the provider sends to the Aiven API and what it gets back, set `TF_LOG_PROVIDER_AIVEN_HTTP` to `DEBUG`
along with the Terraform provider logs:

```sh
TF_LOG_PROVIDER=DEBUG TF_LOG_PROVIDER_AIVEN_HTTP=DEBUG TF_LOG_PATH=terraform.log terraform apply
```

Each API call adds a `Sending API request` and a `Received API response` entry to the log, in the `provider.http`
module. They show the operation ID, method, URL, status code, duration and body. `TF_LOG=DEBUG` alone doesn't log them.

The provider masks the secrets in these logs:

* Values of secret fields, such as `password`, `access_key` or `token`, are replaced with `***`.
* Connection URIs keep the user but not the password, for example, `postgres://avnadmin:xxxxx@host:12691/defaultdb`.
* PEM private keys and JWTs are masked under any field.
* Request and response headers, including the API token, aren't logged.

The masking relies on field names and value patterns, so check the log for secrets before you share it, for example,
in a GitHub issue.

## A create times out but the service exists

Creating a service, especially a fork or a restore from a backup, can take longer than the default create timeout
of 20 minutes. Raise it with a `timeouts` block:

```hcl
resource "aiven_pg" "example" {
  # ...

  timeouts {
    create = "60m"
  }
}
```

A service that times out isn't added to the Terraform state, so the next apply fails to create it again. If
`terraform state list` doesn't show it, wait until the service is running and [import it](importing-resources.md).
