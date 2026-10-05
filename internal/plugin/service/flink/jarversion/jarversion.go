package jarversion

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/flinkjarapplicationversion"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/fileupload"
)

const fileSha256Field = "file_sha256"

func init() {
	ResourceOptions.Create = createViewUploadJar
	ResourceOptions.Read = readViewMirrorHash
	ResourceOptions.RefreshStateCheck = fileIsReady
}

// modifyPlan hashes the local jar and marks the resource for replacement when the content
// differs from the backend-stored hash.
func modifyPlan(_ context.Context, _ avngen.Client, d adapter.ResourceData) error {
	return fileupload.ModifyPlan(d, "source", fileSha256Field)
}

// createViewUploadJar creates the version then uploads the jar; a failed upload deletes
// the orphaned version so nothing references a jar that was never stored.
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

func uploadJar(ctx context.Context, d adapter.ResourceData) error {
	url, _ := fileInfo(d)["url"].(string)
	if url == "" {
		return fmt.Errorf("jar application version created without an upload url")
	}

	// modifyPlan already hashed the file, so reuse that value from state.
	checksum, _ := d.Get(fileSha256Field).(string)
	return fileupload.Upload(ctx, d.Get("source").(string), checksum, "application/java-archive", url)
}

// readViewMirrorHash reports "not converged yet" on the 409 the API returns until the upload
// is processed, so the post-create refresh keeps polling instead of failing. It also mirrors
// file_info.0.file_sha256 to the flat file_sha256 attribute so an imported resource exposes
// the backend-computed hash; modifyPlan's locally computed value, when present, wins.
func readViewMirrorHash(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	err := readView(ctx, client, d)
	if err != nil {
		var apiErr avngen.Error
		if errors.As(err, &apiErr) && apiErr.Status == http.StatusConflict {
			return fmt.Errorf("%w: %s", adapter.ErrRefreshStateDesired, apiErr.Message)
		}
		return err
	}

	// Only mirror when modifyPlan hasn't written a hash. Backend-computed file_sha256 can differ
	// from the raw-bytes sha256 we compute (e.g. the backend may hash canonical ZIP content), so
	// overwriting here would make the post-create apply result diverge from the plan.
	if sha, _ := d.Get(fileSha256Field).(string); sha != "" {
		return nil
	}
	sha, _ := fileInfo(d)["file_sha256"].(string)
	if sha == "" {
		return nil
	}
	return d.Set(fileSha256Field, sha)
}

// fileIsReady keeps the post-create refresh polling until the backend verifies the upload and
// reports its sha256.
func fileIsReady(d adapter.ResourceData) error {
	info := fileInfo(d)
	status, _ := info["file_status"].(string)

	switch flinkjarapplicationversion.FileStatusType(status) {
	case flinkjarapplicationversion.FileStatusTypeFailed:
		message, _ := info["verify_error_message"].(string)
		return fmt.Errorf("%w: jar file verification failed: %s", adapter.ErrRefreshStateFailed, message)
	case flinkjarapplicationversion.FileStatusTypeReady:
		if sha, _ := info["file_sha256"].(string); sha != "" {
			return nil
		}
	}

	return fmt.Errorf("jar file status is not ready: %q", status)
}

func fileInfo(d adapter.ResourceData) map[string]any {
	list, _ := d.Get("file_info").([]any)
	if len(list) == 0 {
		return nil
	}

	info, _ := list[0].(map[string]any)
	return info
}
