# Find an id from an existing project in the Timescale Cloud console
variable "aiven_project_id" {
  type = string
}

# Create a token at https://portal.timescale.cloud/profile/auth
variable "timescale_api_token" {
  type = string
}
