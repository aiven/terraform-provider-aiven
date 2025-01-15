package flink

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
	"github.com/aiven/go-client-codegen/handler/flinkjarapplicationversion"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func ResourceFlinkJarApplicationVersion() *schema.Resource {
	s := flinkJarApplicationVersionSchema()
	s["source"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The path to the jar file to upload.",
		// Terraform requires this to be set to true.
		// Because the resource cannot be updated.
		ForceNew: true,
		DiffSuppressFunc: func(_, _, _ string, d *schema.ResourceData) bool {
			// Ignores file renames.
			// The checksum is used to detect changes.
			// Doesn't suppress the diff for new resources.
			return d.Id() != ""
		},
	}

	s["source_checksum"] = &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		ForceNew:    true,
		Description: "The sha256 checksum of the jar file to upload.",
	}

	return &schema.Resource{
		Description:   "Creates and manages an Aiven for Apache Flink® jar application version.",
		CreateContext: common.WithGenClient(flinkJarApplicationVersionCreate),
		ReadContext:   common.WithGenClient(flinkJarApplicationVersionRead),
		DeleteContext: common.WithGenClient(flinkJarApplicationVersionDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   s,
		CustomizeDiff: func(_ context.Context, diff *schema.ResourceDiff, _ any) error {
			sourcePath := diff.Get("source").(string)
			checksum, err := filePathChecksum(sourcePath)
			if err != nil {
				return fmt.Errorf("failed to calculate checksum: %w", err)
			}
			return diff.SetNew("source_checksum", checksum)
		},
	}
}

func flinkJarApplicationVersionCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	applicationID := d.Get("application_id").(string)

	rsp, err := client.ServiceFlinkCreateJarApplicationVersion(ctx, project, serviceName, applicationID)
	if err != nil {
		return err
	}

	sourcePath := d.Get("source").(string)
	sourceChecksum := d.Get("source_checksum").(string)
	err = uploadJarFile(ctx, sourcePath, sourceChecksum, *rsp.FileInfo.Url)
	if err != nil {
		return fmt.Errorf("failed to upload jar file: %w", err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, applicationID, rsp.Id))

	// Waits until the file is uploaded.
	// Retries until the context is canceled or the file is ready.
	err = retry.Do(
		func() error {
			v, err := client.ServiceFlinkGetJarApplicationVersion(ctx, project, serviceName, applicationID, rsp.Id)
			switch {
			case avngen.IsNotFound(err):
				// 404 means something went completely wrong. Retrying won't help
				return retry.Unrecoverable(err)
			case err != nil:
				// The rest is retryable
				return err
			case v.FileInfo == nil:
				// Not sure if this is possible. File info should always be present
				return fmt.Errorf("file status is not ready")
			case v.FileInfo.FileStatus != flinkjarapplicationversion.FileStatusTypeReady:
				return fmt.Errorf("file status is not ready: %q", v.FileInfo.FileStatus)
			}

			// Nothing to retry
			return nil
		},
		retry.Context(ctx),
		retry.Delay(time.Second*5),
	)

	if err != nil {
		return fmt.Errorf("failed to wait for jar application version: %w", err)
	}

	return flinkJarApplicationVersionRead(ctx, d, client)
}

func flinkJarApplicationVersionRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	project, serviceName, applicationID, version, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return err
	}

	rsp, err := client.ServiceFlinkGetJarApplicationVersion(ctx, project, serviceName, applicationID, version)
	if err != nil {
		return schemautil.ResourceReadHandleNotFound(err, d)
	}

	// This is for import. Triggers change detection.
	err = d.Set("source_checksum", rsp.FileInfo.FileSha256)
	if err != nil {
		return err
	}

	return schemautil.ResourceDataSet(
		flinkJarApplicationVersionSchema(), d, rsp,
		schemautil.RenameAliasesReverse(flinkJarApplicationVersionRename()),
	)
}

func flinkJarApplicationVersionDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	project, serviceName, applicationID, version, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return err
	}

	_, err = client.ServiceFlinkDeleteJarApplicationVersion(ctx, project, serviceName, applicationID, version)
	return schemautil.OmitNotFound(err)
}

func uploadJarFile(ctx context.Context, sourcePath, sourceChecksum, urlPath string) error {
	file, err := os.Open(filepath.Clean(sourcePath))
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	size, err := fileSize(file)
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, urlPath, file)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/java-archive")
	req.Header.Set("Content-SHA256", sourceChecksum)
	req.ContentLength = size

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer rsp.Body.Close()

	b, err := io.ReadAll(rsp.Body)
	if err != nil {
		// This is a connection error or something else, not an API error
		return fmt.Errorf("failed to read response: %w", err)
	}

	if len(b) > 0 {
		// This is an API error
		return fmt.Errorf("s3 error: %s", b)
	}
	return nil
}

func fileSize(file *os.File) (int64, error) {
	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}

	if stat.IsDir() {
		return 0, fmt.Errorf("file is a directory")
	}

	return stat.Size(), nil
}

func filePathChecksum(filePath string) (string, error) {
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return "", err
	}
	defer file.Close()
	return fileChecksum(file)
}

func fileChecksum(file *os.File) (string, error) {
	h := sha256.New()
	_, err := io.Copy(h, file)
	if err != nil {
		return "", err
	}
	s := fmt.Sprintf("%x", h.Sum(nil))
	_, err = file.Seek(0, 0)
	if err != nil {
		return "", err
	}
	return s, nil
}
