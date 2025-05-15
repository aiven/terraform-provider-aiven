package opensearch

import (
	"context"
	"sync"

	"github.com/aiven/aiven-go-client/v2"
)

// this mutex is needed to serialize calls to modify the remote config
// since it's an abstraction that first GETs, modifies and then PUTs again
var resourceOpenSearchACLModifierMutex sync.Mutex

// GETs the remote config, applies the modifiers and PUTs it again
// The Config that is passed to the modifiers is guaranteed to be not nil
func resourceOpenSearchACLModifyRemoteConfig(
	ctx context.Context,
	project string,
	serviceName string,
	client *aiven.Client,
	modifiers ...func(*aiven.OpenSearchACLConfig),
) error {
	resourceOpenSearchACLModifierMutex.Lock()
	defer resourceOpenSearchACLModifierMutex.Unlock()

	r, err := client.OpenSearchACLs.Get(ctx, project, serviceName)
	if err != nil {
		return err
	}

	config := r.OpenSearchACLConfig
	for i := range modifiers {
		modifiers[i](&config)
	}

	_, err = client.OpenSearchACLs.Update(
		ctx,
		project,
		serviceName,
		aiven.OpenSearchACLRequest{OpenSearchACLConfig: config})
	if err != nil {
		return err
	}
	return nil
}

// some modifiers

func resourceOpenSearchACLModifierUpdateACLRule(
	ctx context.Context,
	username string,
	index string,
	permission string,
) func(*aiven.OpenSearchACLConfig) {
	return func(cfg *aiven.OpenSearchACLConfig) {
		cfg.Add(resourceOpenSearchACLRuleMkAivenACL(username, index, permission))

		// delete the old acl if it's there
		if prevPerm, ok := resourceOpenSearchACLRuleGetPermissionFromACLResponse(*cfg, username, index); ok && prevPerm != permission {
			cfg.Delete(ctx, resourceOpenSearchACLRuleMkAivenACL(username, index, prevPerm))
		}
	}
}

func resourceOpenSearchACLModifierDeleteACLRule(
	ctx context.Context,
	username string,
	index string,
	permission string,
) func(*aiven.OpenSearchACLConfig) {
	return func(cfg *aiven.OpenSearchACLConfig) {
		cfg.Delete(ctx, resourceOpenSearchACLRuleMkAivenACL(username, index, permission))
	}
}

func resourceOpenSearchACLModifierToggleConfigFields(enabled, extednedACL bool) func(*aiven.OpenSearchACLConfig) {
	return func(cfg *aiven.OpenSearchACLConfig) {
		cfg.Enabled = enabled
		cfg.ExtendedAcl = extednedACL
	}
}
