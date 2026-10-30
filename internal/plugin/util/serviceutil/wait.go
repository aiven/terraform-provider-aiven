package serviceutil

import (
	"context"
	"net"
	"slices"
	"strconv"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/aiven/go-client-codegen/handler/staticip"
	"github.com/samber/lo"
)

// dial is replaced by the tests
var dial = net.DialTimeout

const (
	pollInterval = 5 * time.Second

	// readyTimes a service can flap right after a change, e.g., RUNNING, REBUILDING, RUNNING
	readyTimes = 5
)

// WaitFor polls until done returns true the given times in a row, or an error, or the context is done.
func WaitFor(ctx context.Context, times int, done func() (bool, error)) error {
	for count := 0; ; {
		ok, err := done()
		if err != nil {
			return err
		}

		count++
		if !ok {
			count = 0
		}

		if count == times {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

// waitForService waits until the service is RUNNING (unless running is false), has a backup,
// its grafana is reachable and its static IPs are assigned.
func waitForService(ctx context.Context, client avngen.Client, project, name string, ips []string, running bool) (*service.ServiceGetOut, error) {
	var rsp *service.ServiceGetOut
	err := WaitFor(ctx, readyTimes, func() (bool, error) {
		var err error
		rsp, err = client.ServiceGet(ctx, project, name)
		switch {
		case avngen.IsNotFound(err):
			// A new service can be missing for a moment
			return false, nil
		case err != nil:
			return false, err
		case running && rsp.State != service.ServiceStateTypeRunning, !backupsReady(rsp), !grafanaReady(rsp):
			return false, nil
		case len(ips) == 0:
			return true, nil
		}

		have, err := serviceStaticIPs(ctx, client, project, name)
		for _, ip := range ips {
			switch have[ip] {
			case staticip.StaticIPStateTypeAvailable, staticip.StaticIPStateTypeAssigned:
			default:
				return false, err
			}
		}
		return true, err
	})
	return rsp, err
}

// backupsReady pg has a backup soon after creation, see
// https://github.com/aiven/terraform-provider-aiven/issues/756
func backupsReady(s *service.ServiceGetOut) bool {
	if s.ServiceType != "pg" {
		return true
	}

	// Read replicas have no backups, see https://github.com/aiven/terraform-provider-aiven/pull/172
	for _, i := range s.ServiceIntegrations {
		switch i.IntegrationType {
		case service.IntegrationTypeReadReplica, service.IntegrationTypeDisasterRecovery:
			if lo.FromPtr(i.DestService) == s.ServiceName {
				return true
			}
		}
	}
	return len(s.Backups) > 0
}

// grafanaReady dials the public grafana, unless the ip_filter blocks it,
// i.e., anything but ["0.0.0.0/0"] or ["0.0.0.0/0", "::/0"].
func grafanaReady(s *service.ServiceGetOut) bool {
	if s.ServiceType != "grafana" {
		return true
	}

	if v, ok := s.UserConfig["ip_filter"]; ok {
		list, _ := v.([]any)
		cidrs := make([]string, len(list))
		for i, f := range list {
			// Either "0.0.0.0/0" or {"network": "0.0.0.0/0"}
			if m, ok := f.(map[string]any); ok {
				f = m["network"]
			}
			cidrs[i], _ = f.(string)
		}

		slices.Sort(cidrs)
		if !slices.Equal(cidrs, []string{"0.0.0.0/0"}) && !slices.Equal(cidrs, []string{"0.0.0.0/0", "::/0"}) {
			return true
		}
	}

	for _, c := range s.Components {
		if c.Route == service.RouteTypePublic && c.Usage == service.UsageTypePrimary {
			conn, err := dial("tcp", net.JoinHostPort(c.Host, strconv.Itoa(c.Port)), time.Second)
			if err != nil {
				return false
			}
			_ = conn.Close()
		}
	}
	return true
}
