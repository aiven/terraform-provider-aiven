package opensearch

import (
	"sync"

	"github.com/aiven/aiven-go-client"
)

var (
	// this mutex is needed to serialize calls to modify the remote config
	// since it`s an abstraction that first GETs, modifies and then PUTs again
	resourceOpensearchACLModifierMutex sync.Mutex
)

// GETs the remote config, applies the modifiers and PUTs it again
// The Config that is passed to the modifiers is guaranteed to be not nil
func resourceOpensearchACLModifyRemoteConfig(project, serviceName string, client *aiven.Client, modifiers ...func(*aiven.ElasticSearchACLConfig)) error {
	resourceOpensearchACLModifierMutex.Lock()
	defer resourceOpensearchACLModifierMutex.Unlock()

	r, err := client.ElasticsearchACLs.Get(project, serviceName)
	if err != nil {
		return err
	}

	config := r.ElasticSearchACLConfig
	for i := range modifiers {
		modifiers[i](&config)
	}

	_, err = client.ElasticsearchACLs.Update(
		project,
		serviceName,
		aiven.ElasticsearchACLRequest{ElasticSearchACLConfig: config})
	if err != nil {
		return err
	}
	return nil
}

// some modifiers

func resourceElasticsearchACLModifierUpdateACLRule(username, index, permission string) func(*aiven.ElasticSearchACLConfig) {
	return func(cfg *aiven.ElasticSearchACLConfig) {
		cfg.Add(resourceOpensearchACLRuleMkAivenACL(username, index, permission))

		// delete the old acl if its there
		if prevPerm, ok := resourceElasticsearchACLRuleGetPermissionFromACLResponse(*cfg, username, index); ok && prevPerm != permission {
			cfg.Delete(resourceOpensearchACLRuleMkAivenACL(username, index, prevPerm))
		}
	}
}

func resourceElasticsearchACLModifierDeleteACLRule(username, index, permission string) func(*aiven.ElasticSearchACLConfig) {
	return func(cfg *aiven.ElasticSearchACLConfig) {
		cfg.Delete(resourceOpensearchACLRuleMkAivenACL(username, index, permission))
	}
}

func resourceElasticsearchACLModifierToggleConfigFields(enabled, extednedAcl bool) func(*aiven.ElasticSearchACLConfig) {
	return func(cfg *aiven.ElasticSearchACLConfig) {
		cfg.Enabled = enabled
		cfg.ExtendedAcl = extednedAcl
	}
}
