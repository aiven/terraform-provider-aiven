package vpc

import (
	"context"
	"log"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

var aivenAWSPrivatelinkSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"principals": {
		Type:        schema.TypeSet,
		Required:    true,
		Description: "List of the ARNs of the AWS accounts or IAM users allowed to connect to the VPC endpoint.",
		Elem:        &schema.Schema{Type: schema.TypeString},
	},
	"aws_service_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "AWS service ID.",
	},
	"aws_service_name": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "AWS service name.",
	},
}

func ResourceAWSPrivatelink() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an [AWS PrivateLink for Aiven services](https://aiven.io/docs/platform/howto/use-aws-privatelinks) in a VPC.",
		CreateContext: resourceAWSPrivatelinkCreate,
		ReadContext:   resourceAWSPrivatelinkRead,
		UpdateContext: resourceAWSPrivatelinkUpdate,
		DeleteContext: resourceAWSPrivatelinkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenAWSPrivatelinkSchema,
	}
}

func resourceAWSPrivatelinkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	var project = d.Get("project").(string)
	var serviceName = d.Get("service_name").(string)

	var principalsSet = d.Get("principals").(*schema.Set)
	principals := make([]string, principalsSet.Len())

	for i, p := range principalsSet.List() {
		principals[i] = p.(string)
	}

	_, err := client.AWSPrivatelink.Create(
		ctx,
		project,
		serviceName,
		principals,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait until the AWS privatelink is active
	w := &AWSPrivatelinkWaiter{
		Context:     ctx,
		Client:      m.(*aiven.Client),
		Project:     project,
		ServiceName: serviceName,
	}

	_, err = w.Conf(d.Timeout(schema.TimeoutCreate)).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for AWS privatelink creation: %s", err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName))

	return resourceAWSPrivatelinkRead(ctx, d, m)
}

func resourceAWSPrivatelinkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	p, err := client.AWSPrivatelink.Get(ctx, project, serviceName)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if err := d.Set("principals", p.Principals); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("aws_service_id", p.AWSServiceID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("aws_service_name", p.AWSServiceName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project", project); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceAWSPrivatelinkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	principalsSet := d.Get("principals").(*schema.Set)
	principals := make([]string, principalsSet.Len())

	for i, p := range principalsSet.List() {
		principals[i] = p.(string)
	}

	_, err = client.AWSPrivatelink.Update(
		ctx,
		project,
		serviceName,
		principals,
	)
	if err != nil && !aiven.IsAlreadyExists(err) {
		return diag.FromErr(err)
	}

	// Wait until the AWS privatelink is active
	w := &AWSPrivatelinkWaiter{
		Context:     ctx,
		Client:      m.(*aiven.Client),
		Project:     project,
		ServiceName: serviceName,
	}

	_, err = w.Conf(d.Timeout(schema.TimeoutCreate)).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for AWS privatelink to be updated: %s", err)
	}

	return resourceAWSPrivatelinkRead(ctx, d, m)
}

func resourceAWSPrivatelinkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.AWSPrivatelink.Delete(ctx, project, serviceName)
	if common.IsCritical(err) {
		return diag.FromErr(err)
	}

	return nil
}

// AWSPrivatelinkWaiter is used to wait for Aiven to build a AWS privatelink
type AWSPrivatelinkWaiter struct {
	Context     context.Context
	Client      *aiven.Client
	Project     string
	ServiceName string
}

// RefreshFunc will call the Aiven client and refresh its state.
func (w *AWSPrivatelinkWaiter) RefreshFunc() retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		pc, err := w.Client.AWSPrivatelink.Get(w.Context, w.Project, w.ServiceName)
		if err != nil {
			return nil, "", err
		}

		log.Printf("[DEBUG] Got %s state while waiting for AWS privatelink to be active.", pc.State)

		return pc, pc.State, nil
	}
}

// Conf sets up the configuration to refresh.
func (w *AWSPrivatelinkWaiter) Conf(timeout time.Duration) *retry.StateChangeConf {
	log.Printf("[DEBUG] Create waiter timeout %.0f minutes", timeout.Minutes())

	return &retry.StateChangeConf{
		Pending: []string{"creating"},
		Target:  []string{"active"},
		Refresh: w.RefreshFunc(),
		Delay:   10 * time.Second,
		Timeout: timeout,
	}
}
