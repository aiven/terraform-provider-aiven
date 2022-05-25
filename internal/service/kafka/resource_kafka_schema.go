package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

var aivenKafkaSchemaSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"subject_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: schemautil.Complex("The Kafka Schema Subject name.").ForceNew().Build(),
	},
	"schema": {
		Type:             schema.TypeString,
		Required:         true,
		ValidateFunc:     validation.StringIsJSON,
		StateFunc:        normalizeJsonString,
		DiffSuppressFunc: diffSuppressJsonObject,
		Description:      "Kafka Schema configuration should be a valid Avro Schema JSON format.",
	},
	"schema_type": {
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		Description:  "Kafka Schema type JSON or AVRO",
		Default:      "AVRO",
		ValidateFunc: validation.StringInSlice([]string{"AVRO", "JSON"}, false),
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
		Description: schemautil.Complex("Kafka Schemas compatibility level.").PossibleValues(schemautil.StringSliceToInterfaceSlice(compatibilityLevels)...).Build(),
	},
}

// diffSuppressJsonObject checks logical equivalences in JSON Kafka Schema values
func diffSuppressJsonObject(_, old, new string, _ *schema.ResourceData) bool {
	var objOld, objNew interface{}

	if err := json.Unmarshal([]byte(old), &objOld); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &objNew); err != nil {
		return false
	}

	return reflect.DeepEqual(objNew, objOld)
}

// normalizeJsonString returns normalized JSON string
func normalizeJsonString(v interface{}) string {
	jsonString, _ := structure.NormalizeJsonString(v)

	return jsonString
}

func ResourceKafkaSchema() *schema.Resource {
	return &schema.Resource{
		Description:   "The Kafka Schema resource allows the creation and management of Aiven Kafka Schemas.",
		CreateContext: resourceKafkaSchemaCreate,
		UpdateContext: resourceKafkaSchemaUpdate,
		ReadContext:   resourceKafkaSchemaRead,
		DeleteContext: resourceKafkaSchemaDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKafkaSchemaState,
		},
		CustomizeDiff: resourceKafkaSchemaCustomizeDiff,

		Schema: aivenKafkaSchemaSchema,
	}
}

func kafkaSchemaSubjectGetLastVersion(m interface{}, project, serviceName, subjectName string) (int, error) {
	client := m.(*aiven.Client)

	r, err := client.KafkaSubjectSchemas.GetVersions(project, serviceName, subjectName)
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
			project,
			serviceName,
			subjectName,
			compatibility.(string),
		)
		if err != nil {
			return diag.Errorf("unable to update configuration: %s", err)
		}
	}

	version, err := kafkaSchemaSubjectGetLastVersion(m, project, serviceName, subjectName)
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
	var project, serviceName, subjectName = schemautil.SplitResourceID3(d.Id())
	client := m.(*aiven.Client)

	if d.HasChange("schema") {
		_, err := client.KafkaSubjectSchemas.Add(
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
	_, ok := d.GetOk("compatibility_level")
	if d.HasChange("compatibility_level") && ok {
		_, err := client.KafkaSubjectSchemas.UpdateConfiguration(
			project,
			serviceName,
			subjectName,
			d.Get("compatibility_level").(string))
		if err != nil {
			return diag.Errorf("unable to update configuration: %s", err)
		}
	}

	return resourceKafkaSchemaRead(ctx, d, m)
}

func resourceKafkaSchemaRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var project, serviceName, subjectName = schemautil.SplitResourceID3(d.Id())
	client := m.(*aiven.Client)

	version, err := kafkaSchemaSubjectGetLastVersion(m, project, serviceName, subjectName)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	r, err := client.KafkaSubjectSchemas.Get(project, serviceName, subjectName, version)
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

	c, err := client.KafkaSubjectSchemas.GetConfiguration(project, serviceName, subjectName)
	if err != nil {
		if !aiven.IsNotFound(err) {
			return diag.FromErr(err)
		}
	} else {
		// only update if was set to not empty values by the user
		if _, ok := d.GetOk("compatibility_level"); ok {
			if err := d.Set("compatibility_level", c.CompatibilityLevel); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return nil
}

func resourceKafkaSchemaDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var project, serviceName, schemaName = schemautil.SplitResourceID3(d.Id())

	err := m.(*aiven.Client).KafkaSubjectSchemas.Delete(project, serviceName, schemaName)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKafkaSchemaState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	di := resourceKafkaSchemaRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get kafka schema: %v", di)
	}

	return []*schema.ResourceData{d}, nil
}

func resourceKafkaSchemaCustomizeDiff(_ context.Context, d *schema.ResourceDiff, m interface{}) error {
	client := m.(*aiven.Client)

	// no previous version: allow the diff, nothing to check compatibility against
	if _, ok := d.GetOk("version"); !ok {
		return nil
	}

	if compatible, err := client.KafkaSubjectSchemas.Validate(
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
