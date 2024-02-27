package kafkaschema

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"golang.org/x/exp/slices"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// whitespaceRegExp is a regular expression to match a whitespace, a new line, or a carriage return
// character in a string. This is used to normalize Protobuf strings to compare them for logical equivalence.
var whitespaceRegExp = regexp.MustCompile(`\s`)

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
		DiffSuppressFunc: func(_, oldValue, _ string, d *schema.ResourceData) bool {
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
		DiffSuppressFunc: func(_, _, new string, _ *schema.ResourceData) bool {
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

	return true
}

// normalizeProtobufString returns normalized Protobuf string.
func normalizeProtobufString(i any) string {
	v := i.(string)

	return whitespaceRegExp.ReplaceAllString(v, "")
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
		CreateContext: resourceKafkaSchemaUpsert,
		UpdateContext: resourceKafkaSchemaUpsert,
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

func resourceKafkaSchemaUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	subjectName := d.Get("subject_name").(string)

	client := m.(*aiven.Client)
	if d.HasChange("schema") {
		// This call returns Schema ID, not its version
		s, err := client.KafkaSubjectSchemas.Add(
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
			return diag.Errorf("unable to add schema: %s", err)
		}

		// Gets Schema's version by its ID
		version, err := getSchemaVersion(ctx, client, project, serviceName, subjectName, s.Id)
		if err != nil {
			return diag.Errorf("unable to get schema version: %s", err)
		}

		if err := d.Set("version", version); err != nil {
			return diag.FromErr(err)
		}
	}

	// if compatibility_level has changed and the new value is not empty
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

	d.SetId(schemautil.BuildResourceID(project, serviceName, subjectName))
	return resourceKafkaSchemaRead(ctx, d, m)
}

// getSchemaVersion polls until the version with given Schema ID appears in the version list
func getSchemaVersion(ctx context.Context, client *aiven.Client, project, serviceName, subjectName string, id int) (int, error) {
	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(time.Second):
			versions, err := client.KafkaSubjectSchemas.GetVersions(ctx, project, serviceName, subjectName)
			if err != nil {
				return 0, err
			}

			for _, v := range versions.Versions {
				s, err := client.KafkaSubjectSchemas.Get(ctx, project, serviceName, subjectName, v)
				if err != nil {
					return 0, err
				}

				if s.Version.Id == id {
					return s.Version.Version, nil
				}
			}
		}
	}
}

func resourceKafkaSchemaRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	project, serviceName, subjectName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	client := m.(*aiven.Client)
	version := d.Get("version").(int)
	if version == 0 {
		// For data source type and "import"
		r, err := client.KafkaSubjectSchemas.GetVersions(ctx, project, serviceName, subjectName)
		if err != nil {
			return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
		}
		version = slices.Max(r.Versions)
		if err := d.Set("version", version); err != nil {
			return diag.FromErr(err)
		}
	}

	s, err := client.KafkaSubjectSchemas.Get(ctx, project, serviceName, subjectName, version)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema", s.Version.Schema); err != nil {
		return diag.FromErr(err)
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
