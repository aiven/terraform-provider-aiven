package kafka

import (
	"context"
	"fmt"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

var quotaFieldsAliases = map[string]string{
	"client_id": "client-id",
}

var aivenKafkaQuotaSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"user": {
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: true,
		Description: `
Represents a logical group of clients, assigned a unique name by the client application.
Quotas can be applied based on user, client-id, or both.
The most relevant quota is chosen for each connection.
All connections within a quota group share the same quota.
It is possible to set default quotas for each (user, client-id), user or client-id group by specifying 'default'`,
		ValidateFunc: schemautil.GetServiceUserValidateFunc(),
		AtLeastOneOf: []string{"user", "client_id"},
	},
	"client_id": {
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: true,
		Description: `
Represents a logical group of clients, assigned a unique name by the client application.
Quotas can be applied based on user, client-id, or both.
The most relevant quota is chosen for each connection.
All connections within a quota group share the same quota.
It is possible to set default quotas for each (user, client-id), user or client-id group by specifying 'default'`,
		ValidateFunc: validation.StringLenBetween(1, 255),
		AtLeastOneOf: []string{"user", "client_id"},
	},
	"consumer_byte_rate": {
		Type:     schema.TypeInt,
		Optional: true,
		Description: `
Defines the bandwidth limit in bytes/sec for each group of clients sharing a quota.
Every distinct client group is allocated a specific quota, as defined by the cluster, on a per-broker basis.
Exceeding this limit results in client throttling.`,
		ValidateFunc: validation.IntBetween(0, 1073741824),
		AtLeastOneOf: []string{"consumer_byte_rate", "producer_byte_rate", "request_percentage"},
	},
	"producer_byte_rate": {
		Type:     schema.TypeInt,
		Optional: true,
		Description: `
Defines the bandwidth limit in bytes/sec for each group of clients sharing a quota.
Every distinct client group is allocated a specific quota, as defined by the cluster, on a per-broker basis.
Exceeding this limit results in client throttling.`,
		ValidateFunc: validation.IntBetween(0, 1073741824),
		AtLeastOneOf: []string{"consumer_byte_rate", "producer_byte_rate", "request_percentage"},
	},
	"request_percentage": {
		Type:     schema.TypeFloat,
		Optional: true,
		Description: `
Sets the maximum percentage of CPU time that a client group can use on request handler I/O and network threads per broker within a quota window.
Exceeding this limit triggers throttling.
The quota, expressed as a percentage, also indicates the total allowable CPU usage for the client groups sharing the quota.`,
		ValidateFunc: validation.FloatBetween(0, 100),
		AtLeastOneOf: []string{"consumer_byte_rate", "producer_byte_rate", "request_percentage"},
	},
}

func ResourceKafkaQuota() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages quotas for an Aiven for Apache Kafka® service user.",
		ReadContext:   common.WithGenClient(resourceKafkaQuotaRead),
		CreateContext: common.WithGenClient(resourceKafkaQuotaCreate),
		UpdateContext: common.WithGenClient(resourceKafkaQuotaUpdate),
		DeleteContext: common.WithGenClient(resourceKafkaQuotaDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenKafkaQuotaSchema,
	}
}

func resourceKafkaQuotaCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var (
		project  = d.Get("project").(string)
		service  = d.Get("service_name").(string)
		user     = d.Get("user").(string)
		clientID = d.Get("client_id").(string)

		req kafka.ServiceKafkaQuotaCreateIn
	)

	if err := schemautil.ResourceDataGet(
		d,
		&req,
		schemautil.RenameAliases(quotaFieldsAliases),
	); err != nil {
		return err
	}

	if err := client.ServiceKafkaQuotaCreate(ctx, project, service, &req); err != nil {
		return err
	}

	d.SetId(schemautil.BuildResourceID(project, service, clientID, user))

	return readKafkaQuotaWithRetry(ctx, d, client)
}

func resourceKafkaQuotaUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var req kafka.ServiceKafkaQuotaCreateIn

	project, service, _, _, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return err
	}

	if err := schemautil.ResourceDataGet(
		d,
		&req,
		schemautil.RenameAliases(quotaFieldsAliases),
	); err != nil {
		return err
	}

	if err := client.ServiceKafkaQuotaCreate(ctx, project, service, &req); err != nil {
		return err
	}

	return readKafkaQuotaWithRetry(ctx, d, client)
}

func resourceKafkaQuotaRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, clientID, user, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return err
	}

	var params [][2]string
	if user != "" {
		params = append(params, kafka.ServiceKafkaQuotaDescribeUser(user))
	}

	if clientID != "" {
		params = append(params, kafka.ServiceKafkaQuotaDescribeClientId(clientID))
	}

	if len(params) == 0 {
		return fmt.Errorf("invalid resource ID: %q, either user or client_id must be set", d.Id())
	}

	resp, err := client.ServiceKafkaQuotaDescribe(
		ctx,
		projectName,
		serviceName,
		params...,
	)
	if err != nil {
		return schemautil.ResourceReadHandleNotFound(err, d)
	}

	return schemautil.ResourceDataSet(
		d, resp, aivenKafkaQuotaSchema,
		schemautil.RenameAliasesReverse(quotaFieldsAliases),
		schemautil.SetForceNew("project", projectName),
		schemautil.SetForceNew("service_name", serviceName),
	)
}

func resourceKafkaQuotaDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var (
		project     = d.Get("project").(string)
		serviceName = d.Get("service_name").(string)
		clientID    = d.Get("client_id").(string)
		user        = d.Get("user").(string)
	)

	var params [][2]string
	if user != "" {
		params = append(params, kafka.ServiceKafkaQuotaDeleteUser(user))
	}

	if clientID != "" {
		params = append(params, kafka.ServiceKafkaQuotaDeleteClientId(clientID))
	}

	return client.ServiceKafkaQuotaDelete(
		ctx,
		project,
		serviceName,
		params...,
	)
}

// readKafkaQuotaWithRetry retries reading a Kafka quota resource to handle eventual consistency on API side.
func readKafkaQuotaWithRetry(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	expectedConsumerRate := d.Get("consumer_byte_rate")
	expectedProducerRate := d.Get("producer_byte_rate")
	expectedRequestPct := d.Get("request_percentage")

	return retry.RetryContext(ctx, 5*time.Second, func() *retry.RetryError {
		err := resourceKafkaQuotaRead(ctx, d, client)
		if err != nil {
			// resource may not be yet available after create
			if avngen.IsNotFound(err) {
				return &retry.RetryError{Err: err, Retryable: true}
			}

			return &retry.RetryError{Err: err, Retryable: false}
		}

		// verify that the values we read match what we expect (eventual consistency on API side)
		actualConsumerRate := d.Get("consumer_byte_rate")
		actualProducerRate := d.Get("producer_byte_rate")
		actualRequestPct := d.Get("request_percentage")

		mismatch := false
		if expectedConsumerRate != nil && expectedConsumerRate != actualConsumerRate {
			mismatch = true
		}
		if expectedProducerRate != nil && expectedProducerRate != actualProducerRate {
			mismatch = true
		}
		if expectedRequestPct != nil && expectedRequestPct != actualRequestPct {
			mismatch = true
		}

		if mismatch {
			return &retry.RetryError{
				Err:       fmt.Errorf("quota values not yet updated on backend"),
				Retryable: true,
			}
		}

		return nil
	})
}
