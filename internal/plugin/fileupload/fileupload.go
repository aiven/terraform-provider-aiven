// Package fileupload hashes a local file and uploads it to a pre-signed URL, shared by
// resources whose API only exposes an upload URL.
package fileupload

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

// ModifyPlan hashes the file at sourceField and marks the resource for replacement when the
// content differs from the sha stored under shaField. It writes the computed hash back to
// shaField so the plan diff shows the change before apply.
//
// Both fields must be top-level string attributes. If the backend only exposes the hash nested
// inside a block, mirror it to a flat computed attribute during Read.
func ModifyPlan(d adapter.ResourceData, sourceField, shaField string) error {
	source, _ := d.Get(sourceField).(string)

	// Hash a known source; leave empty for Unknown (a terraform_data output, etc.). Upload
	// hashes the resolved file during apply.
	var checksum string
	if source != "" {
		h, err := Checksum(source)
		if err != nil {
			return err
		}
		checksum = h
	}

	// For an existing resource, mark replace when source is Unknown (pre-empts Terraform's
	// "Provider produced inconsistent final plan" once apply resolves to a different hash) or
	// when the local content differs from state. An empty prev also triggers replace: with
	// -refresh=false after a schema upgrade (e.g. v4.63.0 renaming source_checksum to
	// shaField), Read does not run to mirror the backend hash, and the resource has no Update
	// handler, so skipping replace here would silently leave the old file in place after a
	// local edit. Mark both fields: Terraform core only acts on RequiresReplace entries whose
	// planned value actually changes, and depending on the scenario either source (unknown vs
	// known) or the hash (new vs old) is what changes.
	if !d.IsNewResource() {
		prev, _ := d.GetState(shaField).(string)
		if checksum == "" || checksum != prev {
			d.RequiresReplace(sourceField, shaField)
		}
	}

	// Unknown source has no hash to write; leave shaField as "known after apply".
	if checksum == "" {
		return nil
	}
	return d.Set(shaField, checksum)
}

// Checksum returns the hex-encoded sha256 of the file at path.
func Checksum(path string) (string, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// Upload PUTs the file at sourcePath to the pre-signed URL, sending the checksum in the
// Content-SHA256 header so the backend can verify it. If checksum is empty, it is computed
// from the file.
func Upload(ctx context.Context, sourcePath, checksum, contentType, url string) error {
	if checksum == "" {
		var err error
		checksum, err = Checksum(sourcePath)
		if err != nil {
			return err
		}
	}

	file, err := os.Open(filepath.Clean(sourcePath)) // #nosec G304 -- caller-supplied path is intentional
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	if stat.IsDir() {
		return fmt.Errorf("source path %q is a directory, not a file", sourcePath)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, file)
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
