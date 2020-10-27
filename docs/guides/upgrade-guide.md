---
parent: Guides
page_title: "Upgrade Guide"
---

# Upgrade Guide

## From 1.2.4

If you have specified `-1` as a placeholder for unset values in user config, you will find a diff in Terraform configuration after upgrading. Even if you apply the Terraform plan, these will not disappear.

The best option is to remove these placeholder values completely from the Terraform config file.

```diff
 resource "aiven_service" "kafka" {
   ...
   kafka_user_config {
     ...
     kafka {
-      default_replication_factor = -1
       ...
     }
   }
 }
```

If you really need to keep the explicit placeholder values (e.g. because you have a default reference config used as a base to dynamically populate other resource configurations) then the solution is to convert the `-1` unset values to `""` as shown below.

```diff
 locals {
   kafka_user_config_default {
-    default_replication_factor = -1
+    default_replication_factor = ""
     ...
   }
 }
```
