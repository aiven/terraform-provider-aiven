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
2. **YAML Definitions** (`definitions/*.yml`) - Configuration for resource generation

**Generated outputs:**
- `zz_resource.go` / `zz_datasource.go` - Terraform schema
- `zz_converter.go` - Data conversion (expand/flatten)
- `zz_view.go` - CRUD operation handlers
- `docs/resources/*.md` - Documentation
- `examples/resources/*/resource.tf` - Example configs

## User Interaction Pattern

When the user requests a new resource, gather requirements BEFORE starting:

### First Question (REQUIRED)

**ALWAYS start by asking what type of Terraform component is needed:**

Use the `AskUserQuestion` tool with these options:
- **Both resource and data source** - Full CRUD resource + read-only data source
- **Resource only** - Just the resource (create, read, update, delete)
- **Data source only** - Read-only data source

Example:
```
Question: "What type of Terraform component do you need?"
Options:
1. Both resource and data source (Recommended) - Full CRUD resource + read-only data source
2. Resource only - Manage the resource lifecycle (create, update, delete)
3. Data source only - Read-only access to existing resources
```

### Additional Information to Gather
1. **Resource name** - What API resource/endpoint? (e.g., "MySQL database", "Kafka topic")
2. **Scope** - Project-level, service-level, organization-level?

### Discovery Process
1. **Update OpenAPI spec** first by running `task get-spec` to ensure you have the latest API schema
2. **Search OpenAPI spec** for operation IDs
3. **Show the user** what you found and ask for confirmation:
   - "I found these operations: `ServiceDatabaseCreate`, `ServiceDatabaseList`, etc. Are these correct?"
4. **Find similar resources** in `definitions/` and show them:
   - "I found `mysql_database.yml` which looks similar. Should I use this as a template?"
5. **Clarify specifics** if needed:
   - Composite ID structure
   - Fields to exclude (`remove`)
   - Fields to rename (`rename`)
   - Need for custom modifiers

### When NOT to Ask
- If the user provided specific operation IDs → proceed
- If the user provided a complete YAML definition → proceed
- If it's a modification to existing definition → read the file and proceed

### Confirmation Before Generation
Always summarize what you're about to create:
```
I'll create definitions/kafka_topic.yml with:
- Operations: KafkaTopicCreate, KafkaTopicGet, KafkaTopicUpdate, KafkaTopicDelete
- Client handler: service
- Composite ID: [project, service_name, topic_name]
- Based on: mysql_database.yml pattern

Proceed?
```

## Learning from Existing Resources

**CRITICAL**: Before creating a new resource, ALWAYS search for similar existing resources to learn patterns:

```bash
# Find all YAML definitions
ls definitions/

# Find resources with similar client handlers
grep -l "clientHandler: service" definitions/*.yml

# Find resources with custom modifiers
grep -l "expandModifier: true" definitions/*.yml

# Find composite ID examples
grep -l "idAttributeComposed:" definitions/*.yml
```

**Key files to reference:**
- `definitions/.schema.yml` - Complete YAML schema specification
- `definitions/*.yml` - Real working examples
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

File: `definitions/my_resource.yml`

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
- `rename` - Field name mappings (API field → Terraform field)
- `schema` - Schema customizations (types, validation, behavior)
- `expandModifier` / `flattenModifier` / `planModifier` - Enable custom Go modifiers

### Operations Configuration

Each operation maps a Terraform action to an OpenAPI operation ID:

```yaml
operations:
  - id: OperationIDFromOpenAPI
    type: create|read|update|delete
    resultKey: nested_field        # Extract from response.nested_field
    resultToKey: wrapper_key       # Wrap response as {wrapper_key: ...}
    resultListLookupKeys:          # For list responses
      APIField: terraform_field    # Match items by field
```

**Find examples**: `grep -A 5 "operations:" definitions/*.yml`

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
    enum: [val1, val2]
    minimum: 1
    maximum: 100
```

**Important**: Use `arrayOrdered` (list) for complex objects, not `array` (set). Sets cause performance issues.

**Find examples**: `grep -A 10 "schema:" definitions/*.yml`

## Custom Modifiers

For complex data transformations, enable modifiers in YAML:

```yaml
expandModifier: true      # TF state -> API
flattenModifier: true     # API -> TF state
planModifier: true        # Modify planned changes
```

Then implement functions in a `.go` file in the resource package. Common utilities:
- `util.ExpandArrayToObjects[T]()` - Convert string array to object array
- `util.FlattenObjectsToArray[T]()` - Convert object array to string array
- `util.ComposeModifiers()` - Combine multiple modifiers

**Find examples**: `grep -l "Modifier" internal/plugin/service/*/billing_group.go`

## Common Resource Patterns

Instead of copying examples here, search the codebase for similar patterns:

### Data Source (List)
`grep -l "datasource:" definitions/*.yml | xargs grep -l "resultToKey"`

### Composite ID
`grep -l "idAttributeComposed:" definitions/*.yml | head -3`

### Custom Modifiers
`grep -l "expandModifier: true" definitions/*.yml`

### Field Renaming
`grep -l "rename:" definitions/*.yml`

**Best approach**: Find a similar resource, read its YAML definition, understand the pattern, adapt to your needs.

## Troubleshooting

| Error | Solution |
|-------|----------|
| `operationID X not found` | Search OpenAPI: `jq '.paths \| to_entries[] \| .value \| to_entries[] \| .value.operationId' openapi.json \| grep -i X` |
| `no 'id' field found` | Add `idAttributeComposed: [field1, field2]` to YAML |
| `ID field "X" not found` | Verify field exists in API response, check `rename` mappings |
| Generation fails | Run `task get-spec`, validate YAML syntax, check operation IDs |
| Infinite plan changes | Use `arrayOrdered` not `array`, ensure `idAttributeComposed` fields are stable |

## Key Principles

1. **Learn from existing code** - Always find and read similar resources in `definitions/`
2. **Start minimal** - Basic YAML first, add complexity incrementally
3. **Use arrayOrdered** - For complex/large nested objects, always use `arrayOrdered` (list) not `array` (set) for performance
4. **Mark sensitive fields** - Use `sensitive: true` for passwords, tokens, API keys, credentials
5. **Follow conventions** - Match patterns from existing resources (naming, structure, etc.)
6. **Don't edit `zz_*` files** - These are generated. Custom logic goes in separate `.go` files

## Testing Requirements

**CRITICAL**: Every generated resource MUST have acceptance tests covering:

### For Resources
Create `<package>_test.go` with tests for:
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
            // Step 1: Create
            {
                Config: testAccMyResourceConfig("value1"),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr(resourceName, "field", "value1"),
                ),
            },
            // Step 2: Update
            {
                Config: testAccMyResourceConfig("value2"),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr(resourceName, "field", "value2"),
                ),
            },
            // Step 3: Import
            {
                ResourceName:      resourceName,
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}
```

### For Data Sources
Test that the data source:
1. Reads attributes correctly
2. Filters/matches resources properly (if applicable)

### If Both Resource + Data Source Added
Create separate tests for each, ensuring:
- Resource test covers full CRUD + import
- Data source test reads from a created resource

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

## Important Notes

- New resources go in `internal/plugin/` (Plugin Framework), NOT `internal/sdkprovider/` (legacy)
- Generated files have `zz_` prefix - never edit them manually
- Custom logic goes in separate `.go` files in the same package (e.g., `billing_group.go`)
- Run `task lint` before committing
- File name in `definitions/` determines resource name: `my_resource.yml` → `aiven_my_resource`

## Migrating Existing Resources

**If you need to migrate an existing SDK resource** from `internal/sdkprovider/` to Plugin Framework, use the **tf-resource-migration** skill instead. That skill covers:
- Analyzing existing SDK code
- Preserving state compatibility
- Ensuring behavior parity
- Deprecation strategy
