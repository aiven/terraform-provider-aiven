resource "aiven_valkey_user" "example" {
  project      = "my-project" // Force new
  service_name = "my-valkey" // Force new
  username     = "testuser" // Force new

  // OPTIONAL FIELDS
  password_wo           = "password123"
  password_wo_version   = 1
  valkey_acl_categories = ["+@write", "+@keyspace"]
  valkey_acl_channels   = ["some*chan"]
  valkey_acl_commands   = ["+set", "+del", "+expire", "-flushall", "-flushdb"]
  valkey_acl_keys       = ["session:*"]

  /* COMPUTED FIELDS
  password_encryption_type = "md5"
  type                     = "foo"
  */
}
