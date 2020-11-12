// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"
)

var aivenMirrorMakerReplicationFlowSchema = map[string]*schema.Schema{
	"project": {
		Type:         schema.TypeString,
		Required:     true,
		Description:  "Project to link the kafka topic to",
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 63),
	},
	"service_name": {
		Type:         schema.TypeString,
		Required:     true,
		Description:  "Service to link the kafka topic to",
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 63),
	},
	"enable": {
		Type:        schema.TypeBool,
		Required:    true,
		Description: "Enable of disable replication flows for a service",
	},
	"source_cluster": {
		Type:         schema.TypeString,
		Required:     true,
		Description:  "Source cluster alias",
		ValidateFunc: validation.StringLenBetween(1, 128),
	},
	"target_cluster": {
		Type:         schema.TypeString,
		Required:     true,
		Description:  "Target cluster alias",
		ValidateFunc: validation.StringLenBetween(1, 128),
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
}

func resourceMirrorMakerReplicationFlow() *schema.Resource {
	return &schema.Resource{
		Create: resourceMirrorMakerReplicationFlowCreate,
		Read:   resourceMirrorMakerReplicationFlowRead,
		Update: resourceMirrorMakerReplicationFlowUpdate,
		Delete: resourceMirrorMakerReplicationFlowDelete,
		Exists: resourceMirrorMakerReplicationFlowExists,
		Importer: &schema.ResourceImporter{
			State: resourceMirrorMakerReplicationFlowState,
		},

		Schema: aivenMirrorMakerReplicationFlowSchema,
	}
}

func resourceMirrorMakerReplicationFlowCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	enable := d.Get("enable").(bool)
	sourceCluster := d.Get("source_cluster").(string)
	targetCluster := d.Get("target_cluster").(string)
	topics := flattenToString(d.Get("topics").([]interface{}))
	topicsBlacklist := flattenToString(d.Get("topics_blacklist").([]interface{}))

	err := client.KafkaMirrorMakerReplicationFlow.Create(project, serviceName, aiven.MirrorMakerReplicationFlowRequest{
		ReplicationFlow: aiven.ReplicationFlow{
			Enabled:         enable,
			SourceCluster:   sourceCluster,
			TargetCluster:   targetCluster,
			Topics:          topics,
			TopicsBlacklist: topicsBlacklist,
		},
	})
	if err != nil && !aiven.IsAlreadyExists(err) {
		return err
	}

	d.SetId(buildResourceID(project, serviceName, sourceCluster, targetCluster))

	return resourceMirrorMakerReplicationFlowRead(d, m)
}

func resourceMirrorMakerReplicationFlowRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project, serviceName, sourceCluster, targetCluster := splitResourceID4(d.Id())
	replicationFlow, err := client.KafkaMirrorMakerReplicationFlow.Get(project, serviceName, sourceCluster, targetCluster)
	if err != nil {
		return err
	}

	if err := d.Set("project", project); err != nil {
		return err
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err := d.Set("enable", replicationFlow.ReplicationFlow.Enabled); err != nil {
		return err
	}
	if err := d.Set("source_cluster", sourceCluster); err != nil {
		return err
	}
	if err := d.Set("target_cluster", targetCluster); err != nil {
		return err
	}
	if err := d.Set("topics", replicationFlow.ReplicationFlow.Topics); err != nil {
		return err
	}
	if err := d.Set("topics_blacklist", replicationFlow.ReplicationFlow.TopicsBlacklist); err != nil {
		return err
	}

	return nil
}

func resourceMirrorMakerReplicationFlowUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	enable := d.Get("enable").(bool)
	topics := flattenToString(d.Get("topics").([]interface{}))
	topicsBlacklist := flattenToString(d.Get("topics_blacklist").([]interface{}))

	project, serviceName, sourceCluster, targetCluster := splitResourceID4(d.Id())
	_, err := client.KafkaMirrorMakerReplicationFlow.Update(
		project,
		serviceName,
		sourceCluster,
		targetCluster,
		aiven.MirrorMakerReplicationFlowRequest{
			ReplicationFlow: aiven.ReplicationFlow{
				Enabled:         enable,
				Topics:          topics,
				TopicsBlacklist: topicsBlacklist,
			},
		},
	)
	if err != nil {
		return err
	}

	return resourceMirrorMakerReplicationFlowRead(d, m)
}

func resourceMirrorMakerReplicationFlowDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project, serviceName, sourceCluster, targetCluster := splitResourceID4(d.Id())

	return client.KafkaMirrorMakerReplicationFlow.Delete(project, serviceName, sourceCluster, targetCluster)
}

func resourceMirrorMakerReplicationFlowExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	project, serviceName, sourceCluster, targetCluster := splitResourceID4(d.Id())
	_, err := client.KafkaMirrorMakerReplicationFlow.Get(project, serviceName, sourceCluster, targetCluster)

	return resourceExists(err)
}

func resourceMirrorMakerReplicationFlowState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 4 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>/<source_cluster>/<target_cluster>", d.Id())
	}

	err := resourceMirrorMakerReplicationFlowRead(d, m)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
