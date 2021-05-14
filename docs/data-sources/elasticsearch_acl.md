# Elasticsearch ACL Data Source

The Elasticsearch ACL data source provides information about the existing Aiven Elasticsearch ACL 
for Elasticsearch service.

## Example Usage

```hcl
data "aiven_elasticsearch_acl" "es-acls" {
    project = aiven_project.es-project.project
    service_name = aiven_elasticsearch.es.service_name
}
```

## Argument Reference

* `project` and `service_name` - (Required) define the project and service the ACL belongs to. 
They should be defined using reference as shown above to set up dependencies correctly.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `enabled` - enables or disables Elasticsearch ACLs.

* `extended_acl` - Index rules can be applied in a limited fashion to the _mget, _msearch and _bulk APIs 
(and only those) by enabling the ExtendedAcl option for the service. When it is enabled, users can use 
 these APIs as long as all operations only target indexes they have been granted access to.
 
* `acl.username` - is the name of the existing service user, and service user must be preliminary added 
to the Elasticsearch service; this can be done using `aiven_service_user` resource. Aiven has a 
default user `avnadmin` which is automatically created as a part of the creation process of Elasticsearch service. 

Elasticsearch ACL support multiple rules for a single user.

* `acl.rule.index` - is the Elasticsearch index pattern.

* `acl.rule.permission` - is the Elasticsearch permission, list of supported permissions: 
`deny`, `admin`, `read`, `readwrite`, `write`.