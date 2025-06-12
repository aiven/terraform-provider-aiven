resource "aiven_pg" "{{ required .resource_name }}" {
  {{- if .additional_disk_space }}
  additional_disk_space = {{ renderValue .additional_disk_space }}
  {{- end }}
  {{- if .cloud_name }}
  cloud_name = {{ renderValue .cloud_name }}
  {{- end }}
  {{- if .disk_space }}
  disk_space = {{ renderValue .disk_space }}
  {{- end }}
  {{- if .maintenance_window_dow }}
  maintenance_window_dow = {{ renderValue .maintenance_window_dow }}
  {{- end }}
  {{- if .maintenance_window_time }}
  maintenance_window_time = {{ renderValue .maintenance_window_time }}
  {{- end }}
  {{- if .pg }}
  pg {
    {{- if (index .pg 0 "params") }}
    params = {{ renderValue (index .pg 0 "params") }}
    {{- end }}
    {{- if (index .pg 0 "standby_uris") }}
    standby_uris = [
      {{- range $idx, $item := (index .pg 0 "standby_uris") }}
      {{ renderValue $item }},
      {{- end }}
    ]
    {{- end }}
    {{- if (index .pg 0 "syncing_uris") }}
    syncing_uris = [
      {{- range $idx, $item := (index .pg 0 "syncing_uris") }}
      {{ renderValue $item }},
      {{- end }}
    ]
    {{- end }}
    {{- if (index .pg 0 "uri") }}
    uri = {{ renderValue (index .pg 0 "uri") }}
    {{- end }}
    {{- if (index .pg 0 "uris") }}
    uris = [
      {{- range $idx, $item := (index .pg 0 "uris") }}
      {{ renderValue $item }},
      {{- end }}
    ]
    {{- end }}
  }
  {{- end }}
  {{- if .pg_user_config }}
  pg_user_config {
    {{- if (index .pg_user_config 0 "additional_backup_regions") }}
    additional_backup_regions = [
      {{- range $idx, $item := (index .pg_user_config 0 "additional_backup_regions") }}
      {{ renderValue $item }},
      {{- end }}
    ]
    {{- end }}
    {{- if (index .pg_user_config 0 "admin_password") }}
    admin_password = {{ renderValue (index .pg_user_config 0 "admin_password") }}
    {{- end }}
    {{- if (index .pg_user_config 0 "admin_username") }}
    admin_username = {{ renderValue (index .pg_user_config 0 "admin_username") }}
    {{- end }}
    {{- if (index .pg_user_config 0 "backup_hour") }}
    backup_hour = {{ renderValue (index .pg_user_config 0 "backup_hour") }}
    {{- end }}
    {{- if (index .pg_user_config 0 "backup_minute") }}
    backup_minute = {{ renderValue (index .pg_user_config 0 "backup_minute") }}
    {{- end }}
    {{- if ne (index .pg_user_config 0 "enable_ipv6") nil }}
    enable_ipv6 = {{ (index .pg_user_config 0 "enable_ipv6") }}
    {{- end }}
    {{- if (index .pg_user_config 0 "ip_filter") }}
    ip_filter = [
      {{- range $idx, $item := (index .pg_user_config 0 "ip_filter") }}
      {{ renderValue $item }},
      {{- end }}
    ]
    {{- end }}
    {{- if (index .pg_user_config 0 "ip_filter_object") }}
    ip_filter_object {
      {{- if (index (index .pg_user_config 0 "ip_filter_object") 0 "description") }}
      description = {{ renderValue (index (index .pg_user_config 0 "ip_filter_object") 0 "description") }}
      {{- end }}
      network = {{ renderValue (required (index (index .pg_user_config 0 "ip_filter_object") 0 "network")) }}
    }
    {{- end }}
    {{- if (index .pg_user_config 0 "ip_filter_string") }}
    ip_filter_string = [
      {{- range $idx, $item := (index .pg_user_config 0 "ip_filter_string") }}
      {{ renderValue $item }},
      {{- end }}
    ]
    {{- end }}
    {{- if (index .pg_user_config 0 "migration") }}
    migration {
      {{- if (index (index .pg_user_config 0 "migration") 0 "dbname") }}
      dbname = {{ renderValue (index (index .pg_user_config 0 "migration") 0 "dbname") }}
      {{- end }}
      host = {{ renderValue (required (index (index .pg_user_config 0 "migration") 0 "host")) }}
      {{- if (index (index .pg_user_config 0 "migration") 0 "ignore_dbs") }}
      ignore_dbs = {{ renderValue (index (index .pg_user_config 0 "migration") 0 "ignore_dbs") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "migration") 0 "ignore_roles") }}
      ignore_roles = {{ renderValue (index (index .pg_user_config 0 "migration") 0 "ignore_roles") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "migration") 0 "method") }}
      method = {{ renderValue (index (index .pg_user_config 0 "migration") 0 "method") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "migration") 0 "password") }}
      password = {{ renderValue (index (index .pg_user_config 0 "migration") 0 "password") }}
      {{- end }}
      port = {{ renderValue (required (index (index .pg_user_config 0 "migration") 0 "port")) }}
      {{- if ne (index (index .pg_user_config 0 "migration") 0 "ssl") nil }}
      ssl = {{ (index (index .pg_user_config 0 "migration") 0 "ssl") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "migration") 0 "username") }}
      username = {{ renderValue (index (index .pg_user_config 0 "migration") 0 "username") }}
      {{- end }}
    }
    {{- end }}
    {{- if (index .pg_user_config 0 "pg") }}
    pg {
      {{- if (index (index .pg_user_config 0 "pg") 0 "autovacuum_analyze_scale_factor") }}
      autovacuum_analyze_scale_factor = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "autovacuum_analyze_scale_factor") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "autovacuum_analyze_threshold") }}
      autovacuum_analyze_threshold = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "autovacuum_analyze_threshold") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "autovacuum_freeze_max_age") }}
      autovacuum_freeze_max_age = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "autovacuum_freeze_max_age") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "autovacuum_max_workers") }}
      autovacuum_max_workers = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "autovacuum_max_workers") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "autovacuum_naptime") }}
      autovacuum_naptime = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "autovacuum_naptime") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "autovacuum_vacuum_cost_delay") }}
      autovacuum_vacuum_cost_delay = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "autovacuum_vacuum_cost_delay") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "autovacuum_vacuum_cost_limit") }}
      autovacuum_vacuum_cost_limit = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "autovacuum_vacuum_cost_limit") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "autovacuum_vacuum_scale_factor") }}
      autovacuum_vacuum_scale_factor = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "autovacuum_vacuum_scale_factor") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "autovacuum_vacuum_threshold") }}
      autovacuum_vacuum_threshold = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "autovacuum_vacuum_threshold") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "bgwriter_delay") }}
      bgwriter_delay = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "bgwriter_delay") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "bgwriter_flush_after") }}
      bgwriter_flush_after = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "bgwriter_flush_after") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "bgwriter_lru_maxpages") }}
      bgwriter_lru_maxpages = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "bgwriter_lru_maxpages") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "bgwriter_lru_multiplier") }}
      bgwriter_lru_multiplier = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "bgwriter_lru_multiplier") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "deadlock_timeout") }}
      deadlock_timeout = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "deadlock_timeout") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "default_toast_compression") }}
      default_toast_compression = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "default_toast_compression") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "idle_in_transaction_session_timeout") }}
      idle_in_transaction_session_timeout = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "idle_in_transaction_session_timeout") }}
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "pg") 0 "jit") nil }}
      jit = {{ (index (index .pg_user_config 0 "pg") 0 "jit") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "log_autovacuum_min_duration") }}
      log_autovacuum_min_duration = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "log_autovacuum_min_duration") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "log_error_verbosity") }}
      log_error_verbosity = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "log_error_verbosity") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "log_line_prefix") }}
      log_line_prefix = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "log_line_prefix") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "log_min_duration_statement") }}
      log_min_duration_statement = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "log_min_duration_statement") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "log_temp_files") }}
      log_temp_files = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "log_temp_files") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "max_files_per_process") }}
      max_files_per_process = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "max_files_per_process") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "max_locks_per_transaction") }}
      max_locks_per_transaction = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "max_locks_per_transaction") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "max_logical_replication_workers") }}
      max_logical_replication_workers = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "max_logical_replication_workers") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "max_parallel_workers") }}
      max_parallel_workers = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "max_parallel_workers") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "max_parallel_workers_per_gather") }}
      max_parallel_workers_per_gather = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "max_parallel_workers_per_gather") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "max_pred_locks_per_transaction") }}
      max_pred_locks_per_transaction = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "max_pred_locks_per_transaction") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "max_prepared_transactions") }}
      max_prepared_transactions = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "max_prepared_transactions") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "max_replication_slots") }}
      max_replication_slots = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "max_replication_slots") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "max_slot_wal_keep_size") }}
      max_slot_wal_keep_size = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "max_slot_wal_keep_size") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "max_stack_depth") }}
      max_stack_depth = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "max_stack_depth") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "max_standby_archive_delay") }}
      max_standby_archive_delay = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "max_standby_archive_delay") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "max_standby_streaming_delay") }}
      max_standby_streaming_delay = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "max_standby_streaming_delay") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "max_wal_senders") }}
      max_wal_senders = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "max_wal_senders") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "max_worker_processes") }}
      max_worker_processes = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "max_worker_processes") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "password_encryption") }}
      password_encryption = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "password_encryption") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "pg_partman_bgw__dot__interval") }}
      pg_partman_bgw__dot__interval = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "pg_partman_bgw__dot__interval") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "pg_partman_bgw__dot__role") }}
      pg_partman_bgw__dot__role = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "pg_partman_bgw__dot__role") }}
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "pg") 0 "pg_stat_monitor__dot__pgsm_enable_query_plan") nil }}
      pg_stat_monitor__dot__pgsm_enable_query_plan = {{ (index (index .pg_user_config 0 "pg") 0 "pg_stat_monitor__dot__pgsm_enable_query_plan") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "pg_stat_monitor__dot__pgsm_max_buckets") }}
      pg_stat_monitor__dot__pgsm_max_buckets = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "pg_stat_monitor__dot__pgsm_max_buckets") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "pg_stat_statements__dot__track") }}
      pg_stat_statements__dot__track = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "pg_stat_statements__dot__track") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "temp_file_limit") }}
      temp_file_limit = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "temp_file_limit") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "timezone") }}
      timezone = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "timezone") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "track_activity_query_size") }}
      track_activity_query_size = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "track_activity_query_size") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "track_commit_timestamp") }}
      track_commit_timestamp = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "track_commit_timestamp") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "track_functions") }}
      track_functions = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "track_functions") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "track_io_timing") }}
      track_io_timing = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "track_io_timing") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "wal_sender_timeout") }}
      wal_sender_timeout = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "wal_sender_timeout") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg") 0 "wal_writer_delay") }}
      wal_writer_delay = {{ renderValue (index (index .pg_user_config 0 "pg") 0 "wal_writer_delay") }}
      {{- end }}
    }
    {{- end }}
    {{- if (index .pg_user_config 0 "pg_qualstats") }}
    pg_qualstats {
      {{- if ne (index (index .pg_user_config 0 "pg_qualstats") 0 "enabled") nil }}
      enabled = {{ (index (index .pg_user_config 0 "pg_qualstats") 0 "enabled") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg_qualstats") 0 "min_err_estimate_num") }}
      min_err_estimate_num = {{ renderValue (index (index .pg_user_config 0 "pg_qualstats") 0 "min_err_estimate_num") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pg_qualstats") 0 "min_err_estimate_ratio") }}
      min_err_estimate_ratio = {{ renderValue (index (index .pg_user_config 0 "pg_qualstats") 0 "min_err_estimate_ratio") }}
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "pg_qualstats") 0 "track_constants") nil }}
      track_constants = {{ (index (index .pg_user_config 0 "pg_qualstats") 0 "track_constants") }}
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "pg_qualstats") 0 "track_pg_catalog") nil }}
      track_pg_catalog = {{ (index (index .pg_user_config 0 "pg_qualstats") 0 "track_pg_catalog") }}
      {{- end }}
    }
    {{- end }}
    {{- if ne (index .pg_user_config 0 "pg_read_replica") nil }}
    pg_read_replica = {{ (index .pg_user_config 0 "pg_read_replica") }}
    {{- end }}
    {{- if (index .pg_user_config 0 "pg_service_to_fork_from") }}
    pg_service_to_fork_from = {{ renderValue (index .pg_user_config 0 "pg_service_to_fork_from") }}
    {{- end }}
    {{- if ne (index .pg_user_config 0 "pg_stat_monitor_enable") nil }}
    pg_stat_monitor_enable = {{ (index .pg_user_config 0 "pg_stat_monitor_enable") }}
    {{- end }}
    {{- if (index .pg_user_config 0 "pg_version") }}
    pg_version = {{ renderValue (index .pg_user_config 0 "pg_version") }}
    {{- end }}
    {{- if (index .pg_user_config 0 "pgaudit") }}
    pgaudit {
      {{- if ne (index (index .pg_user_config 0 "pgaudit") 0 "feature_enabled") nil }}
      feature_enabled = {{ (index (index .pg_user_config 0 "pgaudit") 0 "feature_enabled") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pgaudit") 0 "log") }}
      log = [
        {{- range $idx, $item := (index (index .pg_user_config 0 "pgaudit") 0 "log") }}
        {{ renderValue $item }},
        {{- end }}
      ]
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "pgaudit") 0 "log_catalog") nil }}
      log_catalog = {{ (index (index .pg_user_config 0 "pgaudit") 0 "log_catalog") }}
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "pgaudit") 0 "log_client") nil }}
      log_client = {{ (index (index .pg_user_config 0 "pgaudit") 0 "log_client") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pgaudit") 0 "log_level") }}
      log_level = {{ renderValue (index (index .pg_user_config 0 "pgaudit") 0 "log_level") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pgaudit") 0 "log_max_string_length") }}
      log_max_string_length = {{ renderValue (index (index .pg_user_config 0 "pgaudit") 0 "log_max_string_length") }}
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "pgaudit") 0 "log_nested_statements") nil }}
      log_nested_statements = {{ (index (index .pg_user_config 0 "pgaudit") 0 "log_nested_statements") }}
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "pgaudit") 0 "log_parameter") nil }}
      log_parameter = {{ (index (index .pg_user_config 0 "pgaudit") 0 "log_parameter") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pgaudit") 0 "log_parameter_max_size") }}
      log_parameter_max_size = {{ renderValue (index (index .pg_user_config 0 "pgaudit") 0 "log_parameter_max_size") }}
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "pgaudit") 0 "log_relation") nil }}
      log_relation = {{ (index (index .pg_user_config 0 "pgaudit") 0 "log_relation") }}
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "pgaudit") 0 "log_rows") nil }}
      log_rows = {{ (index (index .pg_user_config 0 "pgaudit") 0 "log_rows") }}
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "pgaudit") 0 "log_statement") nil }}
      log_statement = {{ (index (index .pg_user_config 0 "pgaudit") 0 "log_statement") }}
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "pgaudit") 0 "log_statement_once") nil }}
      log_statement_once = {{ (index (index .pg_user_config 0 "pgaudit") 0 "log_statement_once") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pgaudit") 0 "role") }}
      role = {{ renderValue (index (index .pg_user_config 0 "pgaudit") 0 "role") }}
      {{- end }}
    }
    {{- end }}
    {{- if (index .pg_user_config 0 "pgbouncer") }}
    pgbouncer {
      {{- if (index (index .pg_user_config 0 "pgbouncer") 0 "autodb_idle_timeout") }}
      autodb_idle_timeout = {{ renderValue (index (index .pg_user_config 0 "pgbouncer") 0 "autodb_idle_timeout") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pgbouncer") 0 "autodb_max_db_connections") }}
      autodb_max_db_connections = {{ renderValue (index (index .pg_user_config 0 "pgbouncer") 0 "autodb_max_db_connections") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pgbouncer") 0 "autodb_pool_mode") }}
      autodb_pool_mode = {{ renderValue (index (index .pg_user_config 0 "pgbouncer") 0 "autodb_pool_mode") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pgbouncer") 0 "autodb_pool_size") }}
      autodb_pool_size = {{ renderValue (index (index .pg_user_config 0 "pgbouncer") 0 "autodb_pool_size") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pgbouncer") 0 "ignore_startup_parameters") }}
      ignore_startup_parameters = [
        {{- range $idx, $item := (index (index .pg_user_config 0 "pgbouncer") 0 "ignore_startup_parameters") }}
        {{ renderValue $item }},
        {{- end }}
      ]
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pgbouncer") 0 "max_prepared_statements") }}
      max_prepared_statements = {{ renderValue (index (index .pg_user_config 0 "pgbouncer") 0 "max_prepared_statements") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pgbouncer") 0 "min_pool_size") }}
      min_pool_size = {{ renderValue (index (index .pg_user_config 0 "pgbouncer") 0 "min_pool_size") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pgbouncer") 0 "server_idle_timeout") }}
      server_idle_timeout = {{ renderValue (index (index .pg_user_config 0 "pgbouncer") 0 "server_idle_timeout") }}
      {{- end }}
      {{- if (index (index .pg_user_config 0 "pgbouncer") 0 "server_lifetime") }}
      server_lifetime = {{ renderValue (index (index .pg_user_config 0 "pgbouncer") 0 "server_lifetime") }}
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "pgbouncer") 0 "server_reset_query_always") nil }}
      server_reset_query_always = {{ (index (index .pg_user_config 0 "pgbouncer") 0 "server_reset_query_always") }}
      {{- end }}
    }
    {{- end }}
    {{- if (index .pg_user_config 0 "pglookout") }}
    pglookout {
      {{- if (index (index .pg_user_config 0 "pglookout") 0 "max_failover_replication_time_lag") }}
      max_failover_replication_time_lag = {{ renderValue (index (index .pg_user_config 0 "pglookout") 0 "max_failover_replication_time_lag") }}
      {{- end }}
    }
    {{- end }}
    {{- if (index .pg_user_config 0 "private_access") }}
    private_access {
      {{- if ne (index (index .pg_user_config 0 "private_access") 0 "pg") nil }}
      pg = {{ (index (index .pg_user_config 0 "private_access") 0 "pg") }}
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "private_access") 0 "pgbouncer") nil }}
      pgbouncer = {{ (index (index .pg_user_config 0 "private_access") 0 "pgbouncer") }}
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "private_access") 0 "prometheus") nil }}
      prometheus = {{ (index (index .pg_user_config 0 "private_access") 0 "prometheus") }}
      {{- end }}
    }
    {{- end }}
    {{- if (index .pg_user_config 0 "privatelink_access") }}
    privatelink_access {
      {{- if ne (index (index .pg_user_config 0 "privatelink_access") 0 "pg") nil }}
      pg = {{ (index (index .pg_user_config 0 "privatelink_access") 0 "pg") }}
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "privatelink_access") 0 "pgbouncer") nil }}
      pgbouncer = {{ (index (index .pg_user_config 0 "privatelink_access") 0 "pgbouncer") }}
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "privatelink_access") 0 "prometheus") nil }}
      prometheus = {{ (index (index .pg_user_config 0 "privatelink_access") 0 "prometheus") }}
      {{- end }}
    }
    {{- end }}
    {{- if (index .pg_user_config 0 "project_to_fork_from") }}
    project_to_fork_from = {{ renderValue (index .pg_user_config 0 "project_to_fork_from") }}
    {{- end }}
    {{- if (index .pg_user_config 0 "public_access") }}
    public_access {
      {{- if ne (index (index .pg_user_config 0 "public_access") 0 "pg") nil }}
      pg = {{ (index (index .pg_user_config 0 "public_access") 0 "pg") }}
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "public_access") 0 "pgbouncer") nil }}
      pgbouncer = {{ (index (index .pg_user_config 0 "public_access") 0 "pgbouncer") }}
      {{- end }}
      {{- if ne (index (index .pg_user_config 0 "public_access") 0 "prometheus") nil }}
      prometheus = {{ (index (index .pg_user_config 0 "public_access") 0 "prometheus") }}
      {{- end }}
    }
    {{- end }}
    {{- if (index .pg_user_config 0 "recovery_target_time") }}
    recovery_target_time = {{ renderValue (index .pg_user_config 0 "recovery_target_time") }}
    {{- end }}
    {{- if ne (index .pg_user_config 0 "service_log") nil }}
    service_log = {{ (index .pg_user_config 0 "service_log") }}
    {{- end }}
    {{- if (index .pg_user_config 0 "service_to_fork_from") }}
    service_to_fork_from = {{ renderValue (index .pg_user_config 0 "service_to_fork_from") }}
    {{- end }}
    {{- if (index .pg_user_config 0 "shared_buffers_percentage") }}
    shared_buffers_percentage = {{ renderValue (index .pg_user_config 0 "shared_buffers_percentage") }}
    {{- end }}
    {{- if ne (index .pg_user_config 0 "static_ips") nil }}
    static_ips = {{ (index .pg_user_config 0 "static_ips") }}
    {{- end }}
    {{- if (index .pg_user_config 0 "synchronous_replication") }}
    synchronous_replication = {{ renderValue (index .pg_user_config 0 "synchronous_replication") }}
    {{- end }}
    {{- if (index .pg_user_config 0 "timescaledb") }}
    timescaledb {
      {{- if (index (index .pg_user_config 0 "timescaledb") 0 "max_background_workers") }}
      max_background_workers = {{ renderValue (index (index .pg_user_config 0 "timescaledb") 0 "max_background_workers") }}
      {{- end }}
    }
    {{- end }}
    {{- if (index .pg_user_config 0 "variant") }}
    variant = {{ renderValue (index .pg_user_config 0 "variant") }}
    {{- end }}
    {{- if (index .pg_user_config 0 "work_mem") }}
    work_mem = {{ renderValue (index .pg_user_config 0 "work_mem") }}
    {{- end }}
  }
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