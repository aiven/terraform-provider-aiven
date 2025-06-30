resource "aiven_kafka" "{{ required .resource_name }}" {
  {{- if .additional_disk_space }}
  additional_disk_space = {{ renderValue .additional_disk_space }}
  {{- end }}
  {{- if .cloud_name }}
  cloud_name = {{ renderValue .cloud_name }}
  {{- end }}
  {{- if ne .default_acl nil }}
  default_acl = {{ .default_acl }}
  {{- end }}
  {{- if .disk_space }}
  disk_space = {{ renderValue .disk_space }}
  {{- end }}
  {{- if .kafka }}
  kafka {
    {{- if (index .kafka 0 "uris") }}
    uris = [
      {{- range $idx, $item := (index .kafka 0 "uris") }}
      {{ renderValue $item }},
      {{- end }}
    ]
    {{- end }}
  }
  {{- end }}
  {{- if .kafka_user_config }}
  kafka_user_config {
    {{- if (index .kafka_user_config 0 "additional_backup_regions") }}
    additional_backup_regions = [
      {{- range $idx, $item := (index .kafka_user_config 0 "additional_backup_regions") }}
      {{ renderValue $item }},
      {{- end }}
    ]
    {{- end }}
    {{- if ne (index .kafka_user_config 0 "aiven_kafka_topic_messages") nil }}
    aiven_kafka_topic_messages = {{ (index .kafka_user_config 0 "aiven_kafka_topic_messages") }}
    {{- end }}
    {{- if (index .kafka_user_config 0 "custom_domain") }}
    custom_domain = {{ renderValue (index .kafka_user_config 0 "custom_domain") }}
    {{- end }}
    {{- if (index .kafka_user_config 0 "follower_fetching") }}
    follower_fetching {
      {{- if ne (index (index .kafka_user_config 0 "follower_fetching") 0 "enabled") nil }}
      enabled = {{ (index (index .kafka_user_config 0 "follower_fetching") 0 "enabled") }}
      {{- end }}
    }
    {{- end }}
    {{- if (index .kafka_user_config 0 "ip_filter") }}
    ip_filter = [
      {{- range $idx, $item := (index .kafka_user_config 0 "ip_filter") }}
      {{ renderValue $item }},
      {{- end }}
    ]
    {{- end }}
    {{- if (index .kafka_user_config 0 "ip_filter_object") }}
    ip_filter_object {
      {{- if (index (index .kafka_user_config 0 "ip_filter_object") 0 "description") }}
      description = {{ renderValue (index (index .kafka_user_config 0 "ip_filter_object") 0 "description") }}
      {{- end }}
      network = {{ renderValue (required (index (index .kafka_user_config 0 "ip_filter_object") 0 "network")) }}
    }
    {{- end }}
    {{- if (index .kafka_user_config 0 "ip_filter_string") }}
    ip_filter_string = [
      {{- range $idx, $item := (index .kafka_user_config 0 "ip_filter_string") }}
      {{ renderValue $item }},
      {{- end }}
    ]
    {{- end }}
    {{- if (index .kafka_user_config 0 "kafka") }}
    kafka {
      {{- if ne (index (index .kafka_user_config 0 "kafka") 0 "auto_create_topics_enable") nil }}
      auto_create_topics_enable = {{ (index (index .kafka_user_config 0 "kafka") 0 "auto_create_topics_enable") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "compression_type") }}
      compression_type = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "compression_type") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "connections_max_idle_ms") }}
      connections_max_idle_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "connections_max_idle_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "default_replication_factor") }}
      default_replication_factor = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "default_replication_factor") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "group_initial_rebalance_delay_ms") }}
      group_initial_rebalance_delay_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "group_initial_rebalance_delay_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "group_max_session_timeout_ms") }}
      group_max_session_timeout_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "group_max_session_timeout_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "group_min_session_timeout_ms") }}
      group_min_session_timeout_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "group_min_session_timeout_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_cleaner_delete_retention_ms") }}
      log_cleaner_delete_retention_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_cleaner_delete_retention_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_cleaner_max_compaction_lag_ms") }}
      log_cleaner_max_compaction_lag_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_cleaner_max_compaction_lag_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_cleaner_min_cleanable_ratio") }}
      log_cleaner_min_cleanable_ratio = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_cleaner_min_cleanable_ratio") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_cleaner_min_compaction_lag_ms") }}
      log_cleaner_min_compaction_lag_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_cleaner_min_compaction_lag_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_cleanup_policy") }}
      log_cleanup_policy = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_cleanup_policy") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_flush_interval_messages") }}
      log_flush_interval_messages = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_flush_interval_messages") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_flush_interval_ms") }}
      log_flush_interval_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_flush_interval_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_index_interval_bytes") }}
      log_index_interval_bytes = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_index_interval_bytes") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_index_size_max_bytes") }}
      log_index_size_max_bytes = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_index_size_max_bytes") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_local_retention_bytes") }}
      log_local_retention_bytes = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_local_retention_bytes") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_local_retention_ms") }}
      log_local_retention_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_local_retention_ms") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "kafka") 0 "log_message_downconversion_enable") nil }}
      log_message_downconversion_enable = {{ (index (index .kafka_user_config 0 "kafka") 0 "log_message_downconversion_enable") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_message_timestamp_difference_max_ms") }}
      log_message_timestamp_difference_max_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_message_timestamp_difference_max_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_message_timestamp_type") }}
      log_message_timestamp_type = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_message_timestamp_type") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "kafka") 0 "log_preallocate") nil }}
      log_preallocate = {{ (index (index .kafka_user_config 0 "kafka") 0 "log_preallocate") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_retention_bytes") }}
      log_retention_bytes = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_retention_bytes") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_retention_hours") }}
      log_retention_hours = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_retention_hours") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_retention_ms") }}
      log_retention_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_retention_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_roll_jitter_ms") }}
      log_roll_jitter_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_roll_jitter_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_roll_ms") }}
      log_roll_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_roll_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_segment_bytes") }}
      log_segment_bytes = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_segment_bytes") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "log_segment_delete_delay_ms") }}
      log_segment_delete_delay_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "log_segment_delete_delay_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "max_connections_per_ip") }}
      max_connections_per_ip = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "max_connections_per_ip") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "max_incremental_fetch_session_cache_slots") }}
      max_incremental_fetch_session_cache_slots = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "max_incremental_fetch_session_cache_slots") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "message_max_bytes") }}
      message_max_bytes = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "message_max_bytes") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "min_insync_replicas") }}
      min_insync_replicas = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "min_insync_replicas") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "num_partitions") }}
      num_partitions = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "num_partitions") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "offsets_retention_minutes") }}
      offsets_retention_minutes = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "offsets_retention_minutes") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "producer_purgatory_purge_interval_requests") }}
      producer_purgatory_purge_interval_requests = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "producer_purgatory_purge_interval_requests") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "replica_fetch_max_bytes") }}
      replica_fetch_max_bytes = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "replica_fetch_max_bytes") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "replica_fetch_response_max_bytes") }}
      replica_fetch_response_max_bytes = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "replica_fetch_response_max_bytes") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "sasl_oauthbearer_expected_audience") }}
      sasl_oauthbearer_expected_audience = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "sasl_oauthbearer_expected_audience") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "sasl_oauthbearer_expected_issuer") }}
      sasl_oauthbearer_expected_issuer = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "sasl_oauthbearer_expected_issuer") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "sasl_oauthbearer_jwks_endpoint_url") }}
      sasl_oauthbearer_jwks_endpoint_url = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "sasl_oauthbearer_jwks_endpoint_url") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "sasl_oauthbearer_sub_claim_name") }}
      sasl_oauthbearer_sub_claim_name = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "sasl_oauthbearer_sub_claim_name") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "socket_request_max_bytes") }}
      socket_request_max_bytes = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "socket_request_max_bytes") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "kafka") 0 "transaction_partition_verification_enable") nil }}
      transaction_partition_verification_enable = {{ (index (index .kafka_user_config 0 "kafka") 0 "transaction_partition_verification_enable") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "transaction_remove_expired_transaction_cleanup_interval_ms") }}
      transaction_remove_expired_transaction_cleanup_interval_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "transaction_remove_expired_transaction_cleanup_interval_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka") 0 "transaction_state_log_segment_bytes") }}
      transaction_state_log_segment_bytes = {{ renderValue (index (index .kafka_user_config 0 "kafka") 0 "transaction_state_log_segment_bytes") }}
      {{- end }}
    }
    {{- end }}
    {{- if (index .kafka_user_config 0 "kafka_authentication_methods") }}
    kafka_authentication_methods {
      {{- if ne (index (index .kafka_user_config 0 "kafka_authentication_methods") 0 "certificate") nil }}
      certificate = {{ (index (index .kafka_user_config 0 "kafka_authentication_methods") 0 "certificate") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "kafka_authentication_methods") 0 "sasl") nil }}
      sasl = {{ (index (index .kafka_user_config 0 "kafka_authentication_methods") 0 "sasl") }}
      {{- end }}
    }
    {{- end }}
    {{- if ne (index .kafka_user_config 0 "kafka_connect") nil }}
    kafka_connect = {{ (index .kafka_user_config 0 "kafka_connect") }}
    {{- end }}
    {{- if (index .kafka_user_config 0 "kafka_connect_config") }}
    kafka_connect_config {
      {{- if (index (index .kafka_user_config 0 "kafka_connect_config") 0 "connector_client_config_override_policy") }}
      connector_client_config_override_policy = {{ renderValue (index (index .kafka_user_config 0 "kafka_connect_config") 0 "connector_client_config_override_policy") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_connect_config") 0 "consumer_auto_offset_reset") }}
      consumer_auto_offset_reset = {{ renderValue (index (index .kafka_user_config 0 "kafka_connect_config") 0 "consumer_auto_offset_reset") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_connect_config") 0 "consumer_fetch_max_bytes") }}
      consumer_fetch_max_bytes = {{ renderValue (index (index .kafka_user_config 0 "kafka_connect_config") 0 "consumer_fetch_max_bytes") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_connect_config") 0 "consumer_isolation_level") }}
      consumer_isolation_level = {{ renderValue (index (index .kafka_user_config 0 "kafka_connect_config") 0 "consumer_isolation_level") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_connect_config") 0 "consumer_max_partition_fetch_bytes") }}
      consumer_max_partition_fetch_bytes = {{ renderValue (index (index .kafka_user_config 0 "kafka_connect_config") 0 "consumer_max_partition_fetch_bytes") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_connect_config") 0 "consumer_max_poll_interval_ms") }}
      consumer_max_poll_interval_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka_connect_config") 0 "consumer_max_poll_interval_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_connect_config") 0 "consumer_max_poll_records") }}
      consumer_max_poll_records = {{ renderValue (index (index .kafka_user_config 0 "kafka_connect_config") 0 "consumer_max_poll_records") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_connect_config") 0 "offset_flush_interval_ms") }}
      offset_flush_interval_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka_connect_config") 0 "offset_flush_interval_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_connect_config") 0 "offset_flush_timeout_ms") }}
      offset_flush_timeout_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka_connect_config") 0 "offset_flush_timeout_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_connect_config") 0 "producer_batch_size") }}
      producer_batch_size = {{ renderValue (index (index .kafka_user_config 0 "kafka_connect_config") 0 "producer_batch_size") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_connect_config") 0 "producer_buffer_memory") }}
      producer_buffer_memory = {{ renderValue (index (index .kafka_user_config 0 "kafka_connect_config") 0 "producer_buffer_memory") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_connect_config") 0 "producer_compression_type") }}
      producer_compression_type = {{ renderValue (index (index .kafka_user_config 0 "kafka_connect_config") 0 "producer_compression_type") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_connect_config") 0 "producer_linger_ms") }}
      producer_linger_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka_connect_config") 0 "producer_linger_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_connect_config") 0 "producer_max_request_size") }}
      producer_max_request_size = {{ renderValue (index (index .kafka_user_config 0 "kafka_connect_config") 0 "producer_max_request_size") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_connect_config") 0 "scheduled_rebalance_max_delay_ms") }}
      scheduled_rebalance_max_delay_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka_connect_config") 0 "scheduled_rebalance_max_delay_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_connect_config") 0 "session_timeout_ms") }}
      session_timeout_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka_connect_config") 0 "session_timeout_ms") }}
      {{- end }}
    }
    {{- end }}
    {{- if (index .kafka_user_config 0 "kafka_connect_secret_providers") }}
    kafka_connect_secret_providers {
      {{- if (index (index .kafka_user_config 0 "kafka_connect_secret_providers") 0 "aws") }}
      aws {
        {{- if (index (index (index .kafka_user_config 0 "kafka_connect_secret_providers") 0 "aws") 0 "access_key") }}
        access_key = {{ renderValue (index (index (index .kafka_user_config 0 "kafka_connect_secret_providers") 0 "aws") 0 "access_key") }}
        {{- end }}
        auth_method = {{ renderValue (required (index (index (index .kafka_user_config 0 "kafka_connect_secret_providers") 0 "aws") 0 "auth_method")) }}
        region = {{ renderValue (required (index (index (index .kafka_user_config 0 "kafka_connect_secret_providers") 0 "aws") 0 "region")) }}
        {{- if (index (index (index .kafka_user_config 0 "kafka_connect_secret_providers") 0 "aws") 0 "secret_key") }}
        secret_key = {{ renderValue (index (index (index .kafka_user_config 0 "kafka_connect_secret_providers") 0 "aws") 0 "secret_key") }}
        {{- end }}
      }
      {{- end }}
      name = {{ renderValue (required (index (index .kafka_user_config 0 "kafka_connect_secret_providers") 0 "name")) }}
      {{- if (index (index .kafka_user_config 0 "kafka_connect_secret_providers") 0 "vault") }}
      vault {
        address = {{ renderValue (required (index (index (index .kafka_user_config 0 "kafka_connect_secret_providers") 0 "vault") 0 "address")) }}
        auth_method = {{ renderValue (required (index (index (index .kafka_user_config 0 "kafka_connect_secret_providers") 0 "vault") 0 "auth_method")) }}
        {{- if (index (index (index .kafka_user_config 0 "kafka_connect_secret_providers") 0 "vault") 0 "engine_version") }}
        engine_version = {{ renderValue (index (index (index .kafka_user_config 0 "kafka_connect_secret_providers") 0 "vault") 0 "engine_version") }}
        {{- end }}
        {{- if (index (index (index .kafka_user_config 0 "kafka_connect_secret_providers") 0 "vault") 0 "prefix_path_depth") }}
        prefix_path_depth = {{ renderValue (index (index (index .kafka_user_config 0 "kafka_connect_secret_providers") 0 "vault") 0 "prefix_path_depth") }}
        {{- end }}
        {{- if (index (index (index .kafka_user_config 0 "kafka_connect_secret_providers") 0 "vault") 0 "token") }}
        token = {{ renderValue (index (index (index .kafka_user_config 0 "kafka_connect_secret_providers") 0 "vault") 0 "token") }}
        {{- end }}
      }
      {{- end }}
    }
    {{- end }}
    {{- if ne (index .kafka_user_config 0 "kafka_rest") nil }}
    kafka_rest = {{ (index .kafka_user_config 0 "kafka_rest") }}
    {{- end }}
    {{- if ne (index .kafka_user_config 0 "kafka_rest_authorization") nil }}
    kafka_rest_authorization = {{ (index .kafka_user_config 0 "kafka_rest_authorization") }}
    {{- end }}
    {{- if (index .kafka_user_config 0 "kafka_rest_config") }}
    kafka_rest_config {
      {{- if ne (index (index .kafka_user_config 0 "kafka_rest_config") 0 "consumer_enable_auto_commit") nil }}
      consumer_enable_auto_commit = {{ (index (index .kafka_user_config 0 "kafka_rest_config") 0 "consumer_enable_auto_commit") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_rest_config") 0 "consumer_idle_disconnect_timeout") }}
      consumer_idle_disconnect_timeout = {{ renderValue (index (index .kafka_user_config 0 "kafka_rest_config") 0 "consumer_idle_disconnect_timeout") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_rest_config") 0 "consumer_request_max_bytes") }}
      consumer_request_max_bytes = {{ renderValue (index (index .kafka_user_config 0 "kafka_rest_config") 0 "consumer_request_max_bytes") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_rest_config") 0 "consumer_request_timeout_ms") }}
      consumer_request_timeout_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka_rest_config") 0 "consumer_request_timeout_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_rest_config") 0 "name_strategy") }}
      name_strategy = {{ renderValue (index (index .kafka_user_config 0 "kafka_rest_config") 0 "name_strategy") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "kafka_rest_config") 0 "name_strategy_validation") nil }}
      name_strategy_validation = {{ (index (index .kafka_user_config 0 "kafka_rest_config") 0 "name_strategy_validation") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_rest_config") 0 "producer_acks") }}
      producer_acks = {{ renderValue (index (index .kafka_user_config 0 "kafka_rest_config") 0 "producer_acks") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_rest_config") 0 "producer_compression_type") }}
      producer_compression_type = {{ renderValue (index (index .kafka_user_config 0 "kafka_rest_config") 0 "producer_compression_type") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_rest_config") 0 "producer_linger_ms") }}
      producer_linger_ms = {{ renderValue (index (index .kafka_user_config 0 "kafka_rest_config") 0 "producer_linger_ms") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_rest_config") 0 "producer_max_request_size") }}
      producer_max_request_size = {{ renderValue (index (index .kafka_user_config 0 "kafka_rest_config") 0 "producer_max_request_size") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "kafka_rest_config") 0 "simpleconsumer_pool_size_max") }}
      simpleconsumer_pool_size_max = {{ renderValue (index (index .kafka_user_config 0 "kafka_rest_config") 0 "simpleconsumer_pool_size_max") }}
      {{- end }}
    }
    {{- end }}
    {{- if (index .kafka_user_config 0 "kafka_sasl_mechanisms") }}
    kafka_sasl_mechanisms {
      {{- if ne (index (index .kafka_user_config 0 "kafka_sasl_mechanisms") 0 "plain") nil }}
      plain = {{ (index (index .kafka_user_config 0 "kafka_sasl_mechanisms") 0 "plain") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "kafka_sasl_mechanisms") 0 "scram_sha_256") nil }}
      scram_sha_256 = {{ (index (index .kafka_user_config 0 "kafka_sasl_mechanisms") 0 "scram_sha_256") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "kafka_sasl_mechanisms") 0 "scram_sha_512") nil }}
      scram_sha_512 = {{ (index (index .kafka_user_config 0 "kafka_sasl_mechanisms") 0 "scram_sha_512") }}
      {{- end }}
    }
    {{- end }}
    {{- if (index .kafka_user_config 0 "kafka_version") }}
    kafka_version = {{ renderValue (index .kafka_user_config 0 "kafka_version") }}
    {{- end }}
    {{- if ne (index .kafka_user_config 0 "letsencrypt_sasl_privatelink") nil }}
    letsencrypt_sasl_privatelink = {{ (index .kafka_user_config 0 "letsencrypt_sasl_privatelink") }}
    {{- end }}
    {{- if (index .kafka_user_config 0 "private_access") }}
    private_access {
      {{- if ne (index (index .kafka_user_config 0 "private_access") 0 "kafka") nil }}
      kafka = {{ (index (index .kafka_user_config 0 "private_access") 0 "kafka") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "private_access") 0 "kafka_connect") nil }}
      kafka_connect = {{ (index (index .kafka_user_config 0 "private_access") 0 "kafka_connect") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "private_access") 0 "kafka_rest") nil }}
      kafka_rest = {{ (index (index .kafka_user_config 0 "private_access") 0 "kafka_rest") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "private_access") 0 "prometheus") nil }}
      prometheus = {{ (index (index .kafka_user_config 0 "private_access") 0 "prometheus") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "private_access") 0 "schema_registry") nil }}
      schema_registry = {{ (index (index .kafka_user_config 0 "private_access") 0 "schema_registry") }}
      {{- end }}
    }
    {{- end }}
    {{- if (index .kafka_user_config 0 "privatelink_access") }}
    privatelink_access {
      {{- if ne (index (index .kafka_user_config 0 "privatelink_access") 0 "jolokia") nil }}
      jolokia = {{ (index (index .kafka_user_config 0 "privatelink_access") 0 "jolokia") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "privatelink_access") 0 "kafka") nil }}
      kafka = {{ (index (index .kafka_user_config 0 "privatelink_access") 0 "kafka") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "privatelink_access") 0 "kafka_connect") nil }}
      kafka_connect = {{ (index (index .kafka_user_config 0 "privatelink_access") 0 "kafka_connect") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "privatelink_access") 0 "kafka_rest") nil }}
      kafka_rest = {{ (index (index .kafka_user_config 0 "privatelink_access") 0 "kafka_rest") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "privatelink_access") 0 "prometheus") nil }}
      prometheus = {{ (index (index .kafka_user_config 0 "privatelink_access") 0 "prometheus") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "privatelink_access") 0 "schema_registry") nil }}
      schema_registry = {{ (index (index .kafka_user_config 0 "privatelink_access") 0 "schema_registry") }}
      {{- end }}
    }
    {{- end }}
    {{- if (index .kafka_user_config 0 "public_access") }}
    public_access {
      {{- if ne (index (index .kafka_user_config 0 "public_access") 0 "kafka") nil }}
      kafka = {{ (index (index .kafka_user_config 0 "public_access") 0 "kafka") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "public_access") 0 "kafka_connect") nil }}
      kafka_connect = {{ (index (index .kafka_user_config 0 "public_access") 0 "kafka_connect") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "public_access") 0 "kafka_rest") nil }}
      kafka_rest = {{ (index (index .kafka_user_config 0 "public_access") 0 "kafka_rest") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "public_access") 0 "prometheus") nil }}
      prometheus = {{ (index (index .kafka_user_config 0 "public_access") 0 "prometheus") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "public_access") 0 "schema_registry") nil }}
      schema_registry = {{ (index (index .kafka_user_config 0 "public_access") 0 "schema_registry") }}
      {{- end }}
    }
    {{- end }}
    {{- if ne (index .kafka_user_config 0 "schema_registry") nil }}
    schema_registry = {{ (index .kafka_user_config 0 "schema_registry") }}
    {{- end }}
    {{- if (index .kafka_user_config 0 "schema_registry_config") }}
    schema_registry_config {
      {{- if ne (index (index .kafka_user_config 0 "schema_registry_config") 0 "leader_eligibility") nil }}
      leader_eligibility = {{ (index (index .kafka_user_config 0 "schema_registry_config") 0 "leader_eligibility") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "schema_registry_config") 0 "retriable_errors_silenced") nil }}
      retriable_errors_silenced = {{ (index (index .kafka_user_config 0 "schema_registry_config") 0 "retriable_errors_silenced") }}
      {{- end }}
      {{- if ne (index (index .kafka_user_config 0 "schema_registry_config") 0 "schema_reader_strict_mode") nil }}
      schema_reader_strict_mode = {{ (index (index .kafka_user_config 0 "schema_registry_config") 0 "schema_reader_strict_mode") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "schema_registry_config") 0 "topic_name") }}
      topic_name = {{ renderValue (index (index .kafka_user_config 0 "schema_registry_config") 0 "topic_name") }}
      {{- end }}
    }
    {{- end }}
    {{- if ne (index .kafka_user_config 0 "service_log") nil }}
    service_log = {{ (index .kafka_user_config 0 "service_log") }}
    {{- end }}
    {{- if (index .kafka_user_config 0 "single_zone") }}
    single_zone {
      {{- if ne (index (index .kafka_user_config 0 "single_zone") 0 "enabled") nil }}
      enabled = {{ (index (index .kafka_user_config 0 "single_zone") 0 "enabled") }}
      {{- end }}
    }
    {{- end }}
    {{- if ne (index .kafka_user_config 0 "static_ips") nil }}
    static_ips = {{ (index .kafka_user_config 0 "static_ips") }}
    {{- end }}
    {{- if (index .kafka_user_config 0 "tiered_storage") }}
    tiered_storage {
      {{- if ne (index (index .kafka_user_config 0 "tiered_storage") 0 "enabled") nil }}
      enabled = {{ (index (index .kafka_user_config 0 "tiered_storage") 0 "enabled") }}
      {{- end }}
      {{- if (index (index .kafka_user_config 0 "tiered_storage") 0 "local_cache") }}
      local_cache {
        {{- if (index (index (index .kafka_user_config 0 "tiered_storage") 0 "local_cache") 0 "size") }}
        size = {{ renderValue (index (index (index .kafka_user_config 0 "tiered_storage") 0 "local_cache") 0 "size") }}
        {{- end }}
      }
      {{- end }}
    }
    {{- end }}
  }
  {{- end }}
  {{- if ne .karapace nil }}
  karapace = {{ .karapace }}
  {{- end }}
  {{- if .maintenance_window_dow }}
  maintenance_window_dow = {{ renderValue .maintenance_window_dow }}
  {{- end }}
  {{- if .maintenance_window_time }}
  maintenance_window_time = {{ renderValue .maintenance_window_time }}
  {{- end }}
  plan = {{ renderValue (required .plan) }}
  project = {{ renderValue (required .project) }}
  {{- if .project_vpc_id }}
  project_vpc_id = {{ renderValue .project_vpc_id }}
  {{- end }}
  {{- if .service_integrations }}
  service_integrations {
    integration_type = {{ renderValue (required (index .service_integrations 0 "integration_type")) }}
    source_service_name = {{ renderValue (required (index .service_integrations 0 "source_service_name")) }}
  }
  {{- end }}
  service_name = {{ renderValue (required .service_name) }}
  {{- if .static_ips }}
  static_ips = [
    {{- range $idx, $item := .static_ips }}
    {{ renderValue $item }},
    {{- end }}
  ]
  {{- end }}
  {{- if .tag }}
  tag {
    key = {{ renderValue (required (index .tag 0 "key")) }}
    value = {{ renderValue (required (index .tag 0 "value")) }}
  }
  {{- end }}
  {{- if .tech_emails }}
  tech_emails {
    email = {{ renderValue (required (index .tech_emails 0 "email")) }}
  }
  {{- end }}
  {{- if ne .termination_protection nil }}
  termination_protection = {{ .termination_protection }}
  {{- end }}
  {{- if .timeouts }}
  timeouts {
    {{- if .timeouts.create }}
    create = {{ renderValue .timeouts.create }}
    {{- end }}
    {{- if .timeouts.read }}
    read = {{ renderValue .timeouts.read }}
    {{- end }}
    {{- if .timeouts.update }}
    update = {{ renderValue .timeouts.update }}
    {{- end }}
    {{- if .timeouts.delete }}
    delete = {{ renderValue .timeouts.delete }}
    {{- end }}
  }
  {{- end }}
}