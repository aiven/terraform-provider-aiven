package kafka

import (
	"context"
	"strings"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkamirrormaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

const configPropsKey = "config_properties_exclude"

var (
	defaultReplicationPolicy = "org.apache.kafka.connect.mirror.DefaultReplicationPolicy"
	// dtoFieldsAliases stores DTO fields mapping: terraform -> json
	dtoFieldsAliases = map[string]string{
		"enable":           "enabled",
		"topics_blacklist": "topics.blacklist",
	}
)

var aivenMirrorMakerReplicationFlowSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaProjectReference,

	"enable": {
		Type:        schema.TypeBool,
		Required:    true,
		Description: "Enable of disable replication flows for a service.",
	},
	"source_cluster": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(1, 128),
		Description:  userconfig.Desc("Source cluster alias.").MaxLen(128).Build(),
	},
	"target_cluster": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringLenBetween(1, 128),
		Description:  userconfig.Desc("Target cluster alias.").MaxLen(128).Build(),
	},
	"topics": {
		Type:        schema.TypeList,
		Optional:    true,
		Description: "List of topics and/or regular expressions to replicate",
		Elem: &schema.Schema{
			Type:     schema.TypeString,
			MaxItems: 256,
		},
	},
	"topics_blacklist": {
		Type:        schema.TypeList,
		Optional:    true,
		Description: "List of topics and/or regular expressions to not replicate.",
		Elem: &schema.Schema{
			Type:     schema.TypeString,
			MaxItems: 256,
		},
	},
	"replication_policy_class": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice(kafkamirrormaker.ReplicationPolicyClassTypeChoices(), false),
		Description: userconfig.Desc("Replication policy class.").
			DefaultValue(defaultReplicationPolicy).
			PossibleValuesString(kafkamirrormaker.ReplicationPolicyClassTypeChoices()...).Build(),
	},
	"sync_group_offsets_enabled": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: userconfig.Desc("Sync consumer group offsets.").DefaultValue(false).Build(),
	},
	"sync_group_offsets_interval_seconds": {
		Type:         schema.TypeInt,
		Optional:     true,
		ValidateFunc: validation.IntAtLeast(1),
		Default:      1,
		Description:  userconfig.Desc("Frequency of consumer group offset sync.").DefaultValue(1).Build(),
	},
	"emit_heartbeats_enabled": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  false,
		Description: userconfig.Desc(
			"Whether to emit heartbeats to the target cluster",
		).DefaultValue(false).Build(),
	},
	"emit_backward_heartbeats_enabled": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  false,
		Description: userconfig.Desc(
			"Whether to emit heartbeats to the direction opposite to the flow, i.e. to the source cluster",
		).DefaultValue(false).Build(),
	},
	"offset_syncs_topic_location": {
		Type:         schema.TypeString,
		Required:     true,
		Description:  "Offset syncs topic location. Possible values are `source` & `target`. There is no default value.",
		ValidateFunc: validation.StringInSlice([]string{"source", "target"}, false),
	},
	"config_properties_exclude": {
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "List of topic configuration properties and/or regular expressions to not replicate. The properties that are not replicated by default are: `follower.replication.throttled.replicas`, `leader.replication.throttled.replicas`, `message.timestamp.difference.max.ms`, `message.timestamp.type`, `unclean.leader.election.enable`, and `min.insync.replicas`. Setting this overrides the defaults. For example, to enable replication for 'min.insync.replicas' and 'unclean.leader.election.enable' set this to: [\"follower\\\\\\\\.replication\\\\\\\\.throttled\\\\\\\\.replicas\", \"leader\\\\\\\\.replication\\\\\\\\.throttled\\\\\\\\.replicas\", \"message\\\\\\\\.timestamp\\\\\\\\.difference\\\\\\\\.max\\\\\\\\.ms\",  \"message\\\\\\\\.timestamp\\\\\\\\.type\"]",
		Elem: &schema.Schema{
			Type:     schema.TypeString,
			MaxItems: 256,
		},
	},
	"replication_factor": {
		Type:         schema.TypeInt,
		Optional:     true,
		ValidateFunc: validation.IntAtLeast(1),
		Description:  "Replication factor, `>= 1`.",
	},
}

func ResourceMirrorMakerReplicationFlow() *schema.Resource {
	return &schema.Resource{
		Description:   "The MirrorMaker 2 Replication Flow resource allows the creation and management of MirrorMaker 2 Replication Flows on Aiven Cloud.",
		CreateContext: common.WithGenClient(resourceMirrorMakerReplicationFlowCreate),
		ReadContext:   common.WithGenClient(resourceMirrorMakerReplicationFlowRead),
		UpdateContext: common.WithGenClient(resourceMirrorMakerReplicationFlowUpdate),
		DeleteContext: common.WithGenClient(resourceMirrorMakerReplicationFlowDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenMirrorMakerReplicationFlowSchema,
	}
}

func resourceMirrorMakerReplicationFlowCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	sourceCluster := d.Get("source_cluster").(string)
	targetCluster := d.Get("target_cluster").(string)

	dto := new(kafkamirrormaker.ServiceKafkaMirrorMakerCreateReplicationFlowIn)
	err := marshalFlow(d, dto)
	if err != nil {
		return err
	}

	err = client.ServiceKafkaMirrorMakerCreateReplicationFlow(ctx, project, serviceName, dto)
	if err != nil {
		return err
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, sourceCluster, targetCluster))

	return resourceMirrorMakerReplicationFlowRead(ctx, d, client)
}

func resourceMirrorMakerReplicationFlowRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	project, serviceName, sourceCluster, targetCluster, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return err
	}

	dto, err := client.ServiceKafkaMirrorMakerGetReplicationFlow(ctx, project, serviceName, sourceCluster, targetCluster)
	if err != nil {
		return schemautil.ResourceReadHandleNotFound(err, d)
	}

	if err := d.Set("project", project); err != nil {
		return err
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}

	err = schemautil.ResourceDataSet(
		aivenMirrorMakerReplicationFlowSchema, d, dto, schemautil.RenameAliasesSet(dtoFieldsAliases),
		func(k string, v any) (string, any) {
			if k == configPropsKey {
				// This field is received as a string
				s := v.(string)
				v = make([]any, 0)
				if s != "" {
					v = strings.Split(s, ",")
				}
			}
			return k, v
		},
	)

	return err
}

func resourceMirrorMakerReplicationFlowUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	project, serviceName, sourceCluster, targetCluster, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return err
	}

	dto := new(kafkamirrormaker.ServiceKafkaMirrorMakerPatchReplicationFlowIn)
	err = marshalFlow(d, dto)
	if err != nil {
		return err
	}

	_, err = client.ServiceKafkaMirrorMakerPatchReplicationFlow(ctx, project, serviceName, sourceCluster, targetCluster, dto)
	if err != nil {
		return err
	}

	return resourceMirrorMakerReplicationFlowRead(ctx, d, client)
}

func resourceMirrorMakerReplicationFlowDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	project, serviceName, sourceCluster, targetCluster, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return err
	}

	err = client.ServiceKafkaMirrorMakerDeleteReplicationFlow(ctx, project, serviceName, sourceCluster, targetCluster)
	return err
}

func marshalFlow(d *schema.ResourceData, dto any) error {
	return schemautil.ResourceDataGet(
		d, dto, schemautil.RenameAliasesGet(dtoFieldsAliases),
		func(k string, v any) (string, any) {
			// This field is sent as a string
			if k == configPropsKey {
				v = strings.Join(schemautil.FlattenToString(v.([]any)), ",")
			}
			return k, v
		},
	)
}
