package custompluginfile

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkaconnectcustomplugin"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

// sourceChecksumField holds the sha256 of the local plugin file and drives replacement:
// the uploaded file is immutable, so a changed file needs a new plugin file record.
const sourceChecksumField = "source_checksum"

func init() {
	ResourceOptions.Create = createView
	ResourceOptions.Read = readViewChecksum
}

// modifyPlan hashes the local file, which lives outside Terraform: without the hash in state
// nothing in the plan reflects an edit to the file.
func modifyPlan(_ context.Context, _ avngen.Client, d adapter.ResourceData) error {
	source, ok := d.GetOk("source")
	if !ok {
		if !d.IsNewResource() {
			d.RequiresReplace(sourceChecksumField)
		}
		return nil
	}

	checksum, err := fileChecksum(source.(string))
	if err != nil {
		return err
	}

	if !d.IsNewResource() && checksum != d.GetState(sourceChecksumField) {
		d.RequiresReplace(sourceChecksumField)
	}

	return d.Set(sourceChecksumField, checksum)
}

// createOnlyDefaults are accepted on create but never returned by the API. Without them, an
// imported file has empty values that differ from the schema defaults, and those attributes
// force replacement.
var createOnlyDefaults = map[string]string{
	"service_type": "kafka_connect",
	"content_type": "application/java-archive",
}

// readViewChecksum stores the hash of the uploaded file, which the API owns.
// An imported file has no other way to tell whether the local file still matches,
// and would otherwise be planned for replacement.
func readViewChecksum(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	if err := readView(ctx, client, d); err != nil {
		return err
	}

	for key, value := range createOnlyDefaults {
		if current, _ := d.Get(key).(string); current == "" {
			if err := d.Set(key, value); err != nil {
				return err
			}
		}
	}

	// The backend hasn't hashed the upload yet.
	checksum, _ := d.Get("file_sha256").(string)
	if checksum == "" {
		return nil
	}

	return d.Set(sourceChecksumField, checksum)
}

// createView creates the plugin file record and uploads the file to the pre-signed URL the
// API returns. The framework then polls readView until file_status reaches READY or FAILED.
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
		return fmt.Errorf("API returned no upload URL; cannot proceed with file upload")
	}

	// plugin_file_id must be in state before uploadFile can fail into deleteView, and before
	// the framework's post-create refresh can read the record back.
	if err := d.Set("plugin_file_id", rsp.PluginFileId); err != nil {
		return err
	}

	checksum := d.Get(sourceChecksumField).(string)
	if err := uploadFile(ctx, d.Get("source").(string), checksum, contentType, *rsp.UploadInfo.Url); err != nil {
		return errors.Join(err, deleteView(ctx, client, d))
	}

	return nil
}

// fileChecksum returns the hex-encoded sha256 of the file at path.
func fileChecksum(path string) (string, error) {
	f, err := os.Open(filepath.Clean(path))
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

// uploadFile performs a presigned-URL PUT upload of the local file to S3.
func uploadFile(ctx context.Context, sourcePath, checksum, contentType, presignedURL string) error {
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
	req.Header.Set("Content-SHA256", checksum)
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

	// A rejected upload doesn't always come with a body, so the status decides.
	if rsp.StatusCode < http.StatusOK || rsp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("s3 upload error: %s: %q", rsp.Status, body)
	}

	if len(body) > 0 {
		return fmt.Errorf("s3 upload error: %s", body)
	}
	return nil
}
