package kafka

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var (
	defaultReplicationPolicy = "org.apache.kafka.connect.mirror.DefaultReplicationPolicy"

	replicationPolicies = []string{
		"org.apache.kafka.connect.mirror.DefaultReplicationPolicy",
		"org.apache.kafka.connect.mirror.IdentityReplicationPolicy",
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
		Optional:     true,
		Default:      defaultReplicationPolicy,
		ValidateFunc: validation.StringInSlice(replicationPolicies, false),
		Description:  userconfig.Desc("Replication policy class.").DefaultValue(defaultReplicationPolicy).PossibleValues(schemautil.StringSliceToInterfaceSlice(replicationPolicies)...).Build(),
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
		Optional:     true,
		Description:  "Offset syncs topic location.",
		ValidateFunc: validation.StringInSlice([]string{"source", "target"}, false),
	},
}

func ResourceMirrorMakerReplicationFlow() *schema.Resource {
	return &schema.Resource{
		Description:   "The MirrorMaker 2 Replication Flow resource allows the creation and management of MirrorMaker 2 Replication Flows on Aiven Cloud.",
		CreateContext: resourceMirrorMakerReplicationFlowCreate,
		ReadContext:   resourceMirrorMakerReplicationFlowRead,
		UpdateContext: resourceMirrorMakerReplicationFlowUpdate,
		DeleteContext: resourceMirrorMakerReplicationFlowDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenMirrorMakerReplicationFlowSchema,
	}
}

func resourceMirrorMakerReplicationFlowCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	enable := d.Get("enable").(bool)
	sourceCluster := d.Get("source_cluster").(string)
	targetCluster := d.Get("target_cluster").(string)

	err := client.KafkaMirrorMakerReplicationFlow.Create(ctx, project, serviceName, aiven.MirrorMakerReplicationFlowRequest{
		ReplicationFlow: aiven.ReplicationFlow{
			Enabled:                         enable,
			SourceCluster:                   sourceCluster,
			TargetCluster:                   d.Get("target_cluster").(string),
			Topics:                          schemautil.FlattenToString(d.Get("topics").([]interface{})),
			TopicsBlacklist:                 schemautil.FlattenToString(d.Get("topics_blacklist").([]interface{})),
			ReplicationPolicyClass:          d.Get("replication_policy_class").(string),
			SyncGroupOffsetsEnabled:         d.Get("sync_group_offsets_enabled").(bool),
			SyncGroupOffsetsIntervalSeconds: d.Get("sync_group_offsets_interval_seconds").(int),
			EmitHeartbeatsEnabled:           d.Get("emit_heartbeats_enabled").(bool),
			EmitBackwardHeartbeatsEnabled:   d.Get("emit_backward_heartbeats_enabled").(bool),
			OffsetSyncsTopicLocation:        d.Get("offset_syncs_topic_location").(string),
		},
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, sourceCluster, targetCluster))

	return resourceMirrorMakerReplicationFlowRead(ctx, d, m)
}

func resourceMirrorMakerReplicationFlowRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, sourceCluster, targetCluster, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	replicationFlow, err := client.KafkaMirrorMakerReplicationFlow.Get(ctx, project, serviceName, sourceCluster, targetCluster)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if err := d.Set("project", project); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enable", replicationFlow.ReplicationFlow.Enabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("source_cluster", sourceCluster); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("target_cluster", targetCluster); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("topics", replicationFlow.ReplicationFlow.Topics); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("topics_blacklist", replicationFlow.ReplicationFlow.TopicsBlacklist); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("replication_policy_class", replicationFlow.ReplicationFlow.ReplicationPolicyClass); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("sync_group_offsets_enabled", replicationFlow.ReplicationFlow.SyncGroupOffsetsEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("sync_group_offsets_interval_seconds", replicationFlow.ReplicationFlow.SyncGroupOffsetsIntervalSeconds); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("emit_heartbeats_enabled", replicationFlow.ReplicationFlow.EmitHeartbeatsEnabled); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(
		"emit_backward_heartbeats_enabled",
		replicationFlow.ReplicationFlow.EmitBackwardHeartbeatsEnabled,
	); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceMirrorMakerReplicationFlowUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, sourceCluster, targetCluster, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.KafkaMirrorMakerReplicationFlow.Update(
		ctx,
		project,
		serviceName,
		sourceCluster,
		targetCluster,
		aiven.MirrorMakerReplicationFlowRequest{
			ReplicationFlow: aiven.ReplicationFlow{
				Enabled:                         d.Get("enable").(bool),
				Topics:                          schemautil.FlattenToString(d.Get("topics").([]interface{})),
				TopicsBlacklist:                 schemautil.FlattenToString(d.Get("topics_blacklist").([]interface{})),
				ReplicationPolicyClass:          d.Get("replication_policy_class").(string),
				SyncGroupOffsetsEnabled:         d.Get("sync_group_offsets_enabled").(bool),
				SyncGroupOffsetsIntervalSeconds: d.Get("sync_group_offsets_interval_seconds").(int),
				EmitHeartbeatsEnabled:           d.Get("emit_heartbeats_enabled").(bool),
				EmitBackwardHeartbeatsEnabled:   d.Get("emit_backward_heartbeats_enabled").(bool),
				OffsetSyncsTopicLocation:        d.Get("offset_syncs_topic_location").(string),
			},
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceMirrorMakerReplicationFlowRead(ctx, d, m)
}

func resourceMirrorMakerReplicationFlowDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, sourceCluster, targetCluster, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.KafkaMirrorMakerReplicationFlow.Delete(ctx, project, serviceName, sourceCluster, targetCluster)
	if err != nil {
		diag.FromErr(err)
	}

	return nil
}
