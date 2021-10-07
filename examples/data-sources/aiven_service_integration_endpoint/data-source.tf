data "aiven_service_integration_endpoint" "myendpoint" {
    project = aiven_project.myproject.project
    endpoint_name = "<ENDPOINT_NAME>"
}

