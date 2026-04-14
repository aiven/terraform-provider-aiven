resource "aiven_organization_application_user_token" "example" {
  organization_id = "org1a23f456789" // Force new
  user_id         = "foo" // Force new

  // OPTIONAL FIELDS
  description      = "Integration client Alpha" // Force new
  extend_when_used = false // Force new
  ip_allowlist     = ["192.168.0.0/24"] // Force new
  max_age_seconds  = 42 // Force new
  scopes           = ["user:read"] // Force new

  /* COMPUTED FIELDS
  create_time                    = "2021-01-01T00:00:00Z"
  created_manually               = true
  currently_active               = true
  expiry_time                    = "2021-01-01T00:00:00Z"
  full_token                     = "foo"
  last_ip                        = "192.168.1.1"
  last_used_time                 = "2021-01-01T00:00:00Z"
  last_user_agent                = "foo"
  last_user_agent_human_readable = "foo"
  token_prefix                   = "foo"
  */
}
