data "aiven_valkey_user" "example" {
  project      = "my-project"
  service_name = "my-valkey"
  username     = "testuser"

  /* COMPUTED FIELDS
  password                 = "password123"
  password_encryption_type = "md5"
  type                     = "foo"
  valkey_acl_categories    = ["+@write", "+@keyspace"]
  valkey_acl_channels      = ["some*chan"]
  valkey_acl_commands      = ["+set", "+del", "+expire", "-flushall", "-flushdb"]
  valkey_acl_keys          = ["session:*"]
  */
}
