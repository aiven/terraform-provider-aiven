# Account Team Project Resource

The Account Team Project resource allows the creation and management of an Account Team Project.

It is intended to link an existing project to the existing account team. 
It is important to note that the project should have an `account_id` property set equal to the
account team you are trying to link to this project. 

## Example Usage

```hcl
resource "aiven_project" "<PROJECT>" {
  project = "project-1"
  account_id = aiven_account_team.<ACCOUNT_RESOURCE>.account_id
}

resource "aiven_account_team_project" "account_team_project1" {
    account_id = aiven_account.<ACCOUNT_RESOURCE>.account_id
    team_id = aiven_account_team.<TEAM_RESOURCE>.team_id
    project_name = aiven_project.<PROJECT>.project
    team_type = "admin"
}
```

## Argument Reference

* `account_id` - (Required) is a unique account id.

* `team_id` - (Required) is an account team id.

* `project_name` - (Optional) is a project name of already existing project.

* `team_type` - (Optional) is an account team project type, can one of the following values: `admin`, 
`developer`, `operator` and `read_only`.