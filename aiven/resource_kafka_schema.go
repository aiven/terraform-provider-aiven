package aiven

import (
	"encoding/json"
	"errors"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"reflect"
)

var aivenKafkaSchemaSchema = map[string]*schema.Schema{
	"project": {
		Type:        schema.TypeString,
		Description: "Project to link the Kafka Schema to",
		Required:    true,
		ForceNew:    true,
	},
	"service_name": {
		Type:        schema.TypeString,
		Description: "Service to link the Kafka Schema to",
		Required:    true,
		ForceNew:    true,
	},
	"subject_name": {
		Type:        schema.TypeString,
		Description: "Kafka Schema Subject name",
		Required:    true,
		ForceNew:    true,
	},
	"schema": {
		Type:             schema.TypeString,
		Description:      "Kafka Schema configuration should be a valid Avro Schema JSON format",
		Required:         true,
		ValidateFunc:     validation.StringIsJSON,
		StateFunc:        normalizeJsonString,
		DiffSuppressFunc: diffSuppressJsonObject,
	},
	"version": {
		Type:        schema.TypeInt,
		Description: "Kafka Schema configuration version",
		Computed:    true,
	},
	"compatibility_level": {
		Type:         schema.TypeString,
		Description:  "Kafka Schemas compatibility level",
		Optional:     true,
		ValidateFunc: validation.StringInSlice(compatibilityLevels, false),
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			// When a compatibility level is not set to any value and consequently is null (empty string).
			// Allow ignoring those.
			return new == ""
		},
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

func resourceKafkaSchema() *schema.Resource {
	return &schema.Resource{
		Create: resourceKafkaSchemaCreate,
		Update: resourceKafkaSchemaUpdate,
		Read:   resourceKafkaSchemaRead,
		Delete: resourceKafkaSchemaDelete,
		Exists: resourceKafkaSchemaExists,
		Importer: &schema.ResourceImporter{
			State: resourceKafkaSchemaState,
		},

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
func resourceKafkaSchemaCreate(d *schema.ResourceData, m interface{}) error {
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
			Schema: d.Get("schema").(string),
		},
	)
	if err != nil {
		return err
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
			return err
		}
	}

	version, err := kafkaSchemaSubjectGetLastVersion(m, project, serviceName, subjectName)
	if err != nil {
		return err
	}

	// newly created versions start from 1
	if version == 0 {
		return errors.New("kafka schema subject after creation has an empty list of versions")
	}

	d.SetId(buildResourceID(project, serviceName, subjectName))

	return resourceKafkaSchemaRead(d, m)
}

func resourceKafkaSchemaUpdate(d *schema.ResourceData, m interface{}) error {
	var project, serviceName, subjectName = splitResourceID3(d.Id())
	client := m.(*aiven.Client)

	if d.HasChange("schema") {
		_, err := client.KafkaSubjectSchemas.Add(
			project,
			serviceName,
			subjectName,
			aiven.KafkaSchemaSubject{
				Schema: d.Get("schema").(string),
			},
		)
		if err != nil {
			return err
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
			return err
		}
	}

	return resourceKafkaSchemaRead(d, m)
}

func resourceKafkaSchemaRead(d *schema.ResourceData, m interface{}) error {
	var project, serviceName, subjectName = splitResourceID3(d.Id())
	client := m.(*aiven.Client)

	version, err := kafkaSchemaSubjectGetLastVersion(m, project, serviceName, subjectName)
	if err != nil {
		return err
	}

	r, err := client.KafkaSubjectSchemas.Get(project, serviceName, subjectName, version)
	if err != nil {
		return err
	}

	if err := d.Set("project", project); err != nil {
		return err
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err := d.Set("subject_name", subjectName); err != nil {
		return err
	}
	if err := d.Set("version", version); err != nil {
		return err
	}
	if err := d.Set("schema", r.Version.Schema); err != nil {
		return err
	}

	c, err := client.KafkaSubjectSchemas.GetConfiguration(project, serviceName, subjectName)
	if err != nil {
		if !aiven.IsNotFound(err) {
			return err
		}
	} else {
		// only update if was set to not empty values by the user
		if _, ok := d.GetOk("compatibility_level"); ok {
			if err := d.Set("compatibility_level", c.CompatibilityLevel); err != nil {
				return err
			}
		}
	}

	return nil
}

func resourceKafkaSchemaDelete(d *schema.ResourceData, m interface{}) error {
	var project, serviceName, schemaName = splitResourceID3(d.Id())

	err := m.(*aiven.Client).KafkaSubjectSchemas.Delete(project, serviceName, schemaName)
	if err != nil && !aiven.IsNotFound(err) {
		return err
	}

	return nil
}

func resourceKafkaSchemaExists(d *schema.ResourceData, m interface{}) (bool, error) {
	return resourceExists(resourceKafkaSchemaRead(d, m))
}

func resourceKafkaSchemaState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	err := resourceKafkaSchemaRead(d, m)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
