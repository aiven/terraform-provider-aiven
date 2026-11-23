package kafkaquota

import (
	"context"
	"errors"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafka"
	"github.com/avast/retry-go/v4"
	"github.com/samber/lo"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

var errRatesNotUpdated = errors.New("the API returns the previous quota rates")

func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	if err := upsert(ctx, client, d); err != nil {
		return err
	}
	return d.SetID(
		d.Get("project").(string),
		d.Get("service_name").(string),
		d.Get("client_id").(string),
		d.Get("user").(string),
	)
}

func updateView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	return upsert(ctx, client, d)
}

// upsert creates or replaces the quota, then waits until Describe returns the new rates.
func upsert(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	req := new(kafka.ServiceKafkaQuotaCreateIn)
	if v, ok := d.GetOk("user"); ok {
		req.User = lo.ToPtr(v.(string))
	}
	if v, ok := d.GetOk("client_id"); ok {
		req.ClientId = lo.ToPtr(v.(string))
	}
	if v, ok := d.GetOk("consumer_byte_rate"); ok {
		req.ConsumerByteRate = lo.ToPtr(float64(v.(int)))
	}
	if v, ok := d.GetOk("producer_byte_rate"); ok {
		req.ProducerByteRate = lo.ToPtr(float64(v.(int)))
	}
	if v, ok := d.GetOk("request_percentage"); ok {
		req.RequestPercentage = lo.ToPtr(v.(float64))
	}

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	if err := client.ServiceKafkaQuotaCreate(ctx, project, serviceName, req); err != nil {
		return err
	}

	params := quotaParams(d, kafka.ServiceKafkaQuotaDescribeUser, kafka.ServiceKafkaQuotaDescribeClientId)
	return retry.Do(
		func() error {
			rsp, err := client.ServiceKafkaQuotaDescribe(ctx, project, serviceName, params...)
			if err != nil {
				return err
			}
			if lo.FromPtr(rsp.ConsumerByteRate) != lo.FromPtr(req.ConsumerByteRate) ||
				lo.FromPtr(rsp.ProducerByteRate) != lo.FromPtr(req.ProducerByteRate) ||
				lo.FromPtr(rsp.RequestPercentage) != lo.FromPtr(req.RequestPercentage) {
				return errRatesNotUpdated
			}
			return nil
		},
		retry.Context(ctx),
		retry.Attempts(0),
		retry.Delay(time.Second),
		retry.DelayType(retry.FixedDelay),
		retry.LastErrorOnly(true),
		retry.WrapContextErrorWithLastError(true),
		retry.RetryIf(func(err error) bool {
			return avngen.IsNotFound(err) || errors.Is(err, errRatesNotUpdated)
		}),
	)
}

// quotaParams selects the quota by user and client ID, either of which can be empty.
func quotaParams(d adapter.ResourceData, user, clientID func(string) [2]string) [][2]string {
	var params [][2]string
	if v := d.Get("user").(string); v != "" {
		params = append(params, user(v))
	}
	if v := d.Get("client_id").(string); v != "" {
		params = append(params, clientID(v))
	}
	return params
}

func readView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	rsp, err := client.ServiceKafkaQuotaDescribe(
		ctx,
		d.Get("project").(string),
		d.Get("service_name").(string),
		quotaParams(d, kafka.ServiceKafkaQuotaDescribeUser, kafka.ServiceKafkaQuotaDescribeClientId)...,
	)
	if err != nil {
		return err
	}

	// The API returns null for unset rates.
	rates := map[string]*float64{
		"consumer_byte_rate": rsp.ConsumerByteRate,
		"producer_byte_rate": rsp.ProducerByteRate,
		"request_percentage": rsp.RequestPercentage,
	}
	for k, v := range rates {
		var value any
		if v != nil {
			value = *v
			if k != "request_percentage" {
				value = int(*v)
			}
		}
		if err := d.Set(k, value); err != nil {
			return err
		}
	}
	return nil
}

func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	return client.ServiceKafkaQuotaDelete(
		ctx,
		d.Get("project").(string),
		d.Get("service_name").(string),
		quotaParams(d, kafka.ServiceKafkaQuotaDeleteUser, kafka.ServiceKafkaQuotaDeleteClientId)...,
	)
}
