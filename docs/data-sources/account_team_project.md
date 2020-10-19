# Account Team Project Data Source

The Account Team Project data source provides information about the existing Account Team Project.

## Example Usage

```hcl
data "aiven_account_team_project" "account_team_project1" {
    account_id = aiven_account.<ACCOUNT_RESOURCE>.account_id
    team_id = aiven_account_team.<TEAM_RESOURCE>.team_id
    project_name = aiven_project.<PROJECT>.project
}
```

## Argument Reference

* `account_id` - (Required) is a unique account id.

* `team_id` - (Required) is an account team id.

* `project_name` - (Required) is a project name of already existing project.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `team_type` - is an account team project type, can one of the following values: `admin`, 
`developer`, `operator` and `read_only`.