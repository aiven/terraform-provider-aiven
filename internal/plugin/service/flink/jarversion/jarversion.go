package jarversion

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
	"github.com/aiven/go-client-codegen/handler/flinkjarapplicationversion"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

// sourceChecksumField holds the sha256 of the uploaded jar file and drives the resource replacement.
const sourceChecksumField = "source_checksum"

func init() {
	ResourceOptions.Create = createViewUploadJar
	ResourceOptions.Read = readViewChecksum
	ResourceOptions.RefreshStateCheck = fileIsReady
}

// modifyPlan hashes the jar file, which lives outside Terraform: without the hash in state nothing
// in the plan reflects an edit to the file. A version is immutable, so new content needs a new one.
func modifyPlan(_ context.Context, _ avngen.Client, d adapter.ResourceData) error {
	source, ok := d.GetOk("source")
	if !ok {
		// The path comes from another resource and isn't known yet. An existing version
		// must already be replacement-sensitive: changing Update to Replace after the path
		// resolves during apply would produce an inconsistent final plan.
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

// createViewUploadJar uploads the jar file to the pre-signed URL that creating the version returns.
// A version without its file can't be deployed, and a failed create keeps nothing in the state,
// so the version is deleted instead of being left behind for no one to reference.
func createViewUploadJar(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	err := createView(ctx, client, d)
	if err != nil {
		return err
	}

	err = uploadJar(ctx, d)
	if err != nil {
		return errors.Join(err, deleteView(ctx, client, d))
	}

	return nil
}

// uploadJar sends the file to the pre-signed URL the create response carries.
func uploadJar(ctx context.Context, d adapter.ResourceData) error {
	url, _ := fileInfo(d)["url"].(string)
	if url == "" {
		return fmt.Errorf("jar application version created without an upload url")
	}

	return uploadFile(ctx, d.Get("source").(string), d.Get(sourceChecksumField).(string), url)
}

// readViewChecksum stores the hash of the uploaded file, which the API owns.
// An imported version has no other way to tell whether the local file still matches.
func readViewChecksum(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	err := readView(ctx, client, d)
	if err != nil {
		// The API rejects the read with a 409 until it has processed the upload. Report it as
		// "not converged yet" so the refresh after create keeps polling instead of failing.
		var apiErr avngen.Error
		if errors.As(err, &apiErr) && apiErr.Status == http.StatusConflict {
			return fmt.Errorf("%w: %s", adapter.ErrRefreshStateDesired, apiErr.Message)
		}

		return err
	}

	checksum, _ := fileInfo(d)["file_sha256"].(string)
	if checksum == "" {
		// The backend hasn't hashed the upload yet.
		return nil
	}

	return d.Set(sourceChecksumField, checksum)
}

// fileIsReady keeps the read after create polling until the backend has verified the upload.
func fileIsReady(d adapter.ResourceData) error {
	info := fileInfo(d)
	status, _ := info["file_status"].(string)

	switch flinkjarapplicationversion.FileStatusType(status) {
	case flinkjarapplicationversion.FileStatusTypeReady:
		return nil
	case flinkjarapplicationversion.FileStatusTypeFailed:
		message, _ := info["verify_error_message"].(string)
		return fmt.Errorf("%w: jar file verification failed: %s", adapter.ErrRefreshStateFailed, message)
	}

	return fmt.Errorf("jar file status is not ready: %q", status)
}

// fileInfo returns the single file_info element, or nil when the API hasn't reported one.
func fileInfo(d adapter.ResourceData) map[string]any {
	list, _ := d.Get("file_info").([]any)
	if len(list) == 0 {
		return nil
	}

	info, _ := list[0].(map[string]any)
	return info
}

func uploadFile(ctx context.Context, sourcePath, checksum, url string) error {
	file, err := os.Open(filepath.Clean(sourcePath))
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}

	if stat.IsDir() {
		return fmt.Errorf("%q is a directory", sourcePath)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, file)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/java-archive")
	req.Header.Set("Content-SHA256", checksum)
	req.ContentLength = stat.Size()

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

	// A rejected upload doesn't always come with a body, so the status decides.
	if rsp.StatusCode < http.StatusOK || rsp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("s3 error: %s: %q", rsp.Status, b)
	}

	if len(b) > 0 {
		// This is an API error
		return fmt.Errorf("s3 error: %s", b)
	}

	return nil
}

func fileChecksum(sourcePath string) (string, error) {
	file, err := os.Open(filepath.Clean(sourcePath))
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	h := sha256.New()
	_, err = io.Copy(h, file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
