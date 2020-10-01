# Elasticsearch ACL Resource

The Elasticsearch ACL resource allows the creation and management of an Aiven Elasticsearch ACLs 
for Elasticsearch service.

## Example Usage

```hcl
resource "aiven_elasticsearch_acl" "es-acls" {
    project = aiven_project.es-project.project
    service_name = aiven_service.es.service_name
    enabled = true
    extended_acl = false
    acl {
        username = aiven_service_user.es-user.username
        rule {
            index = "_*"
            permission = "admin"
        }
    
        rule {
            index = "*"
            permission = "admin"
        }
    }
        
    acl {
        username = "avnadmin"
        rule {
            index = "_*"
            permission = "read"
        }
        
        rule {
            index = "*"
            permission = "read"
        }
    }
    }
```

## Argument Reference

* `project` and `service_name` - (Required) define the project and service the ACL belongs to. 
They should be defined using reference as shown above to set up dependencies correctly.

All other properties except `project` and `service_name` can be changed after creation of the 
resource and will not trigger recreation of Elasticsearch entire ACL's. 

* `enabled` - (Optional) enables of disables Elasticsearch ACL's.

* `extended_acl` - (Optional) Index rules can be applied in a limited fashion to the _mget, _msearch and _bulk APIs 
(and only those) by enabling the ExtendedAcl option for the service. When it is enabled, users can use 
 these APIs as long as all operations only target indexes they have been granted access to.
 
* `acl.username` - (Optional) is the name of the existing service user, and service user must be preliminary added 
to the Elasticsearch service; this can be done using `aiven_service_user` resource. Aiven has a 
default user `avnadmin` which is automatically created as a part of the creation process of Elasticsearch service. 

Elasticsearch ACL support multiple rules for a single user.

* `acl.rule.index` - (Optional) is the Elasticsearch index pattern.

* `acl.rule.permission` - (Optional) is the Elasticsearch permission, list of supported permissions: 
`deny`, `admin`, `read`, `readwrite`, `write`.