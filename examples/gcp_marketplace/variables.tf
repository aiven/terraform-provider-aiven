# Find your account name from the Aiven GCP Cloud console https://console.gcp.aiven.io
variable "aiven_account_name" {
  type = string
}

# Create a token at https://console.gcp.aiven.io/profile/auth
# Note that this will be different from any tokens you have for https://console.aiven.io
variable "aiven_api_token" {
  type = string
}
