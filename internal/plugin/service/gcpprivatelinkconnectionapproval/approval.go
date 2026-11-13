package gcpprivatelinkconnectionapproval

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/privatelink"
	"github.com/avast/retry-go/v4"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const (
	connectionPollInterval = 5 * time.Second
	// Refresh queues an expensive background reconciliation. A new refresh can invalidate
	// an in-flight reconciliation, so discovery polls the saved list more often than it refreshes.
	connectionRefreshInterval = time.Minute
)

var errApprovalPending = errors.New("waiting for GCP Private Service Connect approval")

func init() {
	ResourceOptions.ImportState = importState
}

func modifyPlan(_ context.Context, _ avngen.Client, d adapter.ResourceData) error {
	if d.IsNewResource() || !d.HasChange("psc_connection_id") {
		return nil
	}
	// SDKv2 stored the actual PSC ID even when configuration used "". Normalizing
	// it keeps the same connection, selected by its saved privatelink_connection_id.
	if selector, ok := d.GetOk("psc_connection_id"); ok && selector.(string) == "" {
		return nil
	}
	// An unknown selector may resolve to a different endpoint at apply time.
	d.RequiresReplace("psc_connection_id")
	return nil
}

func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	userIPAddress := d.Get("user_ip_address").(string)
	connectionID := d.Get("privatelink_connection_id").(string)
	pscID := d.Get("psc_connection_id").(string)

	// Computed fields may be unknown in an Update plan. Keep the existing connection
	// selected even when another consumer endpoint has appeared since the last Read.
	if value, ok := d.GetStateOk("privatelink_connection_id"); connectionID == "" && ok {
		connectionID = value.(string)
	}
	if value, ok := d.GetStateOk("psc_connection_id"); pscID == "" && ok {
		pscID = value.(string)
	}

	if err := client.ServicePrivatelinkGoogleRefresh(ctx, project, serviceName); err != nil {
		return err
	}

	nextRefresh := time.Now().Add(connectionRefreshInterval)
	return retry.Do(func() error {
		connections, err := client.ServicePrivatelinkGoogleConnectionList(ctx, project, serviceName)
		if err != nil {
			return err
		}
		connection, err := findConnection(connections, connectionID, pscID)
		if errors.Is(err, adapter.ErrNotFound) {
			if !time.Now().Before(nextRefresh) {
				if err := client.ServicePrivatelinkGoogleRefresh(ctx, project, serviceName); err != nil {
					return err
				}
				nextRefresh = time.Now().Add(connectionRefreshInterval)
			}
			return fmt.Errorf("%w: %w", errApprovalPending, err)
		}
		if err != nil {
			return err
		}

		// Pin the first selection for retries. Never switch to an unrelated sole connection
		// if the original disappears between list and approve.
		connectionID, pscID = connection.PrivatelinkConnectionId, connection.PscConnectionId
		if connection.State != privatelink.ConnectionStateTypePendingUserApproval {
			if err := checkApproval(connection, userIPAddress); err != nil {
				return err
			}
			return flattenConnection(d, &connection)
		}

		approved, err := client.ServicePrivatelinkGoogleConnectionCreate(ctx, project, serviceName, connectionID,
			&privatelink.ServicePrivatelinkGoogleConnectionCreateIn{UserIPAddress: userIPAddress})
		if err != nil {
			if retryableApprovalError(err) {
				// Approve is not idempotent: a committed request retried by the HTTP client
				// returns 409. Always read again before considering another POST.
				return fmt.Errorf("%w: %w", errApprovalPending, err)
			}
			return err
		}
		connection = privatelink.ServicePrivatelinkGoogleConnectionListOut{
			PrivatelinkConnectionId: approved.PrivatelinkConnectionId,
			PscConnectionId:         approved.PscConnectionId,
			State:                   privatelink.ConnectionStateType(approved.State),
			UserIPAddress:           approved.UserIPAddress,
		}
		if err := checkApproval(connection, userIPAddress); err != nil {
			return err
		}
		return flattenConnection(d, approved)
	},
		retry.Context(ctx),
		retry.Attempts(0),
		retry.Delay(connectionPollInterval),
		retry.DelayType(retry.FixedDelay),
		retry.LastErrorOnly(true),
		retry.WrapContextErrorWithLastError(true),
		retry.RetryIf(func(err error) bool { return errors.Is(err, errApprovalPending) }),
	)
}

func retryableApprovalError(err error) bool {
	var apiError avngen.Error
	if errors.As(err, &apiError) {
		return apiError.Status == http.StatusConflict || apiError.Status == http.StatusNotFound ||
			apiError.Status == http.StatusRequestTimeout || apiError.Status == http.StatusTooManyRequests ||
			apiError.Status >= http.StatusInternalServerError
	}
	var networkError net.Error
	return errors.As(err, &networkError) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF)
}

func checkApproval(connection privatelink.ServicePrivatelinkGoogleConnectionListOut, userIPAddress string) error {
	if connection.UserIPAddress != userIPAddress {
		return fmt.Errorf("GCP Private Service Connect connection %q is %q with user_ip_address %q, expected %q; "+
			"the endpoint IP cannot be set or changed after approval or GCP acceptance. "+
			"Recreate the consumer forwarding rule and use its new psc_connection_id to approve the desired IP",
			connection.PrivatelinkConnectionId, connection.State, connection.UserIPAddress, userIPAddress)
	}
	// GCP reconciliation uses connected only when no IP is recorded in Aiven.
	// With an approved IP, user-approved transitions directly to active.
	switch connection.State {
	case privatelink.ConnectionStateTypeUserApproved, privatelink.ConnectionStateTypeActive:
		return nil
	default:
		return fmt.Errorf("GCP Private Service Connect connection %q has unexpected approval state %q", connection.PrivatelinkConnectionId, connection.State)
	}
}

func updateView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	if !d.HasChange("user_ip_address") {
		return adapter.ErrNoChanges
	}
	// Keep SDKv2's updatable schema, but surface an immutable-IP error instead of
	// silently returning the old IP. Pending connections can still be approved.
	return createView(ctx, client, d)
}

func readView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	if d.ID() != "" {
		project, serviceName, pscID, err := parseApprovalID(d.ID())
		if err != nil {
			return err
		}
		if err := d.Set("project", project); err != nil {
			return err
		}
		if err := d.Set("service_name", serviceName); err != nil {
			return err
		}
		if pscID != "" && d.Get("psc_connection_id").(string) == "" {
			if err := d.Set("psc_connection_id", pscID); err != nil {
				return err
			}
		}
	}

	// List is the only connection read endpoint. Refresh requires write permissions
	// and queues cloud changes, so it belongs only to Create/Update discovery.
	connections, err := client.ServicePrivatelinkGoogleConnectionList(ctx, d.Get("project").(string), d.Get("service_name").(string))
	if err != nil {
		return err
	}
	connection, err := findConnection(connections, d.Get("privatelink_connection_id").(string), d.Get("psc_connection_id").(string))
	if err != nil {
		return err
	}
	if configuredIP, ok := d.GetConfigOk("user_ip_address"); ok {
		// Post-Create/Update reads must agree with the configured IP. Ordinary refresh
		// and import have no configuration and report the actual remote values.
		if connection.State == privatelink.ConnectionStateTypePendingUserApproval {
			// Singleflight may share a GET started before approval committed. Retry
			// without flattening its stale, empty IP over the planned value.
			return fmt.Errorf("%w: waiting for approval of GCP Private Service Connect connection %q",
				adapter.ErrRefreshStateDesired, connection.PrivatelinkConnectionId)
		}
		if err := checkApproval(connection, configuredIP.(string)); err != nil {
			return err
		}
	}
	// The client decodes nullable response strings as ""; Flatten writes these as Terraform null.
	return flattenConnection(d, &connection)
}

func flattenConnection(d adapter.ResourceData, connection any) error {
	return d.Flatten(connection, func(d adapter.ResourceData, dto map[string]any) error {
		// An explicitly empty selector must stay empty in Terraform. Omitting the
		// response key preserves it; writing "" here would normalize it to null.
		if selector, ok := d.GetOk("psc_connection_id"); ok && selector.(string) == "" {
			delete(dto, "psc_connection_id")
		}
		return nil
	})
}

func findConnection(connections []privatelink.ServicePrivatelinkGoogleConnectionListOut, connectionID, pscID string) (privatelink.ServicePrivatelinkGoogleConnectionListOut, error) {
	if connectionID != "" {
		connection, err := adapter.FindOne(connections, func(i int) bool {
			return connections[i].PrivatelinkConnectionId == connectionID
		})
		if !errors.Is(err, adapter.ErrNotFound) || pscID == "" {
			if err != nil {
				return connection, fmt.Errorf("lookup privatelink_connection_id %q: %w", connectionID, err)
			}
			return connection, nil
		}
	}

	connection, err := adapter.FindOne(connections, func(i int) bool {
		return pscID == "" || connections[i].PscConnectionId == pscID
	})
	if errors.Is(err, adapter.ErrMultiple) {
		if pscID == "" {
			return connection, fmt.Errorf("multiple privatelink connections found; set psc_connection_id to select one: %w", err)
		}
		return connection, fmt.Errorf("multiple privatelink connections match psc_connection_id %q: %w", pscID, err)
	}
	if err != nil {
		return connection, fmt.Errorf("lookup GCP Private Service Connect connection with psc_connection_id %q: %w", pscID, err)
	}
	return connection, nil
}

func parseApprovalID(id string) (project, serviceName, pscID string, err error) {
	count := strings.Count(id, "/") + 1
	if count != 2 && count != 3 {
		return "", "", "", fmt.Errorf("invalid approval ID %q: expected PROJECT/SERVICE_NAME or PROJECT/SERVICE_NAME/PSC_CONNECTION_ID", id)
	}
	parts, err := schemautil.SplitResourceID(id, count)
	if err != nil {
		return "", "", "", err
	}
	for _, part := range parts {
		if part == "" {
			return "", "", "", fmt.Errorf("invalid approval ID %q: ID components must not be empty", id)
		}
	}
	if count == 3 {
		pscID = parts[2]
	}
	return parts[0], parts[1], pscID, nil
}

func importState(ctx context.Context, _ avngen.Client, req resource.ImportStateRequest, rsp *resource.ImportStateResponse) {
	project, serviceName, pscID, err := parseApprovalID(req.ID)
	if err != nil {
		rsp.Diagnostics.AddError("Invalid approval ID", err.Error())
		return
	}
	for name, value := range map[string]string{
		"id": schemautil.BuildResourceID(project, serviceName), "project": project, "service_name": serviceName,
	} {
		rsp.Diagnostics.Append(rsp.State.SetAttribute(ctx, path.Root(name), value)...)
	}
	if pscID != "" {
		rsp.Diagnostics.Append(rsp.State.SetAttribute(ctx, path.Root("psc_connection_id"), pscID)...)
	}
}
