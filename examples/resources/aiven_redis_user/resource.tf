resource "aiven_redis_user" "foo" {
  service_name = aiven_redis.bar.service_name
  project      = "my-project"
  username     = "user-1"
  password     = "Test$1234"
}
