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

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const ipChangeError = "the IP address of an approved Google Private Service Connect connection cannot be changed; recreate the endpoint in Google Cloud and set its new psc_connection_id"

func init() {
	// Update approves imported connections that are still pending.
	// Leave Delete unset: the Aiven API cannot revoke Google connection approval.
	ResourceOptions.Update = createView
}

func modifyPlan(_ context.Context, _ avngen.Client, d adapter.ResourceData) error {
	// Terraform replacement alone keeps the same cloud endpoint, so another
	// approve would fail with 409. Reject IP changes for that endpoint here.
	if d.IsNewResource() || !d.HasChange("user_ip_address") ||
		d.GetState("state").(string) == string(privatelink.ConnectionStateTypePendingUserApproval) {
		return nil
	}
	if _, ok := d.GetOk("user_ip_address"); !ok {
		return nil // Check the resolved IP at apply time.
	}
	for _, key := range []string{"project", "service_name", "psc_connection_id"} {
		if _, ok := d.GetOk(key); !ok || d.HasChange(key) {
			return nil // A replacement endpoint may have a different IP.
		}
	}
	return errors.New(ipChangeError)
}

// createView discovers and approves an existing Google endpoint. The adapter
// waits for active afterward: connected can mean accepted without an IP.
func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project := d.Get("project").(string)
	service := d.Get("service_name").(string)
	ip := d.Get("user_ip_address").(string)
	pscID := d.Get("psc_connection_id").(string)
	connectionID := d.Get("privatelink_connection_id").(string)

	return retry.Do(func() error {
		// Refresh queues an asynchronous cloud sync. List only reads Aiven's database.
		if err := client.ServicePrivatelinkGoogleRefresh(ctx, project, service); err != nil {
			return err
		}
		conn, err := findConnection(ctx, client, project, service, connectionID, pscID)
		if err != nil {
			if errors.Is(err, adapter.ErrNotFound) && connectionID == "" {
				// Before discovery, absence is expected while the queued sync runs.
				return fmt.Errorf("%w: waiting for Google endpoint discovery: %w", adapter.ErrRefreshStateDesired, err)
			}
			return err
		}
		// Pin retries to this record so they cannot approve a replacement endpoint.
		connectionID = conn.PrivatelinkConnectionId
		if conn.State == privatelink.ConnectionStateTypePendingUserApproval {
			rsp, approveErr := client.ServicePrivatelinkGoogleConnectionCreate(ctx, project, service, connectionID,
				&privatelink.ServicePrivatelinkGoogleConnectionCreateIn{UserIPAddress: ip})
			if approveErr == nil {
				if err := d.SetID(schemautil.BuildResourceID(project, service)); err != nil {
					return err
				}
				return flattenConnection(d, rsp)
			}

			// Approval may have succeeded despite an error; repeating it returns 409.
			// Read back the state and IP before deciding whether to retry.
			conn, err = findConnection(ctx, client, project, service, connectionID, pscID)
			if err != nil {
				return errors.Join(approveErr, err)
			}
			if conn.State == privatelink.ConnectionStateTypePendingUserApproval {
				// A 409 can also mean the parent PrivateLink is still creating.
				var retryable bool
				if apiErr, ok := errors.AsType[avngen.Error](approveErr); ok {
					retryable = apiErr.Status == http.StatusConflict || apiErr.Status == http.StatusRequestTimeout ||
						apiErr.Status == http.StatusTooManyRequests || apiErr.Status >= http.StatusInternalServerError
				} else {
					var networkErr net.Error
					retryable = errors.As(approveErr, &networkErr) || errors.Is(approveErr, io.EOF) || errors.Is(approveErr, io.ErrUnexpectedEOF)
				}
				if retryable {
					return fmt.Errorf("%w: approval has not completed: %w", adapter.ErrRefreshStateDesired, approveErr)
				}
				return approveErr
			}
		}
		if err := checkApprovedIP(conn, ip); err != nil {
			return err
		}
		if err := d.SetID(schemautil.BuildResourceID(project, service)); err != nil {
			return err
		}
		return flattenConnection(d, conn)
	}, retry.Context(ctx), retry.Attempts(0), retry.Delay(5*time.Second), retry.DelayType(retry.FixedDelay),
		retry.LastErrorOnly(true), retry.WrapContextErrorWithLastError(true),
		retry.RetryIf(func(err error) bool { return errors.Is(err, adapter.ErrRefreshStateDesired) }))
}

func readView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	// Import accepts an optional PSC selector after PROJECT/SERVICE_NAME.
	// Store only PROJECT/SERVICE_NAME to preserve SDK state compatibility.
	parts := strings.Split(d.ID(), "/")
	if len(parts) != 2 && len(parts) != 3 {
		return fmt.Errorf("expected PROJECT/SERVICE_NAME or PROJECT/SERVICE_NAME/PSC_CONNECTION_ID, got %q", d.ID())
	}
	parts, err := schemautil.SplitResourceID(d.ID(), len(parts))
	if err != nil {
		return err
	}
	if err := d.Set("project", parts[0]); err != nil {
		return err
	}
	if err := d.Set("service_name", parts[1]); err != nil {
		return err
	}
	if len(parts) == 3 && d.Get("psc_connection_id").(string) == "" {
		if err := d.Set("psc_connection_id", parts[2]); err != nil {
			return err
		}
	}
	if err := d.SetID(schemautil.BuildResourceID(parts[0], parts[1])); err != nil {
		return err
	}
	project := d.Get("project").(string)
	service := d.Get("service_name").(string)
	// The adapter supplies config only when polling after Create/Update.
	// Ordinary Read/import keep the SDK's behavior of reading only the database.
	configuredIP, applying := d.GetConfigOk("user_ip_address")
	if applying {
		// Aiven does not periodically sync Google endpoints; refresh each activation poll.
		if err := client.ServicePrivatelinkGoogleRefresh(ctx, project, service); err != nil {
			return err
		}
	}
	conn, err := findConnection(ctx, client, project, service,
		d.Get("privatelink_connection_id").(string), d.Get("psc_connection_id").(string))
	if err != nil {
		if applying && adapter.IsNotFound(err) {
			// Approval is stored synchronously and cloud syncs update the DB atomically.
			// A missing row is terminal; ErrNotFound would make the adapter retry it.
			return fmt.Errorf("%w: approved connection %q no longer exists", adapter.ErrRefreshStateFailed, d.Get("privatelink_connection_id"))
		}
		return err
	}
	if applying && conn.UserIPAddress != configuredIP.(string) {
		// Preserve the planned IP while a pending response still has an empty IP;
		// writing that empty value would violate Terraform's apply contract.
		if conn.State == privatelink.ConnectionStateTypePendingUserApproval {
			return fmt.Errorf("%w: waiting for the approved IP address", adapter.ErrRefreshStateDesired)
		}
		return checkApprovedIP(conn, configuredIP.(string))
	}
	return flattenConnection(d, conn)
}

func flattenConnection(d adapter.ResourceData, conn any) error {
	return d.Flatten(conn, func(d adapter.ResourceData, dto map[string]any) error {
		// Terraform treats an explicit "" selector as configured and must retain it.
		// Omit the response key to preserve it: Flatten would convert "" to null.
		if pscID, ok := d.GetOk("psc_connection_id"); ok && pscID.(string) == "" {
			delete(dto, "psc_connection_id")
		}
		return nil
	})
}

func checkApprovedIP(conn *privatelink.ServicePrivatelinkGoogleConnectionListOut, ip string) error {
	switch conn.State {
	case privatelink.ConnectionStateTypeUserApproved, privatelink.ConnectionStateTypeConnected, privatelink.ConnectionStateTypeActive:
		if conn.UserIPAddress != ip {
			return fmt.Errorf("connection %q has IP %q, requested %q: %s", conn.PrivatelinkConnectionId, conn.UserIPAddress, ip, ipChangeError)
		}
		return nil
	default:
		return fmt.Errorf("unexpected state %q for connection %q", conn.State, conn.PrivatelinkConnectionId)
	}
}

func findConnection(ctx context.Context, client avngen.Client, project, service, connectionID, pscID string) (*privatelink.ServicePrivatelinkGoogleConnectionListOut, error) {
	connections, err := client.ServicePrivatelinkGoogleConnectionList(ctx, project, service)
	if err != nil {
		return nil, err
	}
	// Fall back to the PSC selector or a sole connection only before the Aiven ID
	// is known. A missing stored ID must not cause adoption of another endpoint.
	conn, err := adapter.FindOne(connections, func(i int) bool {
		if connectionID != "" {
			return connections[i].PrivatelinkConnectionId == connectionID
		}
		return pscID == "" || connections[i].PscConnectionId == pscID
	})
	if err != nil {
		if errors.Is(err, adapter.ErrMultiple) {
			if pscID != "" {
				return nil, fmt.Errorf("multiple privatelink connections match psc_connection_id %q: %w", pscID, err)
			}
			return nil, fmt.Errorf("multiple privatelink connections found; set psc_connection_id to select one: %w", err)
		}
		return nil, fmt.Errorf("privatelink connection not found (privatelink_connection_id=%q, psc_connection_id=%q): %w", connectionID, pscID, err)
	}
	if pscID != "" && conn.PscConnectionId != pscID {
		return nil, fmt.Errorf("connection %q has psc_connection_id %q, expected %q", connectionID, conn.PscConnectionId, pscID)
	}
	return &conn, nil
}
