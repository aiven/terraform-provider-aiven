// Package serviceutil is the logic the service resources share, e.g., aiven_pg and aiven_kafka.
package serviceutil

import (
	"context"
	"fmt"
	"slices"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/aiven/go-client-codegen/handler/staticip"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

type operation = func(context.Context, avngen.Client, adapter.ResourceData) error

// Service is the logic of a service type. Its methods are the hooks the generated code calls:
//
//	var svc = serviceutil.Service{Type: "pg"}
//	var expandModifier = svc.ExpandModifier
type Service struct {
	Type string
}

// userConfig is the user config block, e.g., "pg_user_config".
func (s Service) userConfig() string {
	return s.Type + "_user_config"
}

// Wrap adds the steps around the generated operations:
//   - plan marks the computed-only values unknown on a change that can reset them,
//   - create waits for the service and sets the tags,
//   - update associates the new static IPs, waits, dissociates the removed ones, sets the tags,
//   - delete waits until the static IPs are released,
//   - read fails on redacted credentials.
func (s Service) Wrap(r *adapter.ResourceOptions, ds *adapter.DataSourceOptions) {
	// E.g., service_host and disk_space_used. The ID and service_type never change.
	var computed []string
	for k, v := range r.Schema(context.Background()).Attributes {
		if v.IsComputed() && !v.IsOptional() && k != "id" && k != "service_type" {
			computed = append(computed, k)
		}
	}

	modifyPlan := r.ModifyPlan
	r.ModifyPlan = func(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
		if err := modifyPlan(ctx, client, d); err != nil || d.IsNewResource() {
			return err
		}

		// These change the disks, the hosts, the credentials, etc. Only kafka has karapace.
		changed := slices.ContainsFunc([]string{"plan", "cloud_name", "project_vpc_id", "disk_space", "additional_disk_space", "karapace", s.userConfig()}, func(k string) bool {
			_, ok := d.Schema().Properties[k]
			return ok && d.HasChange(k)
		})
		if !changed {
			return nil
		}
		return d.SetUnknown(computed...)
	}

	create, update, del := r.Create, r.Update, r.Delete
	r.Create = func(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
		// The create view stores the response, which has no tags and static IPs yet
		tags, ips := d.Get("tag"), toStrings(d.Get("static_ips"))
		if err := create(ctx, client, d); err != nil {
			return err
		}

		project, name := d.Get("project").(string), d.Get("service_name").(string)
		rsp, err := waitForService(ctx, client, project, name, ips, true)
		if err != nil {
			return fmt.Errorf("error waiting for service creation: %w", err)
		}

		// ServiceGet may still return 404 for a new service, while its databases, users, etc. can be created
		schemautil.ServiceReadyStore(rsp, project, name)
		return replaceTags(ctx, client, project, name, tags)
	}

	r.Update = func(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
		project, name := d.Get("project").(string), d.Get("service_name").(string)
		tags, ips := d.Get("tag"), toStrings(d.Get("static_ips"))
		have, err := serviceStaticIPs(ctx, client, project, name)
		if err != nil {
			return err
		}

		// Associates first, so static_ips can be enabled in the user config of an existing service
		for _, ip := range ips {
			if _, ok := have[ip]; ok {
				continue
			}
			_, err := client.ProjectStaticIPAssociate(ctx, project, ip, &staticip.ProjectStaticIPAssociateIn{ServiceName: name})
			if err != nil {
				return fmt.Errorf("error associating Static IP (%s) to a service: %w", ip, err)
			}
		}

		if err := update(ctx, client, d); err != nil {
			return err
		}

		rsp, err := waitForService(ctx, client, project, name, ips, false)
		if err != nil {
			return fmt.Errorf("error waiting for service (%s) update: %w", name, err)
		}

		// The update powers the service on, so a POWEROFF result is stale
		if rsp.State == service.ServiceStateTypePoweroff {
			schemautil.ServiceReadyForget(project, name)
		} else {
			schemautil.ServiceReadyStore(rsp, project, name)
		}

		removed := false
		for ip := range have {
			if slices.Contains(ips, ip) {
				continue
			}
			removed = true
			if _, err := client.ProjectStaticIPDissociate(ctx, project, ip); err != nil {
				return fmt.Errorf("error dissociating Static IP (%s) from the service (%s): %w", ip, name, err)
			}
		}

		if removed {
			err := WaitFor(ctx, readyTimes, func() (bool, error) {
				have, err := serviceStaticIPs(ctx, client, project, name)
				for ip := range have {
					if !slices.Contains(ips, ip) {
						return false, err
					}
				}
				return true, err
			})
			if err != nil {
				return fmt.Errorf("error waiting for Static IPs dissociation: %w", err)
			}
		}
		return replaceTags(ctx, client, project, name, tags)
	}

	r.Delete = func(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
		project, name := d.Get("project").(string), d.Get("service_name").(string)
		ips := toStrings(d.Get("static_ips"))
		if err := del(ctx, client, d); common.IsCritical(err) {
			return err
		}
		schemautil.ServiceReadyForget(project, name)

		// The released static IPs stay "assigned" for a while
		err := WaitFor(ctx, readyTimes, func() (bool, error) {
			if _, err := client.ServiceGet(ctx, project, name); common.IsCritical(err) {
				return false, fmt.Errorf("unable to check if service is gone: %w", err)
			}

			if len(ips) == 0 {
				return true, nil
			}

			list, err := client.StaticIPList(ctx, project)
			if err != nil {
				return false, err
			}

			for _, v := range list {
				if v.State == staticip.StaticIPStateTypeAssigned && slices.Contains(ips, v.StaticIPAddressId) {
					return false, nil
				}
			}
			return true, nil
		})
		if err != nil {
			return fmt.Errorf("error waiting for service deletion: %w", err)
		}
		return nil
	}

	r.Read = s.checkRedactedCreds(r.Read)
	ds.Read = s.checkRedactedCreds(ds.Read)
}

// checkRedactedCreds fails when the read returns redacted credentials, though it requests secrets.
// Only the read requests secrets, so the Create and Update responses aren't checked.
func (s Service) checkRedactedCreds(read operation) operation {
	return func(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
		if err := read(ctx, client, d); err != nil {
			return err
		}
		return schemautil.ContainsRedactedCreds(map[string]any{"config": d.Get(s.userConfig())})
	}
}

// replaceTags sets the tags, e.g., [{key: "env", value: "prod"}] as {"env": "prod"}.
func replaceTags(ctx context.Context, client avngen.Client, project, name string, tags any) error {
	in := &service.ProjectServiceTagsReplaceIn{Tags: make(map[string]string)}
	list, _ := tags.([]any)
	for _, v := range list {
		tag := v.(map[string]any)
		in.Tags[tag["key"].(string)] = tag["value"].(string)
	}

	if err := client.ProjectServiceTagsReplace(ctx, project, name, in); err != nil {
		return fmt.Errorf("error setting service tags: %w", err)
	}
	return nil
}

// serviceStaticIPs returns the static IPs of the service and their states.
func serviceStaticIPs(ctx context.Context, client avngen.Client, project, name string) (map[string]staticip.StaticIPStateType, error) {
	list, err := client.StaticIPList(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch static ips for project %q: %w", project, err)
	}

	result := make(map[string]staticip.StaticIPStateType)
	for _, v := range list {
		if v.ServiceName == name {
			result[v.StaticIPAddressId] = v.State
		}
	}
	return result, nil
}

// toStrings returns the known strings of a list: the plan has an unknown one as tftypes.Value.
func toStrings(v any) []string {
	list, _ := v.([]any)
	result := make([]string, 0, len(list))
	for _, item := range list {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
