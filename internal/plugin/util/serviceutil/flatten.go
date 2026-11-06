package serviceutil

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/samber/lo"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// zeroCMKID removes the CMK, the API returns null then
const zeroCMKID = "00000000-0000-0000-0000-000000000000"

// FlattenModifier is the reverse of ExpandModifier, plus the values SDKv2 read elsewhere, e.g.,
//
//	disk_space_mb: 112640, plan: "startup-4" (80GiB)
//	                                → disk_space_used: "110GiB", additional_disk_space: "30GiB"
//	service_uri_params: {host: "foo", port: "123"}
//	                                → service_host: "foo", service_port: 123
//	tags: {"env": "prod"}           → tag: [{key: "env", value: "prod"}]
//	connection_info: {pg: ["postgres://..."]}
//	                                → pg: [{uris: ["postgres://..."]}]
func (s Service) FlattenModifier(ctx context.Context, client avngen.Client) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		rsp := new(service.ServiceGetOut)
		if err := schemautil.Remarshal(dto, rsp); err != nil {
			return err
		}

		project := d.Get("project").(string)
		plan, err := s.planSpecs(ctx, client, project, rsp.Plan)
		if err != nil {
			return err
		}

		disk := lo.FromPtr(rsp.DiskSpaceMb)
		dto["disk_space_used"] = toSize(disk)
		dto["disk_space_default"] = toSize(plan.DiskSizeMBDefault)
		dto["disk_space_step"] = toSize(plan.DiskSizeMBStep)
		dto["disk_space_cap"] = toSize(plan.DiskSizeMBMax)
		additional := ""
		if !plan.IsEmpty() && disk > plan.DiskSizeMBDefault {
			additional = toSize(disk - plan.DiskSizeMBDefault)
		}

		// Keeps "30G" when the API returns the same "30GiB"
		if toMB(additional) != toMB(d.Get("additional_disk_space").(string)) {
			dto["additional_disk_space"] = additional
		}

		if m := rsp.Maintenance; m != nil {
			dto["maintenance_window_dow"] = m.Dow
			dto["maintenance_window_time"] = m.Time
			dto["maintenance_window_enabled"] = m.Enabled
		}

		// An empty string is stored as null, but nil as ""
		dto["project_vpc_id"] = ""
		if rsp.ProjectVpcId != "" {
			dto["project_vpc_id"] = schemautil.BuildResourceID(project, rsp.ProjectVpcId)
		}

		// The zero UUID removes the CMK, then the API returns null
		if rsp.CMKId == nil && d.Get("cmk_id") != zeroCMKID {
			dto["cmk_id"] = ""
		}

		// Kafka has no credentials in the URI
		params := rsp.ServiceUriParams
		dto["service_host"] = params["host"]
		dto["service_port"], _ = strconv.Atoi(params["port"])
		user, userOK := params["user"]
		password, passwordOK := params["password"]
		for _, u := range rsp.Users {
			if u.Username == "avnadmin" && (!userOK || !passwordOK) {
				user, password, userOK, passwordOK = u.Username, u.Password, true, true
			}
		}
		if userOK {
			dto["service_username"] = user
		}
		if passwordOK {
			dto["service_password"] = password
		}

		components := make([]any, len(rsp.Components))
		for i, c := range rsp.Components {
			components[i] = map[string]any{
				"component":                   c.Component,
				"host":                        c.Host,
				"port":                        c.Port,
				"connection_uri":              fmt.Sprintf("%s:%d", c.Host, c.Port),
				"kafka_authentication_method": c.KafkaAuthenticationMethod,
				"kafka_ssl_ca":                c.KafkaSslCa,
				"route":                       c.Route,
				"ssl":                         lo.FromPtrOr(c.Ssl, true),
				"usage":                       c.Usage,
				"privatelink_connection_id":   c.PrivatelinkConnectionId,
			}
		}
		dto["components"] = components

		tags := make([]any, 0, len(rsp.Tags))
		for k, v := range rsp.Tags {
			tags = append(tags, map[string]any{"key": k, "value": v})
		}
		dto["tag"] = tags

		ips, err := serviceStaticIPs(ctx, client, project, rsp.ServiceName)
		if err != nil {
			return err
		}
		dto["static_ips"] = lo.Keys(ips)

		// The integrations set on creation, i.e., this service is the destination.
		// A resource reads them only when it has some, or it's just imported (has no plan).
		integrations := make([]any, 0)
		for _, v := range rsp.ServiceIntegrations {
			if slices.Contains(s.bootstrapIntegrations(), v.IntegrationType) && lo.FromPtr(v.DestService) == rsp.ServiceName {
				integrations = append(integrations, map[string]any{
					"integration_type":    v.IntegrationType,
					"source_service_name": v.SourceService,
				})
			}
		}
		dto["service_integrations"] = integrations
		if d.IsResource() && len(d.Get("service_integrations").([]any)) == 0 && d.Get("plan") != "" {
			delete(dto, "service_integrations")
		}

		if info := s.connectionInfo(rsp); info != nil {
			dto[s.Type] = []any{info}
		}

		for _, v := range rsp.ServiceNotifications {
			if v.Type == service.ServiceNotificationTypeServiceEndOfLife {
				const detail = "See the [documentation](%s) for more information on end of life for Aiven services."
				adapter.AddWarning(ctx, v.Message, fmt.Sprintf(detail, lo.FromPtr(v.Metadata.EndOfLifeHelpArticleUrl)))
			}
		}
		return nil
	}
}

// bootstrapIntegrations the integration types that can be set on creation only
func (s Service) bootstrapIntegrations() []service.IntegrationType {
	switch s.Type {
	case "pg":
		return []service.IntegrationType{service.IntegrationTypeReadReplica, service.IntegrationTypeDisasterRecovery}
	case "mysql", "valkey":
		return []service.IntegrationType{service.IntegrationTypeReadReplica}
	}
	return nil
}

// connectionInfo returns the connection block of the service type, e.g., the "pg" block for "pg".
// The first PG URI and params are also set on the block, as SDKv2 did.
func (s Service) connectionInfo(rsp *service.ServiceGetOut) map[string]any {
	c := lo.FromPtrOr(rsp.ConnectionInfo, service.ConnectionInfoOut{})
	switch s.Type {
	case "kafka":
		return map[string]any{
			"uris":                c.Kafka,
			"access_cert":         c.KafkaAccessCert,
			"access_key":          c.KafkaAccessKey,
			"connect_uri":         c.KafkaConnectUri,
			"rest_uri":            c.KafkaRestUri,
			"schema_registry_uri": c.SchemaRegistryUri,
		}
	case "pg":
		info := map[string]any{
			"uris":            c.PG,
			"standby_uris":    c.PGStandby,
			"syncing_uris":    c.PGSyncing,
			"replica_uri":     c.PGReplicaUri,
			"max_connections": rsp.Metadata["max_connections"],
		}
		if len(c.PG) > 0 {
			info["uri"] = c.PG[0]
		}

		params := make([]any, len(c.PGParams))
		for i, p := range c.PGParams {
			params[i] = map[string]any{
				"host":          p.Host,
				"port":          p.Port,
				"sslmode":       p.Sslmode,
				"user":          p.User,
				"password":      p.Password,
				"database_name": p.Dbname,
			}
			if i == 0 {
				info["host"] = p.Host
				info["port"] = p.Port
				info["sslmode"] = p.Sslmode
				info["user"] = p.User
				info["password"] = p.Password
				info["dbname"] = p.Dbname
			}
		}
		info["params"] = params
		return info
	case "mysql":
		params := make([]any, len(c.MysqlParams))
		for i, p := range c.MysqlParams {
			params[i] = map[string]any{
				"host":          p.Host,
				"port":          p.Port,
				"sslmode":       p.SslMode,
				"user":          p.User,
				"password":      p.Password,
				"database_name": p.Dbname,
			}
		}
		return map[string]any{
			"uris":         c.Mysql,
			"params":       params,
			"standby_uris": c.MysqlStandby,
			"replica_uri":  c.MysqlReplicaUri,
		}
	case "opensearch":
		return map[string]any{
			"uris":                      c.Opensearch,
			"opensearch_dashboards_uri": c.OpensearchDashboardsUri,
			"username":                  c.OpensearchUsername,
			"password":                  c.OpensearchPassword,
		}
	case "valkey":
		return map[string]any{
			"uris":        c.Valkey,
			"slave_uris":  c.ValkeySlave,
			"replica_uri": c.ValkeyReplicaUri,
			"password":    c.ValkeyPassword,
		}
	case "thanos":
		return map[string]any{
			"uris":                      c.Thanos,
			"query_frontend_uri":        c.QueryFrontendUri,
			"query_uri":                 c.QueryUri,
			"receiver_remote_write_uri": c.ReceiverRemoteWriteUri,
		}
	case "grafana":
		return map[string]any{"uris": c.Grafana}
	case "clickhouse":
		return map[string]any{"uris": c.Clickhouse}
	case "flink":
		return map[string]any{"host_ports": c.Flink}
	}
	return nil
}
