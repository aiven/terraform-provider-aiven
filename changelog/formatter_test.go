package main

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSample = `# Changelog

<!-- Always keep the following header in place: -->
<!-- ## [MAJOR.MINOR.PATCH] - YYYY-MM-DD -->

## [MAJOR.MINOR.PATCH] - YYYY-MM-DD

- Fix  aiven_project : can't migrate from  account_id  (deprecated) to  parent_id
- Add  aiven_organization_user_list  beta resource
- Add AIVEN_ALLOW_IP_FILTER_PURGE environment variable to allow purging of IP filters. This is a safety feature to
  prevent accidental purging of IP filters, which can lead to loss of access to services. To enable purging, set the
  environment variable to any value before running Terraform commands.
- Use TypeSet for ip_filter_object
- Deprecated account_id in aiven_project and aiven_billing_group resources
  - Please use parent_id instead, account_id is going to be removed in the next major release
- Fix incorrect behavior of aiven_mirrormaker_replication_flow schema fields:
  - sync_group_offsets_enabled
  - sync_group_offsets_interval_seconds
- Fix project creation with account_id empty and add possibility to dissociate project from an account by not
  setting account_id
- Fix typos in documentation and examples

## [4.26.0] - 2024-09-25

- Remove  aiven_valkey  from beta resources
- Remove  aiven_valkey_user  from beta resources
- Adds aiven_organization_permission  example
- Add
  capability to map external service user with internal aiven user with external_identity data source

## [1.0.0] - 2018-09-27

Support all Aiven resource types. Also large changes to previously
supported resource types, such as full support for all user config
options.
`

const testSampleExpected = `# Changelog

<!-- Always keep the following header in place: -->
<!-- ## [MAJOR.MINOR.PATCH] - YYYY-MM-DD -->

## [MAJOR.MINOR.PATCH] - YYYY-MM-DD

- Fix aiven_project : can't migrate from account_id
  (deprecated) to parent_id
- Add aiven_organization_user_list beta resource
- Add AIVEN_ALLOW_IP_FILTER_PURGE environment variable
  to allow purging of IP filters. This is a safety feature
  to prevent accidental purging of IP filters, which can lead
  to loss of access to services. To enable purging,
  set the environment variable to any value before running
  Terraform commands.
- Use TypeSet for ip_filter_object
- Deprecated account_id in aiven_project
  and aiven_billing_group resources
  - Please use parent_id instead, account_id is going
    to be removed in the next major release
- Fix incorrect behavior of aiven_mirrormaker_replication_flow
  schema fields:
  - sync_group_offsets_enabled
  - sync_group_offsets_interval_seconds
- Fix project creation with account_id empty
  and add possibility to dissociate project from an account
  by not setting account_id
- Fix typos in documentation and examples foo bar

## [4.26.0] - 2024-09-25

- Remove aiven_valkey from beta resources
- Remove aiven_valkey_user from beta resources
- Adds aiven_organization_permission example
- Add capability to map external service user with internal
  aiven user with external_identity data source

## [1.0.0] - 2018-09-27

- Support all Aiven resource types. Also large changes
  to previously supported resource types, such as full support
  for all user config options.
`

func TestUpdateChangelog(t *testing.T) {
	result, err := updateChangelog(testSample, 60, true, "foo", "bar", "Use TypeSet for ip_filter_object")
	require.NoError(t, err)
	assert.Empty(t, cmp.Diff(testSampleExpected, result))
}

func TestUpdateChangelog_nothing_to_change(t *testing.T) {
	sample := `---
title: Changelog
parent: README
nav_order: 1
---

# Changelog

<!-- Always keep the following header in place: -->
<!--## [MAJOR.MINOR.PATCH] - YYYY-MM-DD -->

## [4.29.0] - 2024-11-14

- Add support for autoscaler service integration
`
	result, err := updateChangelog(sample, 60, false)
	require.NoError(t, err)
	assert.Empty(t, cmp.Diff(sample, result))
}

func TestUpdateChangelog_empty_changelog(t *testing.T) {
	sample := `---
title: Changelog
parent: README
nav_order: 1
---

# Changelog

<!-- Always keep the following header in place: -->
<!--## [MAJOR.MINOR.PATCH] - YYYY-MM-DD -->

## [4.29.0] - 2024-11-14

- Add support for autoscaler service integration
`
	expect := `---
title: Changelog
parent: README
nav_order: 1
---

# Changelog

<!-- Always keep the following header in place: -->
<!--## [MAJOR.MINOR.PATCH] - YYYY-MM-DD -->

## [MAJOR.MINOR.PATCH] - YYYY-MM-DD

- Foo
- Bar

## [4.29.0] - 2024-11-14

- Add support for autoscaler service integration
`
	result, err := updateChangelog(sample, 60, false, "Foo", "Bar")
	require.NoError(t, err)
	assert.Empty(t, cmp.Diff(expect, result))
}

func TestLineWrapping(t *testing.T) {
	input := `Add capability to map external service user with internal aiven user with external_identity data source`
	expectList := []string{
		"Add capability",
		"to map external service",
		"user with internal aiven",
		"user with",
		"external_identity data source",
	}
	expectPoint := "- Add capability\n  to map external service\n  user with internal aiven\n  user with\n  external_identity data source"
	assert.Equal(t, expectList, softWrap(input, 25))
	assert.Equal(t, expectPoint, addBullet("- ", expectList))
}

func TestChangelogFile(t *testing.T) {
	b, err := os.ReadFile("../CHANGELOG.md")
	require.NoError(t, err)

	result, err := updateChangelog(string(b), 80, true)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}
