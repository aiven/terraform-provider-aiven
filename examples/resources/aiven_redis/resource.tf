resource "aiven_redis" "redis1" {
    project = data.aiven_project.pr1.project
    cloud_name = "google-europe-west1"
    plan = "business-4"
    service_name = "my-redis1"
    maintenance_window_dow = "monday"
    maintenance_window_time = "10:00:00"
    
    redis_user_config {
        redis_maxmemory_policy = "allkeys-random"		
        
        public_access {
            redis = true
        }
    }
}
