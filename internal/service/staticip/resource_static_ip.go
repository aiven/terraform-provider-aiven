package staticip

import (
	"context"
	"fmt"
	"log"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenStaticIPSchema = map[string]*schema.Schema{
	"project": schemautil.CommonSchemaProjectReference,

	"cloud_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: schemautil.Complex("Specifies the cloud that the static ip belongs to.").ForceNew().Build(),
	},
	"ip_address": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: schemautil.Complex("The address of the static ip").Build(),
	},
	"service_name": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: schemautil.Complex("The service name the static ip is associated with.").Build(),
	},
	"state": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: schemautil.Complex("The state the static ip is in.").Build(),
	},
	"static_ip_address_id": {
		Type:     schema.TypeString,
		Computed: true,
		Description: schemautil.Complex(
			"The static ip id of the resource. Should be used as a reference elsewhere.",
		).Build(),
	},
}

func ResourceStaticIP() *schema.Resource {
	return &schema.Resource{
		Description: "The aiven_static_ip resource allows the creation and deletion of static ips. " +
			"Please not that once a static ip is in the 'assigned' state it it is bound to the node it is assigned " +
			"to and cannot be deleted or disassociated until the node is recycled. Plans that would delete static " +
			"ips that are in the assigned state will be blocked.",
		CreateContext: resourceStaticIPCreate,
		ReadContext:   resourceStaticIPRead,
		DeleteContext: resourceStaticIPDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: aivenStaticIPSchema,
	}
}

func resourceStaticIPRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, staticIPAddressID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.StaticIPs.List(project)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	for _, sip := range r.StaticIPs {
		if sip.StaticIPAddressID == staticIPAddressID {
			err = setStaticIPState(d, project, &sip) //nolint:gosec
			if err != nil {
				return diag.Errorf("error setting static ip for resource %s: %s", d.Id(), err)
			}

			return nil
		}
	}

	return nil
}
func resourceStaticIPCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	cloudName := d.Get("cloud_name").(string)

	r, err := client.StaticIPs.Create(project, aiven.CreateStaticIPRequest{CloudName: cloudName})
	if err != nil {
		return diag.Errorf("unable to create static ip: %s", err)
	}

	d.SetId(schemautil.BuildResourceID(project, r.StaticIPAddressID))

	if err := resourceStaticIPWait(ctx, d, m); err != nil {
		return diag.Errorf("unable to wait for static ip to become active: %s", err)
	}

	return resourceStaticIPRead(ctx, d, m)
}

func resourceStaticIPDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, staticIPAddressID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	staticIP, err := client.StaticIPs.Get(project, staticIPAddressID)
	if err != nil {
		if aiven.IsNotFound(err) {
			return nil
		}

		return diag.FromErr(err)
	}

	if staticIP.State == schemautil.StaticIPAvailable {
		if err = client.StaticIPs.Dissociate(project, staticIPAddressID); err != nil {
			return diag.FromErr(err)
		}
	}

	err = client.StaticIPs.Delete(
		project,
		aiven.DeleteStaticIPRequest{
			StaticIPAddressID: staticIPAddressID,
		})
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceStaticIPWait(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project, staticIPAddressID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	conf := resource.StateChangeConf{
		Target:  []string{schemautil.StaticIPCreated},
		Pending: []string{"waiting", schemautil.StaticIPCreating},
		Timeout: d.Timeout(schema.TimeoutCreate),
		Refresh: func() (result interface{}, state string, err error) {
			log.Println("[DEBUG] checking if static ip", staticIPAddressID, "is in 'created' state")
			r, err := client.StaticIPs.List(project)
			if err != nil {
				return nil, "", fmt.Errorf("unable to fetch static ips: %w", err)
			}
			for _, sip := range r.StaticIPs {
				if sip.StaticIPAddressID == staticIPAddressID {
					log.Println("[DEBUG] static ip", staticIPAddressID, "is in state", sip.State)

					return struct{}{}, sip.State, nil
				}
			}
			log.Println("[DEBUG] static ip", staticIPAddressID, "not found in project")

			return struct{}{}, "waiting", nil
		},
	}

	if _, err := conf.WaitForStateContext(ctx); err != nil {
		return fmt.Errorf("error waiting for static ip to be created: %w", err)
	}

	return nil
}

func setStaticIPState(d *schema.ResourceData, project string, staticIP *aiven.StaticIP) error {
	if err := d.Set("project", project); err != nil {
		return fmt.Errorf("error setting static ips `project` for resource %s: %w", d.Id(), err)
	}

	if err := d.Set("cloud_name", staticIP.CloudName); err != nil {
		return fmt.Errorf("error setting static ips `cloud_name` for resource %s: %w", d.Id(), err)
	}

	if err := d.Set("ip_address", staticIP.IPAddress); err != nil {
		return fmt.Errorf("error setting static ips `ip_address` for resource %s: %w", d.Id(), err)
	}

	if err := d.Set("service_name", staticIP.ServiceName); err != nil {
		return fmt.Errorf("error setting static ips `service_name` for resource %s: %w", d.Id(), err)
	}

	if err := d.Set("state", staticIP.State); err != nil {
		return fmt.Errorf("error setting static ips `state` for resource %s: %w", d.Id(), err)
	}

	if err := d.Set("static_ip_address_id", staticIP.StaticIPAddressID); err != nil {
		return fmt.Errorf("error setting static ips `static_ip_address_id` for resource %s: %w", d.Id(), err)
	}

	return nil
}
