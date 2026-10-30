package serviceutil

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestSizes(t *testing.T) {
	assert.Equal(t, 30720, toMB("30GiB"))
	assert.Equal(t, 30720, toMB("30G"))
	assert.Equal(t, 0, toMB(""))
	assert.Equal(t, "30GiB", toSize(30720))
	assert.Equal(t, "1.5GiB", toSize(1536))
	assert.Equal(t, "0B", toSize(0))
}

func TestBackupsReady(t *testing.T) {
	replica := service.ServiceIntegrationOut{IntegrationType: service.IntegrationTypeReadReplica, DestService: lo.ToPtr("foo")}
	cases := []struct {
		name string
		svc  service.ServiceGetOut
		want bool
	}{
		{name: "kafka has no backups", svc: service.ServiceGetOut{ServiceType: "kafka"}, want: true},
		{name: "pg waits for a backup", svc: service.ServiceGetOut{ServiceType: "pg"}, want: false},
		{name: "pg has a backup", svc: service.ServiceGetOut{ServiceType: "pg", Backups: []service.BackupOut{{}}}, want: true},
		{
			name: "pg replica has no backups",
			svc:  service.ServiceGetOut{ServiceType: "pg", ServiceName: "foo", ServiceIntegrations: []service.ServiceIntegrationOut{replica}},
			want: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, backupsReady(&tc.svc))
		})
	}
}

// TestGrafanaReady dials grafana, unless ip_filter blocks it, e.g., ["10.0.0.0/8"].
func TestGrafanaReady(t *testing.T) {
	dial = func(string, string, time.Duration) (net.Conn, error) {
		return nil, errors.New("unreachable")
	}
	t.Cleanup(func() { dial = net.DialTimeout })

	public := []service.ComponentOut{{Route: service.RouteTypePublic, Usage: service.UsageTypePrimary, Host: "foo", Port: 443}}
	cases := []struct {
		name     string
		ipFilter any
		want     bool
	}{
		{name: "default ip_filter dials", ipFilter: []any{"0.0.0.0/0", "::/0"}, want: false},
		{name: "objects dial too", ipFilter: []any{map[string]any{"network": "0.0.0.0/0"}}, want: false},
		{name: "private ip_filter doesn't dial", ipFilter: []any{"10.0.0.0/8"}, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &service.ServiceGetOut{ServiceType: "grafana", Components: public, UserConfig: map[string]any{"ip_filter": tc.ipFilter}}
			assert.Equal(t, tc.want, grafanaReady(svc))
		})
	}
}
