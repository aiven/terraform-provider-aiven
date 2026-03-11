---
name: tf-resource-generator
description: Generate new Terraform resources and data sources from YAML definitions and OpenAPI specs. Use when adding resources, creating YAML definitions, working with the Plugin Framework generator, or when the user asks about code generation.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write, Task
---

# Terraform Resource Generator Skill

Generate new Terraform resources and data sources by learning from existing patterns in the codebase.

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
- `legacyTimeouts` - Enable SDK v2-style timeout blocks

### Resource Configuration

```yaml
resource:
  description: "..."
  deprecationMessage: "..."
  terminationProtection: true       # Check field before delete
  refreshState: true                # Call Read after Create/Update
  refreshStateDelay: "15s"          # Wait before Read
  removeMissing: true               # Remove from state on 404
```

### Datasource Configuration

```yaml
datasource:
  description: "..."
  deprecationMessage: "..."
  exactlyOneOf: [field1, field2]    # Generates ConfigValidators
```

### Operations Configuration

Each operation maps a Terraform action to an OpenAPI operation ID:

```yaml
operations:
  - id: OperationIDFromOpenAPI
    type: create|read|update|delete
    disableView: true               # Don't generate view function (for custom override)
    resultKey: nested_field          # Extract from response.nested_field
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
- `internal/plugin/service/flink/application/application.go` — resolves application name to ID via API call (for datasource lookup by name)
- `internal/plugin/service/organization/unit/unit.go` — resolves organization name to ID via API call

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

When a data source should look up by a field that's NOT in `idAttributeComposed` (e.g., lookup by `name` but resource ID uses `application_id`), use `datasource.exactlyOneOf` + `planModifier`:

**YAML:**
```yaml
datasource:
  exactlyOneOf: [application_id, name]
planModifier: true
idAttributeComposed: [project, service_name, application_id]
```

The `planModifier` resolves the alternative key to the ID field before the read:
```go
func planModifier(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
    if d.Get("application_id").(string) != "" {
        return nil
    }
    // List and find by name, then set application_id and ID
    apps, err := client.ServiceFlinkListApplications(ctx, d.Get("project").(string), d.Get("service_name").(string))
    // ... find matching app, d.Set("application_id", app.Id), d.SetID(...)
}
```

**Reference**: See `internal/plugin/service/flink/application/application.go`.

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
| Alt lookup key | `grep -l "exactlyOneOf:" definitions/aiven_*.yml` |

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
| `flink_application` | Alt data source lookup key (`exactlyOneOf` + `planModifier`), `ConfigValidators`, field rename |
| `flink_application_deployment` | Renamed ID field, planModifier for backward compat, custom delete with state machine |
| `organization/unit` | planModifier for name -> ID resolution via API, expand/flatten for parent ID |

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
