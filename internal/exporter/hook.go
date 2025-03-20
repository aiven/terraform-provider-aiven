package exporter

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// ExportHookFunc is the signature for a function that can export ResourceData
type ExportHookFunc func(d *schema.ResourceData, resourceType string) error

var (
	// exportHook holds the registered export function
	exportHook ExportHookFunc
)

// exportEnabled is set at build time //TODO: add ldflags
var exportEnabled = true //todo: Should be always false

func init() {
	if exportEnabled {
		// Register hook only if enabled
		RegisterExportHook(func(d *schema.ResourceData, resourceType string) error {
			// Export implementation
			return nil
		})
	}
}

// IsEnabled returns whether exporting is enabled in this build
func IsEnabled() bool {
	return exportEnabled
}

// RegisterExportHook registers a function to be called when exporting resources
func RegisterExportHook(hook ExportHookFunc) {
	exportHook = hook
}

// ExportResource calls the registered export hook if one exists
func ExportResource(d *schema.ResourceData, resourceType string) error {
	if exportHook != nil {
		return exportHook(d, resourceType)
	}

	return nil
}
