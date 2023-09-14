package opensearch

import (
	"context"
	"sync"

	"github.com/aiven/aiven-go-client/v2"
)

var (
	// this mutex is needed to serialize calls to modify the remote config
	// since it's an abstraction that first GETs, modifies and then PUTs again
	resourceOpenSearchACLModifierMutex sync.Mutex
)

// GETs the remote config, applies the modifiers and PUTs it again
// The Config that is passed to the modifiers is guaranteed to be not nil
func resourceOpenSearchACLModifyRemoteConfig(
	ctx context.Context,
	project string,
	serviceName string,
	client *aiven.Client,
	modifiers ...func(*aiven.ElasticSearchACLConfig),
) error {
	resourceOpenSearchACLModifierMutex.Lock()
	defer resourceOpenSearchACLModifierMutex.Unlock()

	r, err := client.ElasticsearchACLs.Get(ctx, project, serviceName)
	if err != nil {
		return err
	}

	config := r.ElasticSearchACLConfig
	for i := range modifiers {
		modifiers[i](&config)
	}

	_, err = client.ElasticsearchACLs.Update(
		ctx,
		project,
		serviceName,
		aiven.ElasticsearchACLRequest{ElasticSearchACLConfig: config})
	if err != nil {
		return err
	}
	return nil
}

// some modifiers

func resourceElasticsearchACLModifierUpdateACLRule(
	ctx context.Context,
	username string,
	index string,
	permission string,
) func(*aiven.ElasticSearchACLConfig) {
	return func(cfg *aiven.ElasticSearchACLConfig) {
		cfg.Add(resourceOpenSearchACLRuleMkAivenACL(username, index, permission))

		// delete the old acl if it's there
		if prevPerm, ok := resourceElasticsearchACLRuleGetPermissionFromACLResponse(*cfg, username, index); ok && prevPerm != permission {
			cfg.Delete(ctx, resourceOpenSearchACLRuleMkAivenACL(username, index, prevPerm))
		}
	}
}

func resourceElasticsearchACLModifierDeleteACLRule(
	ctx context.Context,
	username string,
	index string,
	permission string,
) func(*aiven.ElasticSearchACLConfig) {
	return func(cfg *aiven.ElasticSearchACLConfig) {
		cfg.Delete(ctx, resourceOpenSearchACLRuleMkAivenACL(username, index, permission))
	}
}

func resourceElasticsearchACLModifierToggleConfigFields(enabled, extednedACL bool) func(*aiven.ElasticSearchACLConfig) {
	return func(cfg *aiven.ElasticSearchACLConfig) {
		cfg.Enabled = enabled
		cfg.ExtendedAcl = extednedACL
	}
}
