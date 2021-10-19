data "aiven_connection_pool" "mytestpool" {
    project = aiven_project.myproject.project
    service_name = aiven_service.myservice.service_name
    pool_name = "mypool"
}

