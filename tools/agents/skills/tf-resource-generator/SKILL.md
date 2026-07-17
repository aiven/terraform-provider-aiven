---
name: tf-resource-generator
description: Generate new Terraform resources and data sources from YAML definitions and OpenAPI specs. Use when adding resources, creating YAML definitions, working with the Plugin Framework generator, or when the user asks about code generation.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write, Task
---

# Terraform Resource Generator Skill

Generate new Terraform resources and data sources by learning from existing patterns in the codebase.

## Tooling Rule (read first)

Always drive generation, formatting, and linting through Task:

- `task generate` — generates code and runs `task fmt` automatically at the end
- `task fmt` — formats Go, Terraform, and whitespace
- `task lint` — runs all linters

Do NOT use `go run ./generators/...`, `go generate`, `gofmt`, `goimports`, `golangci-lint`, or `make` for these workflows. See `AGENTS.md` and `Taskfile.dist.yml` (run `task --list`) for the full command surface.

## Overview

The generator creates Terraform resources by combining:
1. **OpenAPI Spec** (`openapi.json`) - Aiven's API schema
2. **YAML Definitions** (`definitions/aiven_*.yml`) - Configuration for resource generation

**Note**: This skill is the canonical reference for YAML syntax, adapter API, and implementation patterns. The **tf-resource-migration** skill builds on this for migrating existing SDK resources.

**Generated outputs per resource:**
- `zz_resource.go` / `zz_datasource.go` - Terraform schema + internal schema
- `zz_view.go` - CRUD operation handlers, `ResourceOptions`, `DataSourceOptions`
- `examples/resources/*/import.sh` - Import command examples

**Generated aggregated output:**
- `internal/plugin/zz_provider.go` - Provider resource/datasource registry

## User Interaction Pattern

When the user requests a new resource, gather requirements BEFORE starting:

### First Question (REQUIRED)

**ALWAYS start by asking what type of Terraform component is needed:**

Use the `AskUserQuestion` tool with these options:
- **Both resource and data source** - Full CRUD resource + read-only data source
- **Resource only** - Just the resource (create, read, update, delete)
- **Data source only** - Read-only data source

### Additional Information to Gather
1. **Resource name** - What API resource/endpoint? (e.g., "MySQL database", "Kafka topic")
2. **Scope** - Project-level, service-level, organization-level?

### Discovery Process
1. **Update OpenAPI spec** first by running `task get-spec` to ensure you have the latest API schema
2. **Search OpenAPI spec** for operation IDs
3. **Show the user** what you found and ask for confirmation
4. **Find similar resources** in `definitions/` and show them
5. **Clarify specifics** if needed (composite ID, fields to exclude/rename, custom modifiers)

### When NOT to Ask
- If the user provided specific operation IDs -> proceed
- If the user provided a complete YAML definition -> proceed
- If it's a modification to existing definition -> read the file and proceed

### Confirmation Before Generation
Always summarize what you're about to create before proceeding.

## Learning from Existing Resources

**CRITICAL**: Before creating a new resource, ALWAYS search for similar existing resources to learn patterns:

```bash
# Find all YAML definitions
ls definitions/aiven_*.yml

# Find resources with similar client handlers
grep -l "clientHandler: service" definitions/aiven_*.yml

# Find resources with custom modifiers
grep -l "expandModifier: true" definitions/aiven_*.yml

# Find composite ID examples
grep -l "idAttributeComposed:" definitions/aiven_*.yml
```

**Key files to reference:**
- `definitions/.schema.yml` - Complete YAML schema specification
- `definitions/aiven_*.yml` - Real working examples
- Existing generated code in `internal/plugin/service/*/zz_*.go`

## Generation Workflow

### 1. Update OpenAPI Spec

Always start by downloading the latest API schema:
```bash
task get-spec
```

### 2. Find the API Operations

Search OpenAPI spec for operation IDs:
```bash
jq '.paths | to_entries[] | .value | to_entries[] | .value.operationId' openapi.json | grep -i "keyword"
```

Common patterns: `ResourceCreate`, `ResourceGet`, `ResourceUpdate`, `ResourceDelete`, `ResourceList`

### 3. Find a Similar Resource

Search `definitions/` for a resource with similar characteristics:
- Same client handler (project, service, organization, etc.)
- Similar ID structure (single vs composite)
- Similar operations (CRUD, list-based reads, etc.)

Read that YAML file to understand the pattern.

### 4. Create Your YAML Definition

File: `definitions/aiven_my_resource.yml`

**IMPORTANT**: Definition files MUST have the `aiven_` prefix. The filename (without `.yml`) becomes the resource name directly: `aiven_my_resource.yml` -> resource `aiven_my_resource`.

Start minimal:
```yaml
# yaml-language-server: $schema=.schema.yml
location: internal/plugin/service/myresource
operations:
  - id: OperationIDFromOpenAPI
    type: create|read|update|delete
resource:
  description: "..."
idAttributeComposed: [field1, field2]
clientHandler: project
```

### 5. Generate and Iterate

```bash
task generate       # Generate code
task build          # Build to check for errors
task lint           # Check code quality
```

Review generated `zz_*.go` files. Refine YAML definition as needed.

## YAML Definition Reference

### Top-Level Fields

**IMPORTANT**: Always check `definitions/.schema.yml` for the complete, authoritative schema specification.

Common fields:
- `location` - Package path (e.g., `internal/plugin/service/mysql`)
- `operations` - Array of CRUD operations (id, type, resultKey, etc.)
- `resource` / `datasource` - Configuration metadata
- `idAttributeComposed` - Fields that compose the ID (e.g., `[project, service_name]`)
- `clientHandler` - API client type: `project`, `service`, `organization`, `organizationbilling`, etc.
- `remove` - Fields to exclude from schema
- `rename` - Field name mappings (API field -> Terraform field)
- `schema` - Schema customizations (types, validation, behavior)
- `expandModifier` / `flattenModifier` / `planModifier` - Enable custom Go modifiers
- `version` - Schema version (for state upgrades)
- `beta` - Mark resource as beta (requires `PROVIDER_AIVEN_ENABLE_BETA` env var)
- `limitedAvailability` - Mark resource as limited availability
- `legacyTimeouts` - Enable SDK v2-style timeout blocks

### Resource Configuration

```yaml
resource:
  description: "..."
  deprecationMessage: "..."
  terminationProtection: true # Check field before delete
  refreshState: true # Call Read after Create/Update
  refreshStateDelay: "15s" # Wait before Read
  refreshStateDesired: # Retry Read until fields match desired values
    state: ACTIVE
  removeMissing: true # Remove from state on 404
  deleteStateDesired: # Poll Read after Delete until the resource reaches its terminal state
    state: DELETED
```

`refreshStateDesired` replaces ad-hoc refresh waiters for simple state transitions. Use it when the resource should settle into a known value after create/update (for example, `state: ACTIVE`).

`deleteStateDesired` replaces custom delete waiters. It is a map, and its presence (even as an empty map, `{}`) enables the delete poller: on every attempt the adapter re-issues `Delete` (ignoring its error) and polls `Read` until the resource is gone (404) or reaches its terminal state, then returns. A 404 always completes the delete. Because `Delete` is re-issued each attempt and its error is ignored, a 409 Conflict from dependents still detaching resolves on its own — no extra configuration is needed.

The map's entries are attribute names mapped to the terminal value the poller must observe (e.g. `state: DELETED`). The poll finishes when every entry matches or the API 404s, whichever comes first. Values are validated against the field's enum when one is declared, as for `refreshStateDesired`. Use an empty map for APIs whose Read never surfaces a terminal state (it just 404s).

Examples:

```yaml
# Wait until the API 404s; no observable terminal state (e.g. aiven_azure_privatelink).
deleteStateDesired: {}
```

```yaml
# Soft-delete API reports state: DELETED, and dependents may still be detaching
# (e.g. aiven_project_vpc, aiven_organization_vpc). Delete is re-issued each attempt.
deleteStateDesired:
  state: DELETED
```

Prefer `deleteStateDesired` over a hand-written delete `disableView` + custom poller whenever the terminal condition is "gone (404)" and/or a fixed attribute value. Keep a custom delete view only for genuinely bespoke flows (e.g. a cancel-then-delete state machine).

### Datasource Configuration

```yaml
datasource:
  description: "..."
  deprecationMessage: "..."
```

For alternative lookup keys (e.g. look up by `name` instead of `id`), see [Data Source with Alternative Lookup Key](#data-source-with-alternative-lookup-key) below.

### Operations Configuration

Each operation maps a Terraform action to an OpenAPI operation ID:

```yaml
operations:
  - id: OperationIDFromOpenAPI
    type: create|read|update|delete
    disableView: true               # Don't generate view function (for custom override)
    datasourceLookup: true          # Inline this read op as readView's id-empty branch on
                                    # the data source (requires resultListLookupKeys).
                                    # See "Data Source with Alternative Lookup Key" below.
    resultIDField: OrganizationId   # Go field on the lookup result that holds the
                                    # primary id. When set, the lookup only resolves
                                    # the id, then control falls through to the canonical
                                    # read body in the same readView. Requires
                                    # datasourceLookup + resultListLookupKeys.
    resultKey: nested_field          # Extract from response.nested_field (schema/JSON path,
                                     # used at generation time to scope the OpenAPI schema)
    resultKeyField: GoField          # Go field name on the client response to drill into at
                                     # runtime, e.g. d.Flatten(rsp.GoField). Use when the Go
                                     # client doesn't strip an extra wrapper exposed by the API
                                     # (response is {accessors: {aws: {...}}} but the client
                                     # returns *CMKAccessorsListOut, so set resultKeyField: Aws).
                                     # When combined with resultListLookupKeys, points at the
                                     # list to search (e.g. resultKeyField: ConnectionPools).
    resultToKey: wrapper_key         # Wrap response as {wrapper_key: ...}
    resultListLookupKeys:            # For list responses
      APIField: terraform_field      # Match items by field
```

**Find examples**: `grep -A 5 "operations:" definitions/aiven_*.yml`

### Schema Customization

Override generated schema fields:

```yaml
schema:
  field_name:
    type: string|integer|number|boolean|array|arrayOrdered|object
    description: "..."
    required: true|false
    computed: true|false
    sensitive: true|false
    writeOnly: true|false
    forceNew: true|false           # Triggers replacement
    useStateForUnknown: true|false # Preserve prior state during planning
    enum: [val1, val2]
    minimum: 1
    maximum: 100
    conflictsWith: [other_field]
    exactlyOneOf: [field_a, field_b]
    atLeastOneOf: [field_a, field_b]
    alsoRequires: [other_field]
    default: value
    deprecationMessage: "..."
```

**Important**: Use `arrayOrdered` (list) for complex objects, not `array` (set). Sets cause performance issues.

**Find examples**: `grep -A 10 "schema:" definitions/aiven_*.yml`

## Generated View Structure

The generator produces `zz_view.go` with this structure:

```go
const typeName = "aiven_my_resource"

func idFields() []string {
    return []string{"project", "service_name", "resource_name"}
}

var ResourceOptions = adapter.ResourceOptions{
    Create:            createView,
    Delete:            deleteView,
    IDFields:          idFields(),
    Read:              readView,
    RefreshState:      true,
    RefreshStateDelay: adapter.MustParseDuration("15s"),
    RefreshStateDesired: map[string]string{"state": "ACTIVE"},
    RemoveMissing:     true,
    Schema:            resourceSchema,
    SchemaInternal:    resourceSchemaInternal(),
    TypeName:          typeName,
}

var DataSourceOptions = adapter.DataSourceOptions{
    IDFields:       idFields(),
    Read:           readView,
    Schema:         datasourceSchema,
    SchemaInternal: datasourceSchemaInternal(),
    TypeName:       typeName,
}
```

**CRUD function signature** (all views follow this pattern):
```go
func createView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error
func readView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error
func updateView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error
func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error
```

### ResourceOptions Additional Fields

Beyond basic CRUD, `ResourceOptions` supports:
- `ModifyPlan` — implements `resource.ResourceWithModifyPlan` (separate from `planModifier` in YAML)
- `ValidateConfig` — implements `resource.ResourceWithValidateConfig`
- `ConfigValidators` — implements `resource.ResourceWithConfigValidators`

## ResourceData Interface

All views and modifiers interact with Terraform state through `adapter.ResourceData`:
```go
type ResourceData interface {
    Get(key string) any              // Get from plan, then config, then state
    GetOk(key string) (any, bool)    // Get with existence check
    GetState(key string) any         // Get specifically from state
    HasChange(key string) bool       // Plan value differs from state
    Set(key string, value any) error // Set a value in current state
    SetID(parts ...string) error     // Set composite ID
    ID() string                      // Get the "id" field
    IsNewResource() bool             // True if ID is empty
    Schema() *Schema                 // Access internal schema
    Expand(out any, modifiers ...MapModifier) error   // Plan -> API request
    Flatten(in any, modifiers ...MapModifier) error    // API response -> state
}
```

### Behavior by Operation

**CRITICAL**: `ResourceData` behaves differently depending on the CRUD operation:

- **Create**: `d.Get()` reads from plan, then config. No state available.
- **Update**: `d.Get()` reads from plan, then config, then state (for ID fields only). `d.GetState()` reads from prior state. `d.HasChange()` compares plan vs state.
- **Read/Delete**: `d.Get()` reads from state. No plan available.

This matters when writing modifiers:
- In `flattenModifier`, `d.Get("field")` returns the **current state** value (from plan during Create/Update, from state during Read)
- In `expandModifier`, `d.Get("field")` returns the **plan** value (what the user configured)
- Use `d.HasChange("field")` to detect if a field was modified (Update only)

## Custom Modifiers

For complex data transformations, enable modifiers in YAML:

```yaml
expandModifier: true      # TF state -> API request
flattenModifier: true     # API response -> TF state
planModifier: true        # Pre-process state before read API call
```

Then implement functions in a `.go` file in the resource package.

### expandModifier / flattenModifier

Transform data between Terraform state and API request/response formats.

**Signature:**
```go
func expandModifier(ctx context.Context, client avngen.Client) adapter.MapModifier
func flattenModifier(ctx context.Context, client avngen.Client) adapter.MapModifier
```

Where `MapModifier` is:
```go
type MapModifier func(d ResourceData, dto map[string]any) error
```

- `d` — the `adapter.ResourceData` interface for accessing plan/state/config values
- `dto` — the raw map being sent to or received from the API

**Example — nested API field** (`pg_allow_replication` stored as `access_control.pg_allow_replication` in API but exposed top-level in TF):
```go
func expandModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
    return func(d adapter.ResourceData, dto map[string]any) error {
        if v, ok := d.GetOk("pg_allow_replication"); ok {
            dto["access_control"] = map[string]any{"pg_allow_replication": v}
            delete(dto, "pg_allow_replication")
        }
        return nil
    }
}

func flattenModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
    return func(d adapter.ResourceData, dto map[string]any) error {
        if v, ok := dto["access_control"]; ok {
            dto["pg_allow_replication"] = v.(map[string]any)["pg_allow_replication"]
            delete(dto, "access_control")
        }
        return nil
    }
}
```

**Composing multiple modifiers** with `adapter.ComposeMapModifiers()`:
```go
func expandModifier(ctx context.Context, client avngen.Client) adapter.MapModifier {
    return adapter.ComposeMapModifiers(
        getFullCardID(ctx, client),
        expandParentID(ctx, client),
        ExpandEmails("billing_emails"),
    )
}
```

**Find examples**: `grep -l "flattenModifier\|expandModifier" internal/plugin/service/**/*.go`

### planModifier

Runs at the start of the generated `readView`, **before** the API call. Use it to pre-process or fix up state so the read has all the data it needs.

**Signature:**
```go
func planModifier(ctx context.Context, client avngen.Client, d adapter.ResourceData) error
```

**When to use:**
- A renamed ID field (`rename: {id: deployment_id}`) isn't in old SDK state — extract it from the composite `id`
- A field needs to be resolved before the read (e.g., name -> ID lookup via API)
- Any state pre-processing required before the generated read API call

**How it works:** Setting `planModifier: true` in YAML causes the generator to insert `planModifier(ctx, client, d)` at the top of `readView`:
```go
// generated readView
func readView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
    err := planModifier(ctx, client, d)  // <-- inserted by generator
    if err != nil {
        return err
    }
    // ... API call using d.Get() fields
}
```

**Reference implementations:**
- `internal/plugin/service/billinggroup/billinggroup.go` — extracts `billing_group_id` from composite `id` for SDK backward compat
- `internal/plugin/service/flink/deployment/deployment.go` — extracts `deployment_id` from composite `id` for SDK backward compat

For name -> ID resolution in data sources, use the generated `datasourceLookup` pattern instead of a hand-written `planModifier`. See [Data Source with Alternative Lookup Key](#data-source-with-alternative-lookup-key).

## Custom View Overrides

The generator produces standard CRUD views, but you can override **any** of them via `init()` when the generated logic is insufficient. Common reasons:

- **Create** needs extra steps (e.g., reset password after user creation)
- **Update** touches multiple API operations for different fields
- **Delete** requires a multi-step state machine (e.g., cancel then delete)
- **Read** needs post-processing not expressible via modifiers

Use `disableView: true` on the operation in YAML to skip generating the view you plan to override.

**Override pattern:**
```go
func init() {
    ResourceOptions.Create = createView   // Custom create
    ResourceOptions.Update = updateView   // Custom update
    ResourceOptions.Delete = deleteView   // Custom delete
}
```

**Example — custom update with multiple API operations** (`pg_user`):
```go
func updateView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
    if d.HasChange("pg_allow_replication") {
        req := &service.ServiceUserCredentialsModifyIn{
            Operation: service.ServiceUserCredentialsModifyOperationTypeSetAccessControl,
            AccessControl: &service.AccessControlIn{
                PgAllowReplication: new(d.Get("pg_allow_replication").(bool)),
            },
        }
        _, err := client.ServiceUserCredentialsModify(ctx,
            d.Get("project").(string),
            d.Get("service_name").(string),
            d.Get("username").(string), req)
        if err != nil {
            return err
        }
    }
    return resetPassword(ctx, client, d)
}
```

**Example — custom delete with state machine** (`flink_application_deployment`):
```go
func deleteView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
    project := d.Get("project").(string)
    serviceName := d.Get("service_name").(string)
    applicationID := d.Get("application_id").(string)
    deploymentID := d.Get("deployment_id").(string)

    for {
        _, err := client.ServiceFlinkGetApplicationDeployment(ctx, project, serviceName, applicationID, deploymentID)
        if avngen.IsNotFound(err) {
            return nil
        }
        _, err = client.ServiceFlinkCancelApplicationDeployment(ctx, project, serviceName, applicationID, deploymentID)
        if err != nil {
            _, _ = client.ServiceFlinkDeleteApplicationDeployment(ctx, project, serviceName, applicationID, deploymentID)
        }
        select {
        case <-ctx.Done():
            return fmt.Errorf("can't delete: %w", ctx.Err())
        case <-time.After(time.Second):
            continue
        }
    }
}
```

**Reference**: See `internal/plugin/service/flink/deployment/deployment.go` and `internal/plugin/service/pg/user/user.go`.

## Write-Only Fields

Write-only fields are values sent to the API but **never stored in Terraform state**. Use `writeOnly: true` in YAML for any field where persisting the value in state is undesirable — passwords, tokens, API keys, configuration payloads, etc.

**How it works:**
- Write-only fields are automatically excluded from data sources
- `ResourceData.Get()` reads write-only fields from **config** (not plan or state, since they're never stored)
- The generator marks them with `WriteOnly: true` in the Plugin Framework schema

**YAML:**
```yaml
schema:
  my_secret:
    type: string
    optional: true
    writeOnly: true
    sensitive: true
```

### Write-Only with Stored Variant

A common pattern offers both a regular field (stored in state) and a write-only variant (not stored), letting users choose. Applicable to passwords, tokens, or any sensitive credential.

**YAML:**
```yaml
schema:
  password:
    type: string
    optional: true
    computed: true
    sensitive: true
    conflictsWith: [password_wo]

  password_wo:
    type: string
    optional: true
    writeOnly: true
    sensitive: true
    conflictsWith: [password]
    alsoRequires: [password_wo_version]

  password_wo_version:
    type: integer
    optional: true
    alsoRequires: [password_wo]
    description: Increment to rotate password_wo.
```

**Key points:**
- `conflictsWith` ensures the user picks one approach or the other
- `password_wo_version` acts as a trigger — incrementing it forces a change even though the write-only value itself isn't tracked in state
- In `flattenModifier`, clear the regular field from state when the write-only variant is active
- In create/update logic, read the write-only field via `d.Get()` (falls through to config)

**Reference**: See `internal/plugin/service/pg/user/user.go` for a complete implementation.

## Data Source with Alternative Lookup Key

When a data source should look up by a field that's NOT in `idAttributeComposed` (e.g., lookup by `name` but resource ID uses `application_id`), add a second `read` operation marked `datasourceLookup: true` with `resultListLookupKeys`. The generator inlines the lookup body into `readView` as an id-empty branch:

```go
func readView(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
    // planModifier (if enabled) runs first
    if d.Get("<id-attr>").(string) == "" {
        // ... datasourceLookup list call + FindOne by resultListLookupKeys ...
        return d.Flatten(&match, ...)         // default: flatten match into state and return
        // OR (when resultIDField is set):
        // d.Set("<id-attr>", match.<resultIDField>)  // resolve id, fall through to canonical read
    }
    // ... canonical read API call by id ...
}
```

The data source uses the same `readView` directly — no separate `dataReadView` is generated. The generator also wires `ExactlyOneOf(id, alt_field...)` (and `RequiredTogether` when `composedOf` has more than one field) on the data source config validators, and marks the alternative fields as optional in the data source schema.

**YAML:**
```yaml
clientHandler: flinkapplication
idAttributeComposed: [project, service_name, application_id]
rename:
  id: application_id
operations:
  - id: ServiceFlinkCreateApplication
    type: create
  - id: ServiceFlinkGetApplication
    type: read
  - id: ServiceFlinkListApplications
    type: read
    datasourceLookup: true
    resultKey: applications        # Unwrap the response: rsp.applications -> []Application
    resultListLookupKeys:
      Name: name                   # API field "Name" matches TF attribute "name"
  - id: ServiceFlinkDeleteApplication
    type: delete
  - id: ServiceFlinkUpdateApplication
    type: update
```

The `rename` map (e.g. `id: application_id`) is forwarded to the lookup branch's `Flatten` call when the API field exists in the lookup response, so the resource id attribute is populated correctly from the list item.

**Multi-field composite lookup**: if `resultListLookupKeys` declares more than one mapping, the data source requires either the id attribute OR all alternative fields together. Example (`aiven_kafka_schema_registry_acl`):
```yaml
resultListLookupKeys:
  Permission: permission
  Resource: resource
  Username: username
```

### `resultIDField` — id resolution only

Set `resultIDField` to the Go field name on the lookup result item that holds the primary id when the lookup endpoint and the canonical read endpoint return _different_ shapes (the list response is just a directory of ids; the actual payload comes from the canonical read). With `resultIDField` set, `readView`'s id-empty branch:

1. Calls the `datasourceLookup` list op.
2. Finds the matching item by `resultListLookupKeys`.
3. Sets the resolved id on state via `d.Set(<id-attr>, match.<resultIDField>)`.
4. Falls through to the canonical read body in the same `readView` (no recursive call).

The lookup response is _not_ merged into the data source schema, so leaked fields from the directory endpoint cannot pollute it.

`resultIDField` requires `datasourceLookup: true` and `resultListLookupKeys` (enforced by `.schema.yml` `dependencies`).

```yaml
operations:
  - id: OrganizationUserList         # canonical read, fetches the actual payload
    type: read
    resultToKey: users
  - id: UserOrganizationsList        # directory used only to resolve the id
    type: read
    datasourceLookup: true
    resultIDField: OrganizationId    # Go field on the list item holding the id
    resultKey: organizations
    resultListLookupKeys:
      OrganizationName: name
rename:
  organization_id: id
```

**No `planModifier` needed**: both patterns fully replace the previous `datasource.exactlyOneOf` + custom `planModifier` approach.

**Reference implementations**:
- `definitions/aiven_flink_application.yml` — single-field name lookup (flatten match into state)
- `definitions/aiven_kafka_schema_registry_acl.yml` — three-field composite lookup
- `definitions/aiven_organizational_unit.yml` — name lookup with field renames
- `definitions/aiven_organization_user_list.yml` — `resultIDField`, lookup resolves id then falls through to the canonical read body

## Computed + Optional Pattern

Use `computed: true` + `optional: true` when the API always returns a value for a field, even when not configured. Without `computed`, you'll get:
```
Error: Provider produced inconsistent result after apply
.field_name: was null, but now cty.False
```

**Solution:**
```yaml
schema:
  field_name:
    type: boolean
    optional: true
    computed: true
    useStateForUnknown: true  # Preserves prior state during planning
```

**Trade-off**: Users who didn't configure the field will see a one-time cosmetic diff on upgrade (`null -> false`), but no resource recreation.

## Common Resource Patterns

Search the codebase for similar patterns:

| Pattern | Search Command |
|---------|---------------|
| Data source (list) | `grep -l "datasource:" definitions/aiven_*.yml \| xargs grep -l "resultToKey"` |
| Composite ID | `grep -l "idAttributeComposed:" definitions/aiven_*.yml \| head -3` |
| Custom modifiers | `grep -l "expandModifier: true" definitions/aiven_*.yml` |
| Field renaming | `grep -l "rename:" definitions/aiven_*.yml` |
| Alt data source lookup key | `grep -l "datasourceLookup: true" definitions/aiven_*.yml` |

**Best approach**: Find a similar resource, read its YAML definition, understand the pattern, adapt to your needs.

## Troubleshooting

| Error | Solution |
|-------|----------|
| `operationID X not found` | Search OpenAPI: `jq '.paths \| to_entries[] \| .value \| to_entries[] \| .value.operationId' openapi.json \| grep -i X` |
| `no 'id' field found` | Add `idAttributeComposed: [field1, field2]` to YAML |
| `ID field "X" not found` | Verify field exists in API response, check `rename` mappings |
| Generation fails | Run `task get-spec`, validate YAML syntax, check operation IDs |
| Infinite plan changes | Use `arrayOrdered` not `array`, ensure `idAttributeComposed` fields are stable |
| "was null, but now cty.X" | Add `computed: true` + `useStateForUnknown: true` to field |

## Reference Implementations

| Resource | Patterns Demonstrated |
|----------|----------------------|
| `pg_user` | Nested API fields, computed+optional, custom create/update, write-only fields, expand/flatten modifiers |
| `mysql_database` | Simple CRUD, list-based read with lookup, termination protection |
| `billing_group` | planModifier, `ComposeMapModifiers`, card ID resolution, organization ID conversion |
| `flink_application` | Alt data source lookup key via `datasourceLookup` read op, field rename |
| `flink_application_deployment` | Renamed ID field, planModifier for backward compat, custom delete with state machine |
| `kafka_schema_registry_acl` | Composite alt data source lookup (multi-field `datasourceLookup` + `resultListLookupKeys`) |
| `organization/unit` | Alt data source lookup via `datasourceLookup` read op, expand/flatten for parent ID |

## Testing Requirements

**CRITICAL**: Every generated resource MUST have acceptance tests covering:

### For Resources
1. **Create** - Basic resource creation
2. **Read** - Verify attributes are read correctly
3. **Update** - Modify attributes, verify changes
4. **Delete** - Implicit in test cleanup
5. **Import** - Test `terraform import` with correct ID format

**Example test structure:**
```go
func TestAccAivenMyResource_basic(t *testing.T) {
    resource.ParallelTest(t, resource.TestCase{
        PreCheck:                 func() { acc.TestAccPreCheck(t) },
        ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccMyResourceConfig("value1"),
                Check:  resource.TestCheckResourceAttr(resourceName, "field", "value1"),
            },
            {
                Config: testAccMyResourceConfig("value2"),
                Check:  resource.TestCheckResourceAttr(resourceName, "field", "value2"),
            },
            {
                ResourceName:      resourceName,
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}
```

**Find test examples**: `ls internal/plugin/service/*/mysql/*_test.go`

## Commands

```bash
task get-spec                          # Update OpenAPI spec
task generate                          # Generate code
task generate no_spec=true             # Generate without spec update
task build                             # Build provider
task lint                              # Run linters
task test-unit                         # Unit tests (no API)
task test-acc -- -run TestName         # Specific acceptance test
```

## Key Principles

1. **Learn from existing code** - Always find and read similar resources in `definitions/`
2. **Start minimal** - Basic YAML first, add complexity incrementally
3. **Use arrayOrdered** - For complex/large nested objects, always use `arrayOrdered` (list) not `array` (set) for performance
4. **Mark sensitive fields** - Use `sensitive: true` for passwords, tokens, API keys, credentials
5. **Follow conventions** - Match patterns from existing resources (naming, structure, etc.)
6. **Don't edit `zz_*` files** - These are generated. Custom logic goes in separate `.go` files

## Important Notes

- New resources go in `internal/plugin/` (Plugin Framework), NOT `internal/sdkprovider/` (legacy)
- Generated files have `zz_` prefix - never edit them manually
- Custom logic goes in separate `.go` files in the same package (e.g., `billinggroup.go`)
- Run `task lint` before committing
- Definition files MUST have `aiven_` prefix: `definitions/aiven_my_resource.yml` -> resource `aiven_my_resource`
- Files without `aiven_` prefix in `definitions/` are ignored by the generator

## Migrating Existing Resources

**If you need to migrate an existing SDK resource** from `internal/sdkprovider/` to Plugin Framework, use the **tf-resource-migration** skill instead. That skill covers:
- Analyzing existing SDK code
- SDK type -> YAML type mapping
- Preserving state compatibility
- Backward compatibility testing
- Deprecation strategy
