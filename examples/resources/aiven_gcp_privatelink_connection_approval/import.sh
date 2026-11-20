terraform import aiven_gcp_privatelink_connection_approval.example PROJECT/SERVICE_NAME
# For a service with multiple PSC connections, run this instead:
# terraform import aiven_gcp_privatelink_connection_approval.example PROJECT/SERVICE_NAME/PSC_CONNECTION_ID

# Identity-based import requires exactly one PSC connection for the service.
# For multiple connections, use the PSC-qualified string import above.
