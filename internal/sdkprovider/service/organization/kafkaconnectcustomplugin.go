package organization

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkaconnectcustomplugin"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

const (
	contentTypeJAR = "application/java-archive"
	contentTypeZIP = "application/zip"
)

var aivenKafkaConnectCustomPluginFileSchema = map[string]*schema.Schema{
	"organization_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The ID of the organization the plugin belongs to.").ForceNew().Build(),
	},
	"plugin_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("Name identifying this custom plugin (e.g. `my-custom-connectors`). Alphanumeric, dashes and underscores, 1–64 characters.").ForceNew().Build(),
	},
	"plugin_version": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("Semantic version for this plugin file upload (e.g. `2.7.14`).").ForceNew().Build(),
	},
	"service_type": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Default:     "kafka_connect",
		Description: userconfig.Desc("The Aiven service type this plugin targets. Currently only `kafka_connect` is supported.").ForceNew().Build(),
	},
	"source": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("Local path to the JAR or ZIP file to upload.").ForceNew().Build(),
		DiffSuppressFunc: func(_, _, _ string, d *schema.ResourceData) bool {
			// Ignore filename-only renames; use source_checksum to detect real changes.
			return d.Id() != ""
		},
	},
	"content_type": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Default:     contentTypeJAR,
		Description: userconfig.Desc("MIME type of the plugin file. Use `application/java-archive` for a single JAR or `application/zip` for a ZIP bundle.").ForceNew().Build(),
		ValidateFunc: func(v any, k string) ([]string, []error) {
			s := v.(string)
			if s != contentTypeJAR && s != contentTypeZIP {
				return nil, []error{fmt.Errorf("%q must be %q or %q", k, contentTypeJAR, contentTypeZIP)}
			}
			return nil, nil
		},
	},
	"plugin_description": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Human-readable description of the plugin (applies to all versions). Can be set on the first file upload and updated at any time via the plugin-level API.",
	},
	"file_description": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Description: "Human-readable change notes specific to this plugin version.",
	},
	"source_checksum": {
		Type:        schema.TypeString,
		Computed:    true,
		ForceNew:    true,
		Description: "SHA-256 checksum of the local file. Computed automatically; forces replacement when the file content changes.",
	},

	// Computed / read-only
	"plugin_file_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Unique identifier for this custom plugin file upload.",
	},
	"plugin_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Unique identifier for the plugin identity record (shared across all versions).",
	},
	"file_status": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Verification status of the uploaded file. Possible values: `INITIAL`, `READY`, `FAILED`.",
	},
	"file_sha256": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "SHA-256 hash of the uploaded file, populated after successful verification.",
	},
	"file_size": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Size of the uploaded file in bytes, populated after successful verification.",
	},
	"verify_error_code": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Machine-readable error code when `file_status` is `FAILED`.",
	},
	"verify_error_message": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Human-readable error message when `file_status` is `FAILED`.",
	},
	"created_by": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Email address of the user who created this resource.",
	},
	"updated_by": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Email address of the user who last updated this resource.",
	},
	"created_at": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Creation timestamp in ISO 8601 format, always in UTC.",
	},
	"updated_at": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Last-update timestamp in ISO 8601 format, always in UTC.",
	},
}

// ResourceKafkaConnectCustomPluginFile returns the Terraform resource definition.
func ResourceKafkaConnectCustomPluginFile() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an Aiven Kafka Connect custom plugin file at the organization level.",
		CreateContext: common.WithGenClient(kafkaConnectCustomPluginFileCreate),
		ReadContext:   common.WithGenClient(kafkaConnectCustomPluginFileRead),
		UpdateContext: common.WithGenClient(kafkaConnectCustomPluginFileUpdate),
		DeleteContext: common.WithGenClient(kafkaConnectCustomPluginFileDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   aivenKafkaConnectCustomPluginFileSchema,
		CustomizeDiff: func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
			sourcePath := diff.Get("source").(string)
			if sourcePath == "" {
				return nil
			}
			checksum, err := pluginFileChecksum(sourcePath)
			if err != nil {
				return fmt.Errorf("failed to calculate checksum for %q: %w", sourcePath, err)
			}
			return diff.SetNew("source_checksum", checksum)
		},
	}
}

func kafkaConnectCustomPluginFileCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	organizationID := d.Get("organization_id").(string)
	sourcePath := d.Get("source").(string)
	sourceChecksum := d.Get("source_checksum").(string)
	contentType := d.Get("content_type").(string)

	in := &kafkaconnectcustomplugin.KafkaConnectCustomPluginFileCreateIn{
		ContentType:   kafkaconnectcustomplugin.ContentType(contentType),
		PluginName:    d.Get("plugin_name").(string),
		PluginVersion: d.Get("plugin_version").(string),
		ServiceType:   kafkaconnectcustomplugin.ServiceType(d.Get("service_type").(string)),
	}
	if v, ok := d.GetOk("plugin_description"); ok {
		s := v.(string)
		in.PluginDescription = &s
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
		return fmt.Errorf("API returned no upload URL; cannot proceed with file upload")
	}

	if err := uploadPluginFile(ctx, sourcePath, sourceChecksum, contentType, *rsp.UploadInfo.Url); err != nil {
		return fmt.Errorf("failed to upload plugin file: %w", err)
	}

	d.SetId(schemautil.BuildResourceID(organizationID, rsp.PluginFileId))

	// Poll until file_status reaches READY or FAILED.
	err = retry.Do(
		func() error {
			info, err := client.KafkaConnectCustomPluginFileGet(ctx, organizationID, rsp.PluginFileId)
			switch {
			case avngen.IsNotFound(err):
				return retry.Unrecoverable(fmt.Errorf("plugin file %q not found after upload: %w", rsp.PluginFileId, err))
			case err != nil:
				return err
			case info.FileStatus == kafkaconnectcustomplugin.FileStatusTypeFailed:
				msg := "plugin file verification failed"
				if info.VerifyErrorMessage != nil {
					msg = fmt.Sprintf("%s: %s", msg, *info.VerifyErrorMessage)
				}
				return retry.Unrecoverable(fmt.Errorf("%s", msg))
			case info.FileStatus != kafkaconnectcustomplugin.FileStatusTypeReady:
				return fmt.Errorf("plugin file status is %q, waiting for READY", info.FileStatus)
			}
			return nil
		},
		retry.Context(ctx),
		retry.Delay(time.Second*5),
	)
	if err != nil {
		return fmt.Errorf("failed waiting for plugin file to become ready: %w", err)
	}

	return kafkaConnectCustomPluginFileRead(ctx, d, client)
}

func kafkaConnectCustomPluginFileRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	organizationID, pluginFileID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	info, err := client.KafkaConnectCustomPluginFileGet(ctx, organizationID, pluginFileID)
	if err != nil {
		return schemautil.ResourceReadHandleNotFound(err, d)
	}

	return setCustomPluginFileData(d, organizationID, info)
}

func kafkaConnectCustomPluginFileUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	if !d.HasChange("plugin_description") {
		return nil
	}

	organizationID, _, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	pluginName := d.Get("plugin_name").(string)
	in := &kafkaconnectcustomplugin.KafkaConnectCustomPluginUpdateIn{}
	if v, ok := d.GetOk("plugin_description"); ok {
		s := v.(string)
		in.PluginDescription = &s
	}

	if _, err := client.KafkaConnectCustomPluginUpdate(ctx, organizationID, pluginName, in); err != nil {
		return fmt.Errorf("failed to update plugin description: %w", err)
	}

	return kafkaConnectCustomPluginFileRead(ctx, d, client)
}

func kafkaConnectCustomPluginFileDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	organizationID, pluginFileID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	return common.OmitNotFound(client.KafkaConnectCustomPluginFileDelete(ctx, organizationID, pluginFileID))
}

// setCustomPluginFileData populates schema.ResourceData from a KafkaConnectCustomPluginFileGetOut response.
func setCustomPluginFileData(d *schema.ResourceData, organizationID string, info *kafkaconnectcustomplugin.KafkaConnectCustomPluginFileGetOut) error {
	pairs := map[string]any{
		"organization_id":  organizationID,
		"plugin_file_id":   info.PluginFileId,
		"plugin_id":        info.PluginId,
		"plugin_name":      info.PluginName,
		"plugin_version":   info.PluginVersion,
		"file_status":      string(info.FileStatus),
		"created_by":       info.CreatedBy,
		"created_at":       info.CreatedAt.String(),
		"updated_at":       info.UpdatedAt.String(),
		"source_checksum":  strOrEmpty(info.FileSha256),
		"file_sha256":      strOrEmpty(info.FileSha256),
		"file_description": strOrEmpty(info.FileDescription),
		"updated_by":       strOrEmpty(info.UpdatedBy),
	}
	if info.FileSize != nil {
		pairs["file_size"] = *info.FileSize
	}
	if info.VerifyErrorCode != nil {
		pairs["verify_error_code"] = *info.VerifyErrorCode
	}
	if info.VerifyErrorMessage != nil {
		pairs["verify_error_message"] = *info.VerifyErrorMessage
	}

	for k, v := range pairs {
		if err := d.Set(k, v); err != nil {
			return fmt.Errorf("failed to set %q: %w", k, err)
		}
	}
	return nil
}

func strOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// pluginFileChecksum returns the hex-encoded SHA-256 of the file at path.
func pluginFileChecksum(filePath string) (string, error) {
	f, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// uploadPluginFile performs a presigned-URL PUT upload of a local file to S3.
func uploadPluginFile(ctx context.Context, sourcePath, sourceChecksum, contentType, presignedURL string) error {
	file, err := os.Open(filepath.Clean(sourcePath)) // #nosec G304 -- user-supplied path is intentional
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	if stat.IsDir() {
		return fmt.Errorf("source path is a directory, not a file")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, presignedURL, file)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Content-SHA256", sourceChecksum)
	req.ContentLength = stat.Size()

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload request failed: %w", err)
	}
	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return fmt.Errorf("failed to read upload response: %w", err)
	}
	if len(body) > 0 {
		return fmt.Errorf("s3 upload error: %s", body)
	}
	return nil
}
