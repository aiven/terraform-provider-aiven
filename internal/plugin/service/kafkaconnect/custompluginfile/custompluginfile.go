package custompluginfile

import (
	"context"
	"errors"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkaconnectcustomplugin"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/fileupload"
)

// fileSha256Field is both the API-reported hash and the plan-time replacement trigger.
const fileSha256Field = "file_sha256"

func init() {
	ResourceOptions.Create = createView
	ResourceOptions.RefreshStateCheck = waitForReady
}

// waitForReady keeps the post-create/update refresh polling until the backend has verified
// the upload and reported its sha256.
func waitForReady(d adapter.ResourceData) error {
	status, _ := d.Get("file_status").(string)
	switch kafkaconnectcustomplugin.FileStatusType(status) {
	case kafkaconnectcustomplugin.FileStatusTypeFailed:
		message, _ := d.Get("verify_error_message").(string)
		return fmt.Errorf("%w: plugin file verification failed: %s", adapter.ErrRefreshStateFailed, message)
	case kafkaconnectcustomplugin.FileStatusTypeReady:
		if sha, _ := d.Get(fileSha256Field).(string); sha != "" {
			return nil
		}
	}
	return fmt.Errorf("plugin file not ready: status=%q", status)
}

// modifyPlan hashes the local file so a content edit shows up in the plan.
func modifyPlan(_ context.Context, _ avngen.Client, d adapter.ResourceData) error {
	return fileupload.ModifyPlan(d, "source", fileSha256Field)
}

// createView creates the plugin file record and uploads to the pre-signed URL.
// The framework then polls readView until file_status reaches READY or FAILED.
func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	organizationID := d.Get("organization_id").(string)
	contentType := d.Get("content_type").(string)

	in := &kafkaconnectcustomplugin.KafkaConnectCustomPluginFileCreateIn{
		ContentType:   kafkaconnectcustomplugin.ContentType(contentType),
		PluginName:    d.Get("plugin_name").(string),
		PluginVersion: d.Get("plugin_version").(string),
		ServiceType:   kafkaconnectcustomplugin.ServiceType(d.Get("service_type").(string)),
	}
	if v, ok := d.GetOk("file_description"); ok {
		s := v.(string)
		in.FileDescription = &s
	}

	rsp, err := client.KafkaConnectCustomPluginFileCreate(ctx, organizationID, in)
	if err != nil {
		return fmt.Errorf("failed to create custom plugin file: %w", err)
	}

	if rsp.UploadInfo.Url == nil {
		return fmt.Errorf("API returned no upload URL")
	}

	// plugin_file_id must be in state before uploadFile can fail into deleteView.
	if err := d.Set("plugin_file_id", rsp.PluginFileId); err != nil {
		return err
	}

	checksum := d.Get(fileSha256Field).(string)
	if err := fileupload.Upload(ctx, d.Get("source").(string), checksum, contentType, *rsp.UploadInfo.Url); err != nil {
		return errors.Join(err, deleteView(ctx, client, d))
	}

	return nil
}
