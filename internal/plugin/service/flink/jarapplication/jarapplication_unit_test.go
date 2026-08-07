package jarapplication

import (
	"testing"

	"github.com/aiven/go-client-codegen/handler/flinkjarapplication"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

const (
	testProject     = "project"
	testServiceName = "flink"
	testAppID       = "00000000-0000-0000-0000-000000000001"
	testVersionID   = "00000000-0000-0000-0000-000000000002"
)

func testResponse() *flinkjarapplication.ServiceFlinkGetJarApplicationOut {
	return &flinkjarapplication.ServiceFlinkGetJarApplicationOut{
		Id:   testAppID,
		Name: "my-app",
		ApplicationVersions: []flinkjarapplication.ApplicationVersionOut{{
			Id:      testVersionID,
			Version: 1,
			FileInfo: &flinkjarapplication.FileInfoOut{
				FileStatus: flinkjarapplication.FileStatusTypeReady,
			},
		}},
		CurrentDeployment: &flinkjarapplication.CurrentDeploymentOut{
			Id:        "00000000-0000-0000-0000-000000000003",
			VersionId: testVersionID,
			Status:    flinkjarapplication.CurrentDeploymentStatusTypeRunning,
		},
	}
}

func flattenResponse(t *testing.T, state map[string]any) adapter.ResourceData {
	t.Helper()

	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		idFields(),
		adapter.WithTestState(state),
	)
	require.NoError(t, err)

	err = d.Flatten(testResponse(), adapter.RenameFields(map[string]string{"id": "application_id"}))
	require.NoError(t, err)
	return d
}

// The versions and the deployment are computed attributes, not blocks planned from the
// configuration, so they store what the API reports.
func TestFlattenStoresReadOnlyAttributes(t *testing.T) {
	t.Parallel()

	d := flattenResponse(t, map[string]any{
		"project":        testProject,
		"service_name":   testServiceName,
		"application_id": testAppID,
	})

	versions, ok := d.GetOk("application_versions")
	require.True(t, ok, "application_versions must be stored")
	require.Len(t, versions, 1)

	version := versions.([]any)[0].(map[string]any)
	require.Equal(t, testVersionID, version["id"])
	require.Equal(t, 1, version["version"])

	fileInfo := version["file_info"].([]any)[0].(map[string]any)
	require.Equal(t, "READY", fileInfo["file_status"])

	deployment, ok := d.GetOk("current_deployment")
	require.True(t, ok, "current_deployment must be stored")
	require.Equal(t, testVersionID, deployment.([]any)[0].(map[string]any)["version_id"])
}

// SDKv2 stored an empty object for a missing deployment. A refresh replaces it.
func TestFlattenReplacesStateWrittenBySDKv2(t *testing.T) {
	t.Parallel()

	d := flattenResponse(t, map[string]any{
		"project":        testProject,
		"service_name":   testServiceName,
		"application_id": testAppID,
		"application_versions": []any{
			map[string]any{"id": "stale000-0000-0000-0000-000000000000", "version": 42},
		},
		"current_deployment": []any{map[string]any{}},
	})

	version := d.Get("application_versions").([]any)[0].(map[string]any)
	require.Equal(t, testVersionID, version["id"])
	require.Equal(t, 1, version["version"])

	deployment := d.Get("current_deployment").([]any)[0].(map[string]any)
	require.Equal(t, "RUNNING", deployment["status"])
}
