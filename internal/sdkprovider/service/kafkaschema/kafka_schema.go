package kafkaschema

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// newlineRegExp is a regular expression that matches a newline.
var newlineRegExp = regexp.MustCompile(`\r?\n`)

var aivenKafkaSchemaSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"subject_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The Kafka Schema Subject name.").ForceNew().Build(),
	},
	"schema": {
		Type:             schema.TypeString,
		Required:         true,
		StateFunc:        normalizeJSONOrProtobufString,
		DiffSuppressFunc: diffSuppressJSONObjectOrProtobufString,
		Description: "Kafka Schema configuration. Should be a valid Avro, JSON, or Protobuf schema," +
			" depending on the schema type.",
	},
	"schema_type": {
		Type:     schema.TypeString,
		Optional: true,
		ForceNew: true,
		Description: "Kafka Schema configuration type. Defaults to AVRO. Possible values are AVRO, JSON, " +
			"and PROTOBUF.",
		Default:      "AVRO",
		ValidateFunc: validation.StringInSlice([]string{"AVRO", "JSON", "PROTOBUF"}, false),
		DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
			// This field can't be retrieved once resource is created.
			// That produces a diff on plan on resource import.
			// Ignores imported field.
			return oldValue == "" && d.Id() != ""
		},
	},
	"version": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Kafka Schema configuration version.",
	},
	"compatibility_level": {
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(compatibilityLevels, false),
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			// When a compatibility level is not set to any value and consequently is null (empty string).
			// Allow ignoring those.
			return new == ""
		},
		Description: userconfig.Desc("Kafka Schemas compatibility level.").PossibleValues(schemautil.StringSliceToInterfaceSlice(compatibilityLevels)...).Build(),
	},
}

// diffSuppressJSONObject checks logical equivalences in JSON Kafka Schema values
func diffSuppressJSONObject(_, old, new string, _ *schema.ResourceData) bool {
	var objOld, objNew interface{}

	if err := json.Unmarshal([]byte(old), &objOld); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &objNew); err != nil {
		return false
	}

	return reflect.DeepEqual(objNew, objOld)
}

// diffSuppressJSONObjectOrProtobufString checks logical equivalences in JSON or Protobuf Kafka Schema values.
func diffSuppressJSONObjectOrProtobufString(k, old, new string, d *schema.ResourceData) bool {
	if !diffSuppressJSONObject(k, old, new, d) {
		return normalizeProtobufString(old) == normalizeProtobufString(new)
	}

	return false
}

// normalizeProtobufString returns normalized Protobuf string.
func normalizeProtobufString(i any) string {
	v := i.(string)

	return newlineRegExp.ReplaceAllString(v, "")
}

// normalizeJSONOrProtobufString returns normalized JSON or Protobuf string.
func normalizeJSONOrProtobufString(i any) string {
	v := i.(string)

	if n, err := structure.NormalizeJsonString(v); err == nil {
		return n
	}

	return normalizeProtobufString(v)
}

func ResourceKafkaSchema() *schema.Resource {
	return &schema.Resource{
		Description:   "The Kafka Schema resource allows the creation and management of Aiven Kafka Schemas.",
		CreateContext: resourceKafkaSchemaCreate,
		UpdateContext: resourceKafkaSchemaUpdate,
		ReadContext:   resourceKafkaSchemaRead,
		DeleteContext: resourceKafkaSchemaDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: resourceKafkaSchemaCustomizeDiff,
		Timeouts:      schemautil.DefaultResourceTimeouts(),

		Schema: aivenKafkaSchemaSchema,
	}
}

func kafkaSchemaSubjectGetLastVersion(
	ctx context.Context,
	m interface{},
	project string,
	serviceName string,
	subjectName string,
) (int, error) {
	client := m.(*aiven.Client)

	r, err := client.KafkaSubjectSchemas.GetVersions(ctx, project, serviceName, subjectName)
	if err != nil {
		return 0, err
	}

	var latestVersion int
	for _, v := range r.Versions {
		if v > latestVersion {
			latestVersion = v
		}
	}

	return latestVersion, nil
}

// Aiven Kafka schema creates a new Kafka Schema Subject with a new version, and if Kafka
// Schema subject with a given name already exists API will validate new Kafka Schema
// configuration against the previous version for compatibility and if compatible will
// create a new version for the same Kafka Schema Subject
func resourceKafkaSchemaCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	subjectName := d.Get("subject_name").(string)

	client := m.(*aiven.Client)

	// create Kafka Schema Subject
	_, err := client.KafkaSubjectSchemas.Add(
		ctx,
		project,
		serviceName,
		subjectName,
		aiven.KafkaSchemaSubject{
			Schema:     d.Get("schema").(string),
			SchemaType: d.Get("schema_type").(string),
		},
	)
	if err != nil {
		return diag.Errorf("unable to create schema: %s", err)
	}

	// set compatibility level if defined for a newly created Kafka Schema Subject
	if compatibility, ok := d.GetOk("compatibility_level"); ok {
		_, err := client.KafkaSubjectSchemas.UpdateConfiguration(
			ctx,
			project,
			serviceName,
			subjectName,
			compatibility.(string),
		)
		if err != nil {
			return diag.Errorf("unable to update configuration: %s", err)
		}
	}

	version, err := kafkaSchemaSubjectGetLastVersion(ctx, m, project, serviceName, subjectName)
	if err != nil {
		return diag.Errorf("unable to get last version: %s", err)
	}

	// newly created versions start from 1
	if version == 0 {
		return diag.Errorf("kafka schema subject after creation has an empty list of versions")
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, subjectName))

	return resourceKafkaSchemaRead(ctx, d, m)
}

func resourceKafkaSchemaUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project, serviceName, subjectName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client := m.(*aiven.Client)

	if d.HasChange("schema") {
		_, err := client.KafkaSubjectSchemas.Add(
			ctx,
			project,
			serviceName,
			subjectName,
			aiven.KafkaSchemaSubject{
				Schema:     d.Get("schema").(string),
				SchemaType: d.Get("schema_type").(string),
			},
		)
		if err != nil {
			return diag.Errorf("unable to update schema: %s", err)
		}
	}

	// if compatibility_level has changed and the new value is not empty
	if compatibility, ok := d.GetOk("compatibility_level"); ok {
		_, err = client.KafkaSubjectSchemas.UpdateConfiguration(
			ctx,
			project,
			serviceName,
			subjectName,
			compatibility.(string))
		if err != nil {
			return diag.Errorf("unable to update configuration: %s", err)
		}
	}

	return resourceKafkaSchemaRead(ctx, d, m)
}

func resourceKafkaSchemaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project, serviceName, subjectName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client := m.(*aiven.Client)

	version, err := kafkaSchemaSubjectGetLastVersion(ctx, m, project, serviceName, subjectName)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	r, err := client.KafkaSubjectSchemas.Get(ctx, project, serviceName, subjectName, version)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if err := d.Set("project", project); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("subject_name", subjectName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("version", version); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("schema", r.Version.Schema); err != nil {
		return diag.FromErr(err)
	}

	c, err := client.KafkaSubjectSchemas.GetConfiguration(ctx, project, serviceName, subjectName)
	if err != nil {
		if !aiven.IsNotFound(err) {
			return diag.FromErr(err)
		}
	} else {
		if err := d.Set("compatibility_level", c.CompatibilityLevel); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceKafkaSchemaDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project, serviceName, schemaName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = m.(*aiven.Client).KafkaSubjectSchemas.Delete(ctx, project, serviceName, schemaName)
	if common.IsCritical(err) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKafkaSchemaCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	client := m.(*aiven.Client)

	// no previous version: allow the diff, nothing to check compatibility against
	if _, ok := d.GetOk("version"); !ok {
		return nil
	}

	if compatible, err := client.KafkaSubjectSchemas.Validate(
		ctx,
		d.Get("project").(string),
		d.Get("service_name").(string),
		d.Get("subject_name").(string),
		d.Get("version").(int),
		aiven.KafkaSchemaSubject{
			Schema:     d.Get("schema").(string),
			SchemaType: d.Get("schema_type").(string),
		},
	); err != nil {
		return fmt.Errorf("unable to check schema validity: %w", err)
	} else if !compatible {
		return fmt.Errorf("schema is not compatible with previous version")
	}

	return nil
}
