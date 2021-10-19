data "aiven_grafana" "gr1" {
    project = data.aiven_project.ps1.project
    service_name = "my-gr1"
}

