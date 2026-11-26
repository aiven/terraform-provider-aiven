// Package kafkaschema implements the aiven_kafka_schema resource and data source.
//
// A schema is a subject version plus the subject's compatibility level. The API may format the
// schema differently from the configuration, so the state keeps the configured value while
// the two are equivalent.
package kafkaschema

import (
	"bytes"
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkaschemaregistry"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hamba/avro/v2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func init() {
	DataSourceOptions.Read = readDataSource
	ResourceOptions.CreateIsUpsert = true
	ResourceOptions.ImportState = importState
}

func importState(ctx context.Context, _ avngen.Client, req resource.ImportStateRequest, rsp *resource.ImportStateResponse) {
	parts, err := schemautil.SplitResourceID(req.ID, len(ResourceOptions.IdentityFields))
	if err != nil {
		rsp.Diagnostics.AddError("Invalid Kafka schema ID", err.Error())
		return
	}
	for i, field := range ResourceOptions.IdentityFields {
		rsp.Diagnostics.Append(rsp.State.SetAttribute(ctx, path.Root(field), parts[i])...)
	}
	rsp.Diagnostics.Append(rsp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
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
// ModifyPlan defers this comparison until all reference values are known.
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

// The service may have been replaced or its registry enabled since the last provider RPC.
func kafkaService(ctx context.Context, client avngen.Client, project, name string) (*service.ServiceGetOut, error) {
	svc, err := schemautil.GetService(ctx, client, project, name)
	if err != nil {
		return nil, err
	}
	return svc, schemautil.ValidateServiceReady(svc, project, name, "kafka")
}

func checkSubjectRegistry(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project, name := d.Get("project").(string), d.Get("service_name").(string)
	svc, err := kafkaService(ctx, client, project, name)
	if d.IsResource() && avngen.IsNotFound(err) {
		// A transient parent 404 must not forget a live subject or skip its deletion.
		subject := d.Get("subject_name").(string)
		versions, listErr := client.ServiceSchemaRegistrySubjectVersionsGet(ctx, project, name, subject)
		switch {
		case listErr == nil && len(versions) > 0:
			return nil
		case listErr == nil:
			return fmt.Errorf("%w: subject %q has no versions", adapter.ErrNotFound, subject)
		case !avngen.IsNotFound(listErr):
			return listErr
		}
		// Both URLs are scoped to the same parent. Check the project list
		// before treating those two 404s as evidence that it is gone.
		services, listErr := client.ServiceList(ctx, project)
		// RemoveMissing must only receive a 404 that confirms absence.
		if avngen.IsNotFound(listErr) {
			return fmt.Errorf("unable to confirm whether service %q still exists: the service list for project %q returned HTTP 404", name, project)
		}
		if listErr != nil {
			return fmt.Errorf("unable to establish whether service %q still exists: %w", name, listErr)
		}
		if !slices.ContainsFunc(services, func(s service.ServiceOut) bool { return s.ServiceName == name }) {
			return fmt.Errorf("%w: service %q is absent from project %q", adapter.ErrNotFound, name, project)
		}
		svc, err = kafkaService(ctx, client, project, name)
		if avngen.IsNotFound(err) {
			return fmt.Errorf("service %q is listed in project %q but ServiceGet still returns HTTP 404", name, project)
		}
	}
	if d.IsResource() && errors.Is(err, schemautil.ErrServiceTypeMismatch) {
		return fmt.Errorf("%w: %w", adapter.ErrNotFound, err)
	}
	if err != nil && !errors.Is(err, schemautil.ErrServicePoweredOff) {
		return err
	}
	// The service still reports its registry setting while powered off.
	if svc.UserConfig["schema_registry"] != true {
		return fmt.Errorf("%w: schema registry is disabled", adapter.ErrNotFound)
	}
	return err
}

func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	if _, err := kafkaService(ctx, client, d.Get("project").(string), d.Get("service_name").(string)); err != nil {
		return err
	}
	if err := postSchema(ctx, client, d); err != nil {
		return err
	}
	if level := plannedCompatibility(d); level != "" {
		if err := putCompatibility(ctx, client, d, level); err != nil {
			return err
		}
	}
	// Registration is an upsert. A failed compatibility write must be retried
	// without tainting and subsequently destroying an existing subject's history.
	return d.Set("id", schemautil.BuildResourceID(d.Get("project").(string), d.Get("service_name").(string), d.Get("subject_name").(string)))
}

func updateView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	register, err := schemaChanged(d)
	if err != nil {
		return err
	}
	level := plannedCompatibility(d)
	writeLevel := level != "" && (register || d.HasChange("compatibility_level"))
	if !register && !writeLevel {
		return adapter.ErrNoChanges
	}
	if _, err := kafkaService(ctx, client, d.Get("project").(string), d.Get("service_name").(string)); err != nil {
		return err
	}
	if register {
		if err := postSchema(ctx, client, d); err != nil {
			return err
		}
	}
	if writeLevel {
		return putCompatibility(ctx, client, d, level)
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
		// SDKv2's Set.List ordered these objects by the decimal string of
		// their hash. Karapace includes this sequence in the schema fingerprint.
		slices.SortStableFunc(refs, func(a, b kafkaschemaregistry.ReferenceIn) int {
			return strings.Compare(sdkReferenceKey(a), sdkReferenceKey(b))
		})
		req.References = &refs
	}

	// Returns the schema ID, not its version.
	schemaID, err := client.ServiceSchemaRegistrySubjectVersionPost(ctx, project, serviceName, subjectName, req)
	if err != nil {
		return fmt.Errorf("unable to add schema: %w", err)
	}

	registered, err := getSchemaVersion(ctx, client, project, serviceName, subjectName, schemaID)
	if err != nil {
		return fmt.Errorf("unable to get schema version: %w", err)
	}
	if err := d.Set("version", registered.Version); err != nil {
		return err
	}
	if err := d.Set("references", flattenReferences(registered.References)); err != nil {
		return err
	}
	kind := cmp.Or(registered.SchemaType, kafkaschemaregistry.SchemaTypeAvro)
	requestedKind := cmp.Or(req.SchemaType, kafkaschemaregistry.SchemaTypeAvro)
	matches := kind == requestedKind && equalSchemas(req.Schema, registered.Schema)
	var verifyErr error
	if !matches && kind == requestedKind && kind == kafkaschemaregistry.SchemaTypeProtobuf {
		matches, verifyErr = registeredProtobufMatches(ctx, client, project, serviceName, req.Schema, registered)
	}
	if !matches {
		// Keep the actual version on partial updates so the next read can find it.
		if err := d.Set("schema", registered.Schema); err != nil {
			return err
		}
		if err := d.Set("schema_type", string(kind)); err != nil {
			return err
		}
	}
	if err := d.CheckpointUpdate("schema", "schema_type", "references", "version"); err != nil {
		return err
	}
	if verifyErr != nil {
		return fmt.Errorf("unable to verify Protobuf subject %q, version %d: %w", subjectName, registered.Version, verifyErr)
	}
	if !matches {
		return fmt.Errorf("registered %s schema differs from the submitted %s schema for subject %q, version %d", kind, requestedKind, subjectName, registered.Version)
	}
	var textPair []byte
	if len(registered.References) == 0 && req.Schema != registered.Schema {
		textPair = registrationTextPair(kind, req.Schema, registered.Schema)
	}
	if err := d.SetPrivate(ctx, registeredTextKey, textPair); err != nil {
		return err
	}
	return d.SetPrivate(ctx, registeredSchemaKey, registryRepresentation(registered))
}

func registeredProtobufMatches(ctx context.Context, client avngen.Client, project, serviceName, submitted string, registered *kafkaschemaregistry.ServiceSchemaRegistrySubjectVersionGetOut) (bool, error) {
	if len(registered.References) == 0 {
		return sameProtobufDeclarations(submitted, registered.Schema, protoregistry.GlobalFiles), nil
	}
	if sameProtobufDeclarations(submitted, registered.Schema, nil) {
		return true, nil
	}
	desc, binary := binaryDescriptor(submitted)
	if !binary {
		return false, nil
	}
	files, err := protobufReferenceFiles(ctx, client, project, serviceName, registered.References, desc.Dependency)
	if err != nil {
		return false, err
	}
	return sameProtobufDeclarations(submitted, registered.Schema, files), nil
}

// A compiled field specifies its kind. The source returned by the registry
// names the field type and may need the pinned imports to resolve that kind.
func protobufReferenceFiles(ctx context.Context, client avngen.Client, project, serviceName string, refs []kafkaschemaregistry.ReferenceOut, imports []string) (*protoregistry.Files, error) {
	pending := slices.Clone(refs)
	seen := make(map[kafkaschemaregistry.ReferenceOut]bool)
	descriptors := make(map[string]*descriptorpb.FileDescriptorProto)
	set := new(descriptorpb.FileDescriptorSet)
	imports = slices.Clone(imports)
	for i := 0; i < len(pending); i++ {
		ref := pending[i]
		if seen[ref] {
			continue
		}
		seen[ref] = true
		s, err := client.ServiceSchemaRegistrySubjectVersionGet(ctx, project, serviceName, ref.Subject, ref.Version)
		if err != nil {
			return nil, fmt.Errorf("read import %q from subject %q, version %d: %w", ref.Name, ref.Subject, ref.Version, err)
		}
		desc, _, err := registryDescriptor(s.Schema)
		if err != nil {
			return nil, fmt.Errorf("parse Protobuf import %q: %w", ref.Name, err)
		}
		desc.Name = proto.String(ref.Name)
		pending = append(pending, s.References...)
		if previous := descriptors[ref.Name]; previous != nil {
			if !proto.Equal(previous, desc) {
				return nil, fmt.Errorf("conflicting declarations for Protobuf import %q", ref.Name)
			}
			continue
		}
		descriptors[ref.Name] = desc
		set.File = append(set.File, desc)
		imports = append(imports, desc.Dependency...)
	}
	// Include the standard descriptors used by this graph.
	for i := 0; i < len(imports); i++ {
		name := imports[i]
		if _, ok := descriptors[name]; ok {
			continue
		}
		descriptors[name] = nil
		if file, err := protoregistry.GlobalFiles.FindFileByPath(name); err == nil {
			desc := protodesc.ToFileDescriptorProto(file)
			set.File = append(set.File, desc)
			imports = append(imports, desc.Dependency...)
		}
	}
	return (protodesc.FileOptions{AllowUnresolvable: true}).NewFiles(set)
}

func sdkReferenceKey(ref kafkaschemaregistry.ReferenceIn) string {
	return strconv.Itoa(sdkschema.HashString(fmt.Sprintf("name:%s;subject:%s;version:%d;", ref.Name, ref.Subject, ref.Version)))
}

func flattenReferences(refs []kafkaschemaregistry.ReferenceOut) []map[string]any {
	return lo.Map(refs, func(r kafkaschemaregistry.ReferenceOut, _ int) map[string]any {
		return map[string]any{"name": r.Name, "subject": r.Subject, "version": r.Version}
	})
}

const (
	registeredSchemaKey  = "registry_schema"
	registeredTextKey    = "registered_source_and_text"
	importedTypeKey      = "sdk_import_without_type"
	replacementSourceKey = "replacement_source"
)

func registryRepresentation(s *kafkaschemaregistry.ServiceSchemaRegistrySubjectVersionGetOut) []byte {
	return fmt.Appendf(nil, `"%d/%d/%x"`, s.Id, s.Version, sha256.Sum256([]byte(s.Schema)))
}

// The registry fingerprints its rendered source. Remember an observed rendering
// so that moving the identical request within this registry can reuse that version.
func registrationTextPair(kind kafkaschemaregistry.SchemaType, submitted, rendered string) []byte {
	return fmt.Appendf(nil, `"%s/%x/%x"`, kind, sha256.Sum256([]byte(submitted)), sha256.Sum256([]byte(rendered)))
}

func plannedCompatibility(d adapter.ResourceData) string {
	if _, configured := d.GetConfigOk("compatibility_level"); !configured {
		return ""
	}
	level, known := d.GetPlanOk("compatibility_level")
	if !known {
		return ""
	}
	return level.(string)
}

func putCompatibility(ctx context.Context, client avngen.Client, d adapter.ResourceData, compatibility string) error {
	observed, err := client.ServiceSchemaRegistrySubjectConfigPut(
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
	if string(observed) == compatibility {
		return d.CheckpointUpdate("compatibility_level")
	}
	return nil
}

// getSchemaVersion polls until the schema ID appears in the subject's versions. A new schema is
// usually the latest version, so they're checked newest first.
func getSchemaVersion(ctx context.Context, client avngen.Client, project, serviceName, subjectName string, id int) (*kafkaschemaregistry.ServiceSchemaRegistrySubjectVersionGetOut, error) {
	for {
		versions, err := client.ServiceSchemaRegistrySubjectVersionsGet(ctx, project, serviceName, subjectName)
		// The registry is eventually consistent: the subject may not be visible right after creation.
		if err != nil && !avngen.IsNotFound(err) {
			return nil, err
		}

		slices.Sort(versions)
		for _, v := range slices.Backward(versions) {
			s, err := client.ServiceSchemaRegistrySubjectVersionGet(ctx, project, serviceName, subjectName, v)
			if avngen.IsNotFound(err) {
				continue
			}
			if err != nil {
				return nil, err
			}
			if s.Id == id {
				return s, nil
			}
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

func readDataSource(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	subjectName := d.Get("subject_name").(string)

	err := readView(ctx, client, d)
	if adapter.IsNotFound(err) {
		return fmt.Errorf("kafka schema subject %s/%s/%s not found: %w", project, serviceName, subjectName, err)
	}
	return err
}

func readView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	subjectName := d.Get("subject_name").(string)

	if err := checkSubjectRegistry(ctx, client, d); err != nil {
		return err
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

	s, err := readPinnedVersion(ctx, client, project, serviceName, subjectName, version)
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

	// On a post-write read, the plan still contains the level actually sent.
	// Check it before changing any planned values so subsequent attempts use the same target.
	level := string(compatibility)
	if target := plannedCompatibility(d); target != "" {
		if level != target {
			return fmt.Errorf("%w: subject %q has compatibility %q, waiting for %q", adapter.ErrRefreshStateDesired, subjectName, level, target)
		}
	} else if planned, known := d.GetPlanOk("compatibility_level"); known {
		// ignore_changes can pin an observed level even when the configuration
		// omits it. Honor that plan during apply. Ordinary Read observes drift.
		level = planned.(string)
	}

	values := map[string]any{
		"project":             project,
		"service_name":        serviceName,
		"subject_name":        subjectName,
		"version":             s.Version,
		"schema_type":         string(cmp.Or(s.SchemaType, kafkaschemaregistry.SchemaTypeAvro)),
		"references":          flattenReferences(s.References),
		"compatibility_level": level,
	}
	representation, err := d.GetPrivate(ctx, registeredSchemaKey)
	if err != nil {
		return err
	}
	if !bytes.Equal(representation, registryRepresentation(s)) {
		if !equalSchemas(d.Get("schema").(string), s.Schema) {
			values["schema"] = s.Schema
		}
		if len(representation) > 0 {
			if err := d.SetPrivate(ctx, registeredSchemaKey, nil); err != nil {
				return err
			}
			if err := d.SetPrivate(ctx, registeredTextKey, nil); err != nil {
				return err
			}
		}
	}
	if d.IsResource() {
		// SDK import suppressed every configured type while its state had no
		// type. Preserve that provenance for the plan following this first Read.
		priorType, _ := d.GetStateOk("schema_type")
		_, hasVersion := d.GetStateOk("version")
		_, afterWrite := d.GetPlanOk("schema")
		var imported []byte
		if hasVersion && priorType == "" && !afterWrite {
			imported = []byte("true")
		}
		if err := d.SetPrivate(ctx, importedTypeKey, imported); err != nil {
			return err
		}
	}

	for k, v := range values {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}
	// SDK string import kept the spelling supplied by the caller, including
	// a literal slash in the final component. Preserve that public value.
	if id, ok := d.GetStateOk("id"); d.IsResource() && ok {
		return d.Set("id", id)
	}
	return d.Set("id", schemautil.BuildResourceID(project, serviceName, subjectName))
}

// A version GET can briefly return 404 while the version is still listed.
// Confirm absence before removing it from state, including under short read timeouts.
func readPinnedVersion(ctx context.Context, client avngen.Client, project, serviceName, subject string, version int) (*kafkaschemaregistry.ServiceSchemaRegistrySubjectVersionGetOut, error) {
	for {
		s, err := client.ServiceSchemaRegistrySubjectVersionGet(ctx, project, serviceName, subject, version)
		if !avngen.IsNotFound(err) {
			return s, err
		}
		versions, listErr := client.ServiceSchemaRegistrySubjectVersionsGet(ctx, project, serviceName, subject)
		if listErr != nil {
			return nil, listErr
		}
		if !slices.Contains(versions, version) {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

// deleteView soft deletes the subject. A subject with references is hard deleted too,
// so the schemas it references can be deleted. Hard deletion requires a soft deletion first.
// Schemas without references stay soft deleted, as before hard deletion was introduced.
func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	subjectName := d.Get("subject_name").(string)

	err := checkSubjectRegistry(ctx, client, d)
	if adapter.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
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

// An SDK import has a version but no saved type. Resolve the pinned version
// before planning a default or replacement, including with -refresh=false.
func previousSchemaType(ctx context.Context, client avngen.Client, d adapter.ResourceData, version int) (kafkaschemaregistry.SchemaType, error) {
	if kind, _ := d.GetStateOk("schema_type"); kind != "" || version == 0 {
		return kafkaschemaregistry.SchemaType(kind.(string)), nil
	}
	id := d.GetState("id").(string)
	project, name, subject, err := schemautil.SplitResourceID3(id)
	if err != nil {
		return "", err
	}
	s, err := client.ServiceSchemaRegistrySubjectVersionGet(ctx, project, name, subject, version)
	if err != nil {
		return "", fmt.Errorf("unable to read type for subject %q, version %d: %w", subject, version, err)
	}
	return cmp.Or(s.SchemaType, kafkaschemaregistry.SchemaTypeAvro), nil
}

type subjectVersion struct {
	Project, Service, Subject string
	Version                   int
}

func modifyPlan(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	stateVersion, hasVersion := d.GetStateOk("version")
	var source subjectVersion
	if hasVersion {
		source = subjectVersion{
			Project: d.GetState("project").(string), Service: d.GetState("service_name").(string),
			Subject: d.GetState("subject_name").(string), Version: stateVersion.(int),
		}
	} else {
		// Terraform replans a replacement as a Create with the original configuration,
		// null prior state and the private data returned by the first plan.
		pending, err := d.GetPrivate(ctx, replacementSourceKey)
		if err != nil {
			return err
		}
		if len(pending) > 0 {
			if err := json.Unmarshal(pending, &source); err != nil {
				return fmt.Errorf("read source of schema replacement: %w", err)
			}
			if err := d.SetPrivate(ctx, replacementSourceKey, nil); err != nil {
				return err
			}
		}
	}
	replacementCreate := !hasVersion && source.Version > 0
	newSubject := hasVersion && (d.HasChange("project") || d.HasChange("service_name") || d.HasChange("subject_name"))
	var priorType kafkaschemaregistry.SchemaType
	if !newSubject {
		// A different identity already requires replacement and Create uses its own type.
		var err error
		priorType, err = previousSchemaType(ctx, client, d, stateVersion.(int))
		if err != nil {
			return err
		}
	}
	kind, known := d.GetPlanOk("schema_type")
	unknownType := !known && d.IsConfigUnknown("schema_type")
	if unknownType && !hasVersion && !replacementCreate {
		return nil
	}
	schemaType := kafkaschemaregistry.SchemaType(kind.(string))
	if schemaType == "" && !unknownType {
		schemaType = cmp.Or(priorType, kafkaschemaregistry.SchemaTypeAvro)
		if err := d.Set("schema_type", string(schemaType)); err != nil {
			return err
		}
	}
	// On an existing subject, even a ForceNew type that is still unknown
	// must pass the SDK's pre-destroy check with the registry's default type.
	replaceType := hasVersion && !newSubject && (unknownType || priorType != schemaType)
	if replaceType {
		pending, err := d.GetPrivate(ctx, importedTypeKey)
		if err != nil {
			return err
		}
		savedType, _ := d.GetStateOk("schema_type")
		if savedType == "" || len(pending) > 0 {
			return fmt.Errorf("SDK import of subject %q did not record its type. The registry reports %s and the configuration requests %s. Run terraform apply -refresh-only to record the observed type, then review the replacement plan with refresh enabled", d.Get("subject_name"), priorType, cmp.Or(string(schemaType), "an unresolved type"))
		}
		d.RequiresReplace("schema_type")
	}
	if newSubject || replaceType {
		// ignore_changes can hide the destination, schema or references in this
		// first plan. Validate once Terraform supplies the actual Create inputs.
		pending, err := json.Marshal(source)
		if err != nil {
			return err
		}
		return d.SetPrivate(ctx, replacementSourceKey, pending)
	}
	payload, known := d.GetPlanOk("schema")
	refs, _ := d.Get("references").([]any)
	if replacementCreate && !known && len(refs) == 0 {
		return fmt.Errorf("schema must be known before replacing subject %q so it can be validated before deletion", source.Subject)
	}
	if !known || (!replacementCreate && !d.IsPlanFullyKnown("references")) {
		return nil
	}
	schemaPayload := payload.(string)
	if hasVersion {
		changed, err := schemaChanged(d)
		if err != nil {
			return err
		}
		if !changed {
			// State-only updates make no registry call and retain an observed level.
			if _, configured := d.GetConfigOk("compatibility_level"); !configured && !d.IsConfigUnknown("compatibility_level") {
				level, ok := d.GetStateOk("compatibility_level")
				if !ok {
					level = nil // No subject override was configured in the registry.
				}
				if err := d.Set("compatibility_level", level); err != nil {
					return err
				}
			}
			return d.Set("version", stateVersion)
		}
	}

	// The registry compatibility request has no references field, and Avro
	// can't parse unresolved imports locally.
	if len(refs) > 0 {
		return nil
	}

	if schemaType == kafkaschemaregistry.SchemaTypeAvro {
		if _, err := avro.Parse(schemaPayload); err != nil {
			return fmt.Errorf("schema validation error: %w", err)
		}
	}

	if source.Version == 0 {
		return nil
	}

	return checkCompatibility(ctx, client, d, schemaPayload, schemaType, source, replacementCreate)
}

func replacementDestinationVersions(ctx context.Context, client avngen.Client, project, serviceName, subjectName string) ([]int, error) {
	versions, err := client.ServiceSchemaRegistrySubjectVersionsGet(ctx, project, serviceName, subjectName)
	if err == nil && len(versions) > 0 {
		return versions, nil
	}
	svc, serviceErr := kafkaService(ctx, client, project, serviceName)
	if serviceErr != nil {
		return nil, fmt.Errorf("unable to read destination service %q in project %q: %w", serviceName, project, serviceErr)
	}
	if svc.UserConfig["schema_registry"] != true {
		// A dependency may enable the registry during this apply.
		return nil, nil
	}
	if err != nil && !avngen.IsNotFound(err) {
		return nil, err
	}
	// The versions URL can briefly return 404 after a successful registration.
	// Confirm the subject is absent before validating against the source instead.
	subjects, listErr := client.ServiceSchemaRegistrySubjects(ctx, project, serviceName)
	if listErr == nil && !slices.Contains(subjects, subjectName) {
		return nil, nil
	}
	versions, err = client.ServiceSchemaRegistrySubjectVersionsGet(ctx, project, serviceName, subjectName)
	if err == nil && len(versions) > 0 {
		return versions, nil
	}
	if listErr != nil {
		return nil, fmt.Errorf("unable to confirm whether destination subject %q exists: %w", subjectName, listErr)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to read versions of existing destination subject %q: %w", subjectName, err)
	}
	return nil, fmt.Errorf("destination subject %q exists but its versions are not yet visible", subjectName)
}

func checkCompatibility(ctx context.Context, client avngen.Client, d adapter.ResourceData, schemaPayload string, schemaType kafkaschemaregistry.SchemaType, source subjectVersion, replacing bool) error {
	var identity [3]string
	for i, field := range [...]string{"project", "service_name", "subject_name"} {
		value, known := d.GetPlanOk(field)
		if !known {
			if replacing {
				return fmt.Errorf("%s must be known before replacing subject %q so the destination can be checked before deletion", field, source.Subject)
			}
			return nil
		}
		identity[i] = value.(string)
	}
	project, serviceName, subjectName := identity[0], identity[1], identity[2]
	version := source.Version
	movedParent := project != source.Project || serviceName != source.Service
	var destinationVersions []int
	requireCompatible := true
	if replacing && (movedParent || subjectName != source.Subject) {
		versions, err := replacementDestinationVersions(ctx, client, project, serviceName, subjectName)
		if err != nil {
			return err
		}
		if len(versions) > 0 {
			version = slices.Max(versions)
			destinationVersions = versions
		} else {
			// A new destination has no history. Validate the payload against
			// the source before letting Terraform delete it.
			requireCompatible = false
			project, serviceName, subjectName = source.Project, source.Service, source.Subject
		}
	}
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

	if requireCompatible && !r.IsCompatible {
		// Karapace returns an existing ID before checking compatibility of a new version.
		// Only inspect history when compatibility would otherwise prevent the move.
		kind := cmp.Or(schemaType, kafkaschemaregistry.SchemaTypeAvro)
		var textPair []byte
		if !movedParent && len(destinationVersions) > 0 {
			textPair, err = d.GetPrivate(ctx, registeredTextKey)
			if err != nil {
				return err
			}
		}
		for _, v := range destinationVersions {
			existing, err := client.ServiceSchemaRegistrySubjectVersionGet(ctx, project, serviceName, subjectName, v)
			if avngen.IsNotFound(err) {
				continue
			}
			if err != nil {
				return fmt.Errorf("unable to read destination subject %q, version %d: %w", subjectName, v, err)
			}
			if len(existing.References) != 0 || cmp.Or(existing.SchemaType, kafkaschemaregistry.SchemaTypeAvro) != kind {
				continue
			}
			if sameRegistrationText(schemaPayload, existing.Schema) ||
				(len(textPair) > 0 && bytes.Equal(textPair, registrationTextPair(kind, schemaPayload, existing.Schema))) {
				return nil
			}
		}
		return fmt.Errorf(
			"subject %q: schema is not compatible with previous version %d: %s",
			subjectName,
			version,
			strings.Join(r.Messages, ", "),
		)
	}
	return nil
}
