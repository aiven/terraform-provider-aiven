package pg

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util/serviceutil"
)

var svc = serviceutil.Service{Type: "pg"}

var (
	expandModifier  = svc.ExpandModifier
	flattenModifier = svc.FlattenModifier
	modifyPlan      = svc.ModifyPlan
)

func init() {
	svc.Wrap(&ResourceOptions, &DataSourceOptions)
	update := ResourceOptions.Update
	ResourceOptions.Update = func(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
		if err := upgradeCheck(ctx, client, d); err != nil {
			return err
		}
		return update(ctx, client, d)
	}
}

// upgradeCheck runs the "upgrade_check" task before changing pg_version, e.g., "16" → "17".
func upgradeCheck(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	config, _ := d.Get("pg_user_config").([]any)
	if len(config) == 0 {
		return nil
	}

	want, _ := config[0].(map[string]any)["pg_version"].(string)
	if want == "" {
		return nil
	}

	project, name := d.Get("project").(string), d.Get("service_name").(string)
	s, err := client.ServiceGet(ctx, project, name)
	if err != nil {
		return err
	}

	have, _ := s.UserConfig["pg_version"].(string)
	if have == want {
		return nil
	}

	task, err := client.ServiceTaskCreate(ctx, project, name, &service.ServiceTaskCreateIn{
		TargetVersion: &want,
		TaskType:      service.TaskTypeUpgradeCheck,
	})
	if err != nil {
		return fmt.Errorf("cannot create PG upgrade check task: %w", err)
	}

	var result *service.ServiceTaskGetOut
	err = serviceutil.WaitFor(ctx, 3, func() (bool, error) {
		result, err = client.ServiceTaskGet(ctx, project, name, task.TaskId)
		return err == nil && result.Result != "", err
	})
	if err != nil {
		return fmt.Errorf("error waiting for Aiven service task to be DONE: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("PG service upgrade check error, version upgrade from %s to %s, result: %s", have, want, result.Result)
	}
	return nil
}
