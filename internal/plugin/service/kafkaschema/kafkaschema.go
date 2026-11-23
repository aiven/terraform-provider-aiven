// Package kafkaschema implements the aiven_kafka_schema resource and data source.
//
// A schema is a subject version plus the subject's compatibility level. The API may format the
// schema differently from the configuration, so the state keeps the configured value while
// the two are equivalent.
package kafkaschema

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkaschemaregistry"
	"github.com/hamba/avro/v2"
	"github.com/samber/lo"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func init() {
	DataSourceOptions.Read = readDataSource
}

var (
	whitespaceRegExp       = regexp.MustCompile(`\s+`)
	punctuationSpaceRegExp = regexp.MustCompile(`\s*([^\w\s])\s*`)
)

// equalSchemas reports whether the schemas are the same JSON, or the same Protobuf up to formatting.
func equalSchemas(a, b string) bool {
	var objA, objB any
	if json.Unmarshal([]byte(a), &objA) == nil && json.Unmarshal([]byte(b), &objB) == nil {
		return reflect.DeepEqual(objA, objB)
	}
	return normalizeProtobuf(a) == normalizeProtobuf(b)
}

// normalizeProtobuf collapses whitespace and removes it around punctuation, so formatting
// doesn't matter but "string f" and "stringf" differ.
func normalizeProtobuf(s string) string {
	s = whitespaceRegExp.ReplaceAllString(strings.TrimSpace(s), " ")
	return punctuationSpaceRegExp.ReplaceAllString(s, "$1")
}

// expandReferences returns the references sorted, so sets compare regardless of order.
func expandReferences(v any) ([]kafkaschemaregistry.ReferenceIn, error) {
	var refs []kafkaschemaregistry.ReferenceIn
	if err := schemautil.Remarshal(v, &refs); err != nil {
		return nil, err
	}
	slices.SortFunc(refs, func(a, b kafkaschemaregistry.ReferenceIn) int {
		return cmp.Or(
			cmp.Compare(a.Name, b.Name),
			cmp.Compare(a.Subject, b.Subject),
			cmp.Compare(a.Version, b.Version),
		)
	})
	return refs, nil
}

// schemaChanged reports whether the plan needs a new schema version.
// A schema that differs from the state only in formatting is stored without an API call.
// Reference values unknown at plan time are zero, so they never equal the state.
func schemaChanged(d adapter.ResourceData) (bool, error) {
	stateSchema, _ := d.GetStateOk("schema")
	if !equalSchemas(stateSchema.(string), d.Get("schema").(string)) {
		return true, nil
	}
	stateRefsValue, _ := d.GetStateOk("references")
	stateRefs, err := expandReferences(stateRefsValue)
	if err != nil {
		return false, err
	}
	planRefs, err := expandReferences(d.Get("references"))
	if err != nil {
		return false, err
	}
	return !slices.Equal(stateRefs, planRefs), nil
}

func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	if err := postSchema(ctx, client, d); err != nil {
		return err
	}
	if err := putCompatibility(ctx, client, d); err != nil {
		return err
	}
	return d.SetID(
		d.Get("project").(string),
		d.Get("service_name").(string),
		d.Get("subject_name").(string),
	)
}

func updateView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	changed, err := schemaChanged(d)
	if err != nil {
		return err
	}
	if changed {
		if err := postSchema(ctx, client, d); err != nil {
			return err
		}
	}

	if d.HasChange("compatibility_level") {
		changed = true
		if err := putCompatibility(ctx, client, d); err != nil {
			return err
		}
	}

	if !changed {
		return adapter.ErrNoChanges
	}
	return nil
}

// postSchema registers the schema as a new version, or gets the existing version of the same schema.
func postSchema(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	subjectName := d.Get("subject_name").(string)

	req := &kafkaschemaregistry.ServiceSchemaRegistrySubjectVersionPostIn{
		Schema:     d.Get("schema").(string),
		SchemaType: kafkaschemaregistry.SchemaType(d.Get("schema_type").(string)),
	}
	refs, err := expandReferences(d.Get("references"))
	if err != nil {
		return err
	}
	if len(refs) > 0 {
		req.References = &refs
	}

	// Returns the schema ID, not its version.
	schemaID, err := client.ServiceSchemaRegistrySubjectVersionPost(ctx, project, serviceName, subjectName, req)
	if err != nil {
		return fmt.Errorf("unable to add schema: %w", err)
	}

	version, err := getSchemaVersion(ctx, client, project, serviceName, subjectName, schemaID)
	if err != nil {
		return fmt.Errorf("unable to get schema version: %w", err)
	}
	return d.Set("version", version)
}

// putCompatibility sets the configured compatibility level. Removing it from the configuration keeps the current one.
func putCompatibility(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	compatibility := d.Get("compatibility_level").(string)
	if compatibility == "" {
		return nil
	}

	_, err := client.ServiceSchemaRegistrySubjectConfigPut(
		ctx,
		d.Get("project").(string),
		d.Get("service_name").(string),
		d.Get("subject_name").(string),
		&kafkaschemaregistry.ServiceSchemaRegistrySubjectConfigPutIn{
			Compatibility: kafkaschemaregistry.CompatibilityType(compatibility),
		},
	)
	if err != nil {
		return fmt.Errorf("unable to update configuration: %w", err)
	}
	return nil
}

// getSchemaVersion polls until the schema ID appears in the subject's versions. A new schema is
// usually the latest version, so they're checked newest first.
func getSchemaVersion(ctx context.Context, client avngen.Client, project, serviceName, subjectName string, id int) (int, error) {
	for {
		versions, err := client.ServiceSchemaRegistrySubjectVersionsGet(ctx, project, serviceName, subjectName)
		// The registry is eventually consistent: the subject may not be visible right after creation.
		if err != nil && !avngen.IsNotFound(err) {
			return 0, err
		}

		slices.Sort(versions)
		for _, v := range slices.Backward(versions) {
			s, err := client.ServiceSchemaRegistrySubjectVersionGet(ctx, project, serviceName, subjectName, v)
			if avngen.IsNotFound(err) {
				continue
			}
			if err != nil {
				return 0, err
			}
			if s.Id == id {
				return s.Version, nil
			}
		}

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

func readDataSource(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	subjectName := d.Get("subject_name").(string)

	err := readView(ctx, client, d)
	if avngen.IsNotFound(err) {
		return fmt.Errorf("kafka schema subject %s/%s/%s not found: %w", project, serviceName, subjectName, err)
	}
	return err
}

func readView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	subjectName := d.Get("subject_name").(string)

	svc, err := schemautil.GetService(ctx, client, project, serviceName)
	if err != nil {
		return err
	}
	if svc.UserConfig["schema_registry"] != true {
		return fmt.Errorf("%w: schema registry is disabled", adapter.ErrNotFound)
	}

	version := d.Get("version").(int)
	if version == 0 {
		// Import and the data source read the latest version.
		versions, err := client.ServiceSchemaRegistrySubjectVersionsGet(ctx, project, serviceName, subjectName)
		if err != nil {
			return err
		}
		if len(versions) == 0 {
			return fmt.Errorf("%w: subject %q has no versions", adapter.ErrNotFound, subjectName)
		}
		version = slices.Max(versions)
	}

	s, err := client.ServiceSchemaRegistrySubjectVersionGet(ctx, project, serviceName, subjectName, version)
	if err != nil {
		return err
	}

	compatibility, err := client.ServiceSchemaRegistrySubjectConfigGet(
		ctx,
		project,
		serviceName,
		subjectName,
		kafkaschemaregistry.ServiceSchemaRegistrySubjectConfigGetGlobalDefaultFallback(false),
	)
	if err != nil && !avngen.IsNotFound(err) {
		return err
	}

	refs := lo.Map(s.References, func(r kafkaschemaregistry.ReferenceOut, _ int) map[string]any {
		return map[string]any{"name": r.Name, "subject": r.Subject, "version": r.Version}
	})

	values := map[string]any{
		"project":             project,
		"service_name":        serviceName,
		"subject_name":        subjectName,
		"version":             s.Version,
		"schema_type":         string(cmp.Or(s.SchemaType, kafkaschemaregistry.SchemaTypeAvro)),
		"references":          refs,
		"compatibility_level": string(compatibility),
	}
	if !equalSchemas(d.Get("schema").(string), s.Schema) {
		values["schema"] = s.Schema
	}

	for k, v := range values {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}
	return d.SetID(project, serviceName, subjectName)
}

// deleteView soft deletes the subject. A subject with references is hard deleted too,
// so the schemas it references can be deleted. Hard deletion requires a soft deletion first.
// Schemas without references stay soft deleted, as before hard deletion was introduced.
func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	subjectName := d.Get("subject_name").(string)

	svc, err := schemautil.GetService(ctx, client, project, serviceName)
	if err != nil {
		return err
	}
	if svc.UserConfig["schema_registry"] != true {
		return nil
	}

	hasReferences, err := hasAnyVersionReferences(ctx, client, project, serviceName, subjectName)
	if err != nil {
		return err
	}

	for _, permanent := range []bool{false, true} {
		if permanent && !hasReferences {
			break
		}

		err = client.ServiceSchemaRegistrySubjectDelete(
			ctx,
			project,
			serviceName,
			subjectName,
			kafkaschemaregistry.ServiceSchemaRegistrySubjectDeletePermanent(permanent),
		)
		if common.IsCritical(err) {
			return err
		}
	}
	return nil
}

func hasAnyVersionReferences(ctx context.Context, client avngen.Client, project, serviceName, subjectName string) (bool, error) {
	versions, err := client.ServiceSchemaRegistrySubjectVersionsGet(ctx, project, serviceName, subjectName)
	if avngen.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	for _, version := range versions {
		s, err := client.ServiceSchemaRegistrySubjectVersionGet(ctx, project, serviceName, subjectName, version)
		if avngen.IsNotFound(err) {
			continue
		}
		if err != nil {
			return false, err
		}
		if len(s.References) > 0 {
			return true, nil
		}
	}
	return false, nil
}

// modifyPlan replaces the resource for a new type, otherwise validates the schema and checks it's
// compatible with the current version.
func modifyPlan(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	// A new type is rarely compatible with the current version, and updateView registers a new
	// version only for a schema change. An empty state is an SDKv2 import, which never read the type.
	if old, _ := d.GetStateOk("schema_type"); old != "" && d.HasChange("schema_type") {
		d.RequiresReplace("schema_type")
		return nil
	}

	schemaPayload := d.Get("schema").(string)
	if schemaPayload == "" {
		// Unknown until apply, and so is the version.
		return nil
	}

	stateVersion, hasVersion := d.GetStateOk("version")
	if hasVersion {
		changed, err := schemaChanged(d)
		if err != nil {
			return err
		}
		if !changed {
			// Otherwise the version stays unknown: the update registers a new one.
			return d.Set("version", stateVersion)
		}
	}

	// The API can't check the compatibility of a schema with references,
	// and avro.Parse can't resolve the types they refer to.
	if refs, _ := d.Get("references").([]any); len(refs) > 0 {
		return nil
	}

	// Unknown on create without a configured type, which the API defaults to AVRO.
	schemaType := cmp.Or(kafkaschemaregistry.SchemaType(d.Get("schema_type").(string)), kafkaschemaregistry.SchemaTypeAvro)
	if schemaType == kafkaschemaregistry.SchemaTypeAvro {
		if _, err := avro.Parse(schemaPayload); err != nil {
			return fmt.Errorf("schema validation error: %w", err)
		}
	}

	if !hasVersion {
		return nil
	}

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	subjectName := d.Get("subject_name").(string)
	version := stateVersion.(int)
	r, err := client.ServiceSchemaRegistryCompatibility(
		ctx,
		project,
		serviceName,
		subjectName,
		version,
		&kafkaschemaregistry.ServiceSchemaRegistryCompatibilityIn{
			Schema:     schemaPayload,
			SchemaType: schemaType,
		},
	)
	if err != nil {
		svc, checkErr := schemautil.GetService(ctx, client, project, serviceName)
		if checkErr != nil {
			return fmt.Errorf("unable to check schema registry status: %w", checkErr)
		}
		if svc.UserConfig["schema_registry"] != true {
			return nil
		}
		return fmt.Errorf("unable to check schema validity: %w", err)
	}

	if !r.IsCompatible {
		return fmt.Errorf(
			"subject %q: schema is not compatible with previous version %d: %s",
			subjectName,
			version,
			strings.Join(r.Messages, ", "),
		)
	}
	return nil
}
