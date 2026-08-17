---
name: tf-resource-migration
description: Migrate existing Terraform resources from SDK provider (sdkprovider) to Plugin Framework. Use when converting legacy resources, ensuring state compatibility, or when the user asks to port/migrate existing resources.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write, Task
---

# Terraform Resource Migration Skill

Migrate existing SDK-based resources to Plugin Framework while maintaining state compatibility and behavior parity.

## Tooling Rule (read first)

Always drive generation, formatting, and linting through Task:

- `task generate` — generates code and runs `task fmt` automatically at the end
- `task fmt` — formats Go, Terraform, and whitespace
- `task lint` — runs all linters

Do NOT use `go run ./generators/...`, `go generate`, `gofmt`, `goimports`, `golangci-lint`, or `make` for these workflows. See `AGENTS.md` and `Taskfile.dist.yml` (run `task --list`) for the full command surface.

## Overview

This skill guides migration of resources from:
- **Source**: `internal/sdkprovider/` (terraform-plugin-sdk/v2)
- **Target**: `internal/plugin/` (terraform-plugin-framework) with YAML-generated code

**Key Challenge**: Preserve exact behavior and state compatibility so users don't experience breaking changes.

**Prerequisites**: This skill builds on **tf-resource-generator**. For YAML syntax, adapter API, modifier patterns, custom view overrides, write-only fields, and all implementation details, see that skill.

## When to Use This Skill

**Use this skill when:**
- Migrating existing `aiven_*` resources from SDK to Plugin Framework
- User says: "migrate", "convert", "port", "move to Plugin Framework"

**Use tf-resource-generator skill when:**
- Creating brand new resources from scratch
- Questions about YAML syntax, adapter API, or implementation patterns

## Migration Workflow

### 1. Analyze the Existing SDK Resource

**Find the SDK resource:**
```bash
# Find resource file
find internal/sdkprovider -name "*resource_*.go" | grep -i "resource_name"

# Find data source file
find internal/sdkprovider -name "*datasource_*.go" | grep -i "resource_name"
```

**Read and document:**
- Schema definition (all fields, types, attributes)
- CRUD functions (Create, Read, Update, Delete)
- Custom logic and transformations
- State upgrade functions (if any)
- Existing tests (critical for parity)

### 2. Identify API Operations

Check what API operations the SDK resource uses:
```bash
# Search for API client calls in the resource
grep -A 5 "client\." internal/sdkprovider/service/resource_name.go

# IMPORTANT: Also check the data source — it may use a different API operation
grep -A 5 "client\." internal/sdkprovider/service/resource_name_data_source.go
```

Then find corresponding OpenAPI operation IDs. See **tf-resource-generator** for OpenAPI search patterns.

**Determine `clientHandler`**: The `clientHandler` YAML value is the Go package name under `github.com/aiven/go-client-codegen/handler/`. Find it by searching for the operation ID in the module cache:
```bash
grep -r "OperationID" $(go env GOMODCACHE)/github.com/aiven/go-client-codegen*/handler/
```
For example, `ServiceFlinkCreateApplication` lives in `handler/flinkapplication/`, so `clientHandler: flinkapplication`.

### 3. Map SDK Schema to YAML

#### SDK Type -> YAML Type

| SDK Type | YAML Type |
|----------|-----------|
| `schema.TypeString` | `type: string` |
| `schema.TypeInt` | `type: integer` |
| `schema.TypeFloat` | `type: number` |
| `schema.TypeBool` | `type: boolean` |
| `schema.TypeList` | `type: arrayOrdered` |
| `schema.TypeSet` | `type: array` (or `arrayOrdered` for performance) |
| `schema.TypeMap` | `additionalProperties: {type: string}` |

#### SDK Attributes -> YAML Attributes

| SDK Attribute | YAML Attribute |
|---------------|----------------|
| `Required: true` | `required: true` |
| `Optional: true` | `optional: true` |
| `Computed: true` | `computed: true` |
| `Sensitive: true` | `sensitive: true` |
| `ForceNew: true` | `forceNew: true` |
| `ConflictsWith: []` | `conflictsWith: []` |
| `ExactlyOneOf: []` | `exactlyOneOf: []` |

#### Nested Blocks

**SDK Set of objects:**
```go
"tags": {
    Type:     schema.TypeSet,
    Elem: &schema.Resource{
        Schema: map[string]*schema.Schema{
            "key":   {Type: schema.TypeString},
            "value": {Type: schema.TypeString},
        },
    },
}
```

**YAML (use arrayOrdered for performance):**
```yaml
schema:
  tags:
    type: arrayOrdered
    items:
      type: object
      properties:
        key:
          type: string
        value:
          type: string
```

#### Read-Only Nested Collections

**Do not reshape a field to fix a diff.** A list stays a list and a set stays a set, which is
what keeps an existing state readable.

Nested collections are the one exception, and the generator decides it: a collection that is
read-only all the way down becomes a **computed nested attribute**, everything else a **block**.
No YAML key asks for this — `Item.RendersAsAttribute` derives it.

The Plugin Framework has no computed blocks: Terraform plans a block from the configuration, so
a value stored for a block the configuration doesn't declare either shows up as a permanent
diff or fails the apply with an inconsistent result. An attribute can be computed, and a
read-only collection gives up nothing by losing the block body — it has no field to write in
it. The two are interchangeable for readers: same `list(object({...}))` in the state, same
expressions (`app.application_versions[0].id`), so no state upgrade is involved.

An `object` stays a one-element collection in both forms, as the SDKv2 state holds it. That is
**due for removal in v5.0.0**, when `current_deployment.0.status` becomes
`current_deployment.status`. Never anticipate it in a migration: it breaks every reader.

**Caveat — this holds only when nothing inside is settable.** An `optional: true` +
`computed: true` collection such as the kafka topic `config` stays a block, because real
configurations write it in block syntax. Marking one nested field `optional: true` is enough to
turn the whole collection back into a block. Data sources keep blocks too: their result is not
planned against a configuration, so a block holding API values causes no diff there.

For an optional+computed block, drop the values the user did not set instead — see
`flattenConfig` in `internal/plugin/service/kafka/topic/topic.go`. `Flatten` writes back only
the keys the response carries, so `delete(dto, name)` leaves whatever prior state holds;
`d.Set(name, nil)` removes it.

Generated example: `application_versions` in
`internal/plugin/service/flink/jarapplication/zz_resource.go`.

### 4. Preserve ID Structure

**CRITICAL**: The ID format MUST stay the same for state compatibility.

**Find the ID format in SDK code:**
```bash
# Look for ResourceData.SetId calls
grep -A 2 "SetId" internal/sdkprovider/service/resource_name.go

# Look for ID builder functions
grep -B 5 "buildResourceID\|parseResourceID" internal/sdkprovider/service/resource_name.go
```

**Common ID patterns:**
- Single field: `project`
- Composite: `project/service_name/database_name`

**Set in YAML:**
```yaml
idAttributeComposed: [project, service_name, database_name]
```

### 5. Handle Custom Logic

**Identify in SDK code and map to generator features:**

| SDK Pattern | Generator Feature |
|-------------|------------------|
| `StateUpgraders` | May need `version` in YAML + state upgrader |
| `CustomizeDiff` | `modifyPlan: true` (writes plan values, `d.RequiresReplace` for replacement) — not `planModifier`, which runs inside Read |
| Flatten/Expand functions | `expandModifier: true` / `flattenModifier: true` |
| `DiffSuppressFunc` | No equivalent, see [Diff Suppression Has No Equivalent](#diff-suppression-has-no-equivalent) |
| Post-Create/Update waiter (poll Read until the resource converges) | `refreshStateExists`, or `stateAttribute` with `refreshStateDesired`/`refreshStateFailed`, or `ResourceOptions.RefreshStateCheck` when the value is nested (see tf-resource-generator) |
| Multiple API calls in Update | Custom `updateView` via `init()` override |
| Delete waiter (poll until gone/terminal state; conflicts clear as Delete is re-issued) | `deleteStateDesired` (see tf-resource-generator); avoid a custom `deleteView` |
| Complex delete (cancel + delete state machine) | Custom `deleteView` via `init()` override |
| Data source looks up by alt key (e.g. name) | A second `read` op with `datasourceLookup: true` + `resultListLookupKeys`; add `resultIDField: <GoField>` when the lookup endpoint differs in shape from the canonical read and should only resolve the id |
| Sensitive field not stored in state | `writeOnly: true` |

For implementation details of each, see **tf-resource-generator** skill.

#### Diff Suppression Has No Equivalent

`DiffSuppressFunc` cannot be ported. Terraform rejects a plan whose value for a configured attribute
differs from the configuration, which is why the Plugin Framework offers no such hook: SDKv2 only got
away with it through its legacy type-system shims. Decide what the suppression was buying:

- **It hid a change that needs no API call** (e.g. a local file path the API never sees). Set
  `forceNew: false` on the attribute and let the change apply in place. No update view is needed:
  the adapter skips a nil `Update` and refreshes the state, so the new value is simply stored.
- **It hid a formatting difference** the API normalizes (casing, trailing slash). Normalize the same
  way in `modifyPlan` — but only for computed attributes, since a configured one must keep the
  configured value.
- **It paired with a computed attribute that drove replacement** (`CustomizeDiff` + `SetNew` on a
  `ForceNew` computed field). Compute the value in `modifyPlan` and call `d.RequiresReplace` on it.

Whichever applies, the behavior changes for at least some configurations, so call it out in
`CHANGELOG.md`. `internal/plugin/service/flink/jarversion` migrated a resource that used both a
`DiffSuppressFunc` and `CustomizeDiff`.

### 6. Create YAML Definition

Create `definitions/aiven_resource_name.yml`. For complete YAML syntax reference, see **tf-resource-generator** skill.

**IMPORTANT**: Definition files MUST have the `aiven_` prefix. The filename becomes the resource name directly.

Focus on migration-specific concerns:
- Match all SDK schema fields exactly
- Preserve ID structure
- Copy descriptions from SDK resource
- Keep every SDK `Sensitive` field as `sensitive: true` in YAML (no regressions)

### 7. Generate and Build

```bash
task generate
task build
task lint
```

### 8. State Compatibility Verification

**CRITICAL**: Ensure state is compatible between SDK and Plugin Framework versions.

**Check schema version in SDK:**
```bash
grep -A 3 "SchemaVersion" internal/sdkprovider/service/resource_name.go
```

If SDK has `SchemaVersion > 0`, you MUST handle state upgrades.

### 9. Backward Compatibility Testing

**CRITICAL**: Test that existing state from SDK version works with Plugin Framework version.

Use `acc.BackwardCompatibilitySteps()` helper:

```go
func TestAccAivenResource_backwardCompat(t *testing.T) {
    resourceName := "aiven_resource_name.test"
    projectName := acc.ProjectName()

    resource.ParallelTest(t, resource.TestCase{
        PreCheck: func() { acc.TestAccPreCheck(t) },
        Steps: acc.BackwardCompatibilitySteps(t, acc.BackwardCompatConfig{
            TFConfig:           testAccResourceConfig(projectName),
            OldProviderVersion: "4.47.0", // Check CHANGELOG.md for latest
            Checks: resource.ComposeTestCheckFunc(
                resource.TestCheckResourceAttr(resourceName, "project", projectName),
                // Add all key attribute checks
            ),
        }),
    })
}
```

**Find the latest version:**
```bash
head -20 CHANGELOG.md
```

**What this test does:**
1. Creates resource with OLD SDK provider version
2. Applies with NEW Plugin Framework version
3. Verifies state is compatible and attributes match

**Examples**:
- `internal/plugin/service/mysql/database/database_test.go` - Basic backward compatibility
- `internal/plugin/service/pg/user/user_test.go` - Complex resource with custom update logic

### 10. Parity Testing

**CRITICAL**: Verify behavior matches SDK resource exactly.

**Find SDK tests:**
```bash
ls internal/sdkprovider/service/*resource_name*_test.go
```

**Ensure Plugin Framework tests cover:**
- All CRUD operations from SDK tests
- Update step: If the resource implements `Update` (mutable fields that are not all `ForceNew`), acceptance tests MUST include a Terraform apply that changes at least one updatable attribute after create — i.e. an explicit `update` step in the test sequence, not only create + destroy. Skip this only when the resource is create/delete-only or every mutable change forces replacement.
- Edge cases
- Error handling
- Import functionality
- Special field behaviors

## State Compatibility Checklist

Before marking migration complete:

- [ ] Resource ID format is identical
- [ ] All schema fields are present (no removals)
- [ ] Field types match exactly, and no optional field turned from a block into an attribute
- [ ] Computed fields work the same way
- [ ] Default values match
- [ ] Required/Optional flags match
- [ ] All SDK-sensitive fields remain `sensitive: true` in YAML
- [ ] ForceNew behavior matches
- [ ] Import works with existing IDs
- [ ] Existing state can be used without migration
- [ ] Backward compatibility test added using `acc.BackwardCompatibilitySteps()`
- [ ] All SDK test scenarios pass with Plugin version
- [ ] When the resource supports Update, acceptance tests include an update step (apply with changed config after create)
- [ ] Changelog entry added to `CHANGELOG.md`

## Common Migration Issues

| Issue | Solution |
|-------|----------|
| Sensitive field no longer marked sensitive | Set `sensitive: true` on the attribute (and nested fields if applicable); match SDK `Sensitive: true` exactly |
| ID format changed accidentally | Verify `idAttributeComposed` matches SDK's ID builder |
| Set ordering causes diffs | Use `arrayOrdered` instead of `array` |
| Read-only list of objects diffs on every plan | It must render as a computed attribute; check nothing inside it is `optional` (see Read-Only Nested Collections) |
| Computed field becomes required | Keep as `computed: true` if API provides it |
| Custom validation lost | Implement in custom modifier or use schema validation |
| State upgrade needed | Implement state upgrader in Plugin Framework |
| DiffSuppressFunc behavior | Not portable — see [Diff Suppression Has No Equivalent](#diff-suppression-has-no-equivalent) |
| Field the API accepts but never returns | Fill it in a `readView` override only when the state has no value, so imports don't keep a null and existing state stays untouched |
| OpenAPI default would plan a change against SDKv2 state | Workaround: `dropDefault: true` plus `useStateForUnknown: true`, and send the API default from the create view. Goes away in v5.0.0 with SDKv2 state support |
| Computed value must force replacement | Compute it in `modifyPlan` and call `d.RequiresReplace(name)`; schema `forceNew` runs too early to see it |
| Post-create waiter polls a nested attribute | `refreshStateExists: true` plus `ResourceOptions.RefreshStateCheck`; `stateAttribute` only accepts top-level attributes |
| Renamed ID field missing in old state | Use `planModifier: true` to extract from composite ID |
| Read fails with 404 after migration | Likely a renamed ID field is empty — use `planModifier` |
| "was null, but now cty.X" error | Add `computed: true` + `useStateForUnknown: true` (see generator skill) |
| Field nested differently in API | Use `expandModifier` and `flattenModifier` (see generator skill) |
| Multiple update operations needed | Override `updateView` via `init()` (see generator skill) |
| Data source lookup key differs | Add a second `read` operation with `datasourceLookup: true` + `resultListLookupKeys`. If the lookup endpoint returns a different shape than the canonical read (e.g. a directory of ids), also set `resultIDField: <GoField>` (see generator skill) |

## Migration-Specific Commands

```bash
# Find SDK resource
find internal/sdkprovider -name "*resource_*.go" | grep -i "name"

# Analyze SDK schema
grep -A 20 "Schema:" internal/sdkprovider/service/resource.go

# Find SDK ID format
grep -A 2 "SetId" internal/sdkprovider/service/resource.go

# Compare implementations
diff internal/sdkprovider/service/resource.go internal/plugin/service/resource/zz_resource.go

# Run backward compatibility test
task test-acc -- -run TestAccAivenResource_backwardCompat
```

## After Migration

Once all tests pass and state compatibility is verified:

1. **Remove SDK resource** - Delete from `internal/sdkprovider/` and remove provider registration
2. **Update documentation** - Ensure docs reflect the Plugin Framework version
3. **Add migration notes if needed** - Document any unavoidable behavioral differences
4. **Add changelog entry** - Add record to `CHANGELOG.md` under the unreleased section:
   ```markdown
   - Migrate `aiven_resource_name` to the Plugin Framework
   - Change `aiven_resource_name`: deprecate `termination_protection` field. Instead, use [prevent_destroy](https://developer.hashicorp.com/terraform/tutorials/state/resource-lifecycle#prevent-resource-deletion)
   ```

**Do not maintain both versions** - this creates maintenance burden and user confusion.

## Key Principles

1. **State compatibility first** - Users should not need to recreate resources
2. **Preserve exact behavior** - Match SDK resource behavior precisely
3. **Preserve sensitivity** - All fields that were `Sensitive` in SDK must stay sensitive in Plugin Framework (`sensitive: true`); never expose secrets in plan or state output by omission
4. **Test thoroughly** - All SDK test scenarios must pass with Plugin version; when Update exists, tests must exercise it via an update apply step
5. **Remove SDK version** - Once verified, delete SDK resource to avoid maintenance burden
