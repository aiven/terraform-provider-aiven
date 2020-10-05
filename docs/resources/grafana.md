# Grafana Resource

The Grafana resource allows the creation and management of an Aiven Grafana services.

## Example Usage

```hcl
resource "aiven_grafana" "gr1" {
    project = data.aiven_project.ps1.project
    cloud_name = "google-europe-west1"
    plan = "startup-1"
    service_name = "my-gr1"
    maintenance_window_dow = "monday"
    maintenance_window_time = "10:00:00"
    
    grafana_user_config {
        alerting_enabled = true
        
        public_access {
            grafana = true
        }
    }
}
```

## Argument Reference

* `project` - (Required) identifies the project the service belongs to. To set up proper dependency
between the project and the service, refer to the project as shown in the above example.
Project cannot be changed later without destroying and re-creating the service.

* `service_name` - (Required) specifies the actual name of the service. The name cannot be changed
later without destroying and re-creating the service so name should be picked based on
intended service usage rather than current attributes.

* `cloud_name` - (Optional) defines where the cloud provider and region where the service is hosted
in. This can be changed freely after service is created. Changing the value will trigger
a potentially lenghty migration process for the service. Format is cloud provider name
(`aws`, `azure`, `do` `google`, `upcloud`, etc.), dash, and the cloud provider
specific region name. These are documented on each Cloud provider's own support articles,
like [here for Google](https://cloud.google.com/compute/docs/regions-zones/) and
[here for AWS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html).

* `plan` - (Optional) defines what kind of computing resources are allocated for the service. It can
be changed after creation, though there are some restrictions when going to a smaller
plan such as the new plan must have sufficient amount of disk space to store all current
data and switching to a plan with fewer nodes might not be supported. The basic plan
names are `hobbyist`, `startup-x`, `business-x` and `premium-x` where `x` is
(roughly) the amount of memory on each node (also other attributes like number of CPUs
and amount of disk space varies but naming is based on memory). The exact options can be
seen from the Aiven web console's Create Service dialog.

* `project_vpc_id` - (Optional) optionally specifies the VPC the service should run in. If the value
is not set the service is not run inside a VPC. When set, the value should be given as a
reference as shown above to set up dependencies correctly and the VPC must be in the same
cloud and region as the service itself. Project can be freely moved to and from VPC after
creation but doing so triggers migration to new servers so the operation can take
significant amount of time to complete if the service has a lot of data.

* `termination_protection` - (Optional) prevents the service from being deleted. It is recommended to
set this to `true` for all production services to prevent unintentional service
deletions. This does not shield against deleting databases or topics but for services
with backups much of the content can at least be restored from backup in case accidental
deletion is done.

* `maintenance_window_dow` - (Optional) day of week when maintenance operations should be performed. 
One monday, tuesday, wednesday, etc.

* `maintenance_window_time` - (Optional) time of day when maintenance operations should be performed. 
UTC time in HH:mm:ss format.

* `grafana_user_config` - (Optional) defines Grafana specific additional configuration options. The following 
configuration options available:
    * `alerting_enabled` - (Optional) Enable or disable Grafana alerting functionality
    * `alerting_error_or_timeout` - (Optional) Default error or timeout setting for new alerting rules
    * `alerting_nodata_or_nullvalues` - (Optional) Default value for 'no data or null values' for
     new alerting rules
    * `allow_embedding` - (Optional) Allow embedding Grafana dashboards with iframe/frame/object/embed 
    tags. Disabled by default to limit impact of clickjacking
    * `auth_basic_enabled` - (Optional) Enable or disable basic authentication form, used by Grafana 
    built-in login.
     
    * `auth_generic_oauth` - (Optional) Generic OAuth integration.
        * `allow_sign_up` - (Optional) Automatically sign-up users on successful sign-in
        * `allowed_domains` - (Optional) Allowed domain
        * `allowed_organizations` - (Optional) Allowed organization
        * `api_url` - (Required) API URL
        * `auth_url` - (Required) Authorization URL
        * `client_id` - (Required) Client ID from provider
        * `client_secret` - (Required) Client secret from provider
        * `name` - (Optional) Name of the OAuth integration
        * `scopes` - (Optional) Scope must be non-empty string without whitespace
        * `token_url` - (Required) Token URL
    
    * `auth_github` - (Optional) Github Auth integration.
        * `auth_github` - (Optional) Automatically sign-up users on successful sign-in
        * `allowed_organizations` - (Optional) Must consist of alpha-numeric characters and dashes"
        * `client_id` - (Required) Client ID from provider
        * `client_secret` - (Required) Client secret from provider
        * `team_ids` - (Optional) Require users to belong to one of given team IDs
    
    * `auth_gitlab` - (Optional) GitLab Auth integration.
        * `allow_sign_up` - (Optional) Automatically sign-up users on successful sign-in
        * `allowed_groups` - (Required) Require users to belong to one of given groups
        * `api_url` - (Optional) API URL. This only needs to be set when using self hosted GitLab
        * `auth_url` - (Optional) Authorization URL. This only needs to be set when using self hosted GitLab
        * `client_id` - (Required) Client ID from provider
        * `client_secret` - (Required) Client secret from provider
        * `token_url` - (Optional) Token URL. This only needs to be set when using self hosted GitLab
    
    * `auth_google` - (Optional) Google Auth integration
        * `allow_sign_up` - (Optional) Automatically sign-up users on successful sign-in
        * `allowed_domain` - (Required) Domains allowed to sign-in to this Grafana
        * `client_id` - (Required) Client ID from provider
        * `client_secret` - (Required) Client secret from provider
    
    * `cookie_samesite` - (Optional) Cookie SameSite attribute: 'strict' prevents sending cookie for 
    cross-site requests, effectively disabling direct linking from other sites to Grafana. 'lax' is the default value.
    * `custom_domain` - (Optional) Serve the web frontend using a custom CNAME pointing to the Aiven DNS name.
    * `dashboards_versions_to_keep` - (Optional) Dashboard versions to keep per dashboard.
    * `dataproxy_send_user_header` - (Optional) Send 'X-Grafana-User' header to data source.
    * `dataproxy_timeout` - (Optional) Timeout for data proxy requests in seconds.
    * `disable_gravatar` - (Optional) Set to true to disable gravatar. Defaults to false 
    (gravatar is enabled).
    * `editors_can_admin` - (Optional) Editors can manage folders, teams and dashboards created by them.
    
    * `external_image_storage` - (Optional) External image store settings
        * `access_key` - (Required) S3 access key. Requires permissions to the S3 bucket for the 
        s3:PutObject and s3:PutObjectAcl actions
        * `bucket_url` - (Required) Bucket URL for S3
        * `provider` - (Required) Provider type
        * `secret_key` - (Required) S3 secret key
    
    * `google_analytics_ua_id` - (Optional) Google Analytics Universal Analytics ID for tracking Grafana usage
    * `ip_filter` - (Optional) Allow incoming connections from CIDR address block, e.g. '10.20.0.0/16'
    * `metrics_enabled` - (Optional) Enable Grafana /metrics endpoint

    * `privgrafanaate_access` - (Optional) Allow access to selected service ports from private networks.
        * `grafana` - (Optional) Allow clients to connect to grafana with a DNS name that always resolves to the 
        service's private IP addresses. Only available in certain network locations.
    
    * `public_access` - (Optional) Allow access to selected service ports from the public Internet.
        * `grafana` - (Optional) Allow clients to connect to grafana from the public internet for service nodes that 
        are in a project VPC or another type of private network.
        
    * `recovery_basebackup_name` - (Optional) Name of the basebackup to restore in forked service.
    * `service_to_fork_from` - (Optional) Name of another service to fork from. This has effect only 
    when a new service is being created.
    
    * `smtp_server` - (Optional) SMTP server settings.
        * `from_address` - (Required) Address used for sending emails
        * `from_name` - (Optional) Name used in outgoing emails, defaults to Grafana
        * `host` - (Required) Server hostname or IP
        * `password` - (Optional) Password for SMTP authentication
        * `port` - (Required) SMTP server port
        * `skip_verify` - (Optional) Skip verifying server certificate. Defaults to false
        * `username` - (Optional) Username for SMTP authentication
        
    * `user_auto_assign_org` - (Optional) Auto-assign new users on signup to main organization. 
    Defaults to false.
    * `user_auto_assign_org_role` - (Optional) Set role for new signups. Defaults to Viewer.
    * `viewers_can_edit` - (Optional) Users with view-only permission can edit but not save dashboards.
    
* `timeouts` - (Optional) a custom client timeouts.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `service_uri` - URI for connecting to the Grafana service.

* `service_host` - Grafana hostname.

* `service_port` - Grafana port.

* `service_password` - Password used for connecting to the Grafana service, if applicable.

* `service_username` - Username used for connecting to the Grafana service, if applicable.

* `state` - Service state.

* `grafana` - Grafana specific server provided values.
