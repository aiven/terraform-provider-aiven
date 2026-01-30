---
name: tf-resource-migration
description: Migrate existing Terraform resources from SDK provider (sdkprovider) to Plugin Framework. Use when converting legacy resources, ensuring state compatibility, or when the user asks to port/migrate existing resources.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write, Task
---

# Terraform Resource Migration Skill

Migrate existing SDK-based resources to Plugin Framework while maintaining state compatibility and behavior parity.

## Overview

This skill guides migration of resources from:
- **Source**: `internal/sdkprovider/` (terraform-plugin-sdk/v2)
- **Target**: `internal/plugin/` (terraform-plugin-framework) with YAML-generated code

**Key Challenge**: Preserve exact behavior and state compatibility so users don't experience breaking changes.

**Prerequisites**: This skill builds on **tf-resource-generator**. For YAML syntax, generation commands, and testing patterns, see that skill.

## When to Use This Skill

**Use this skill when:**
- Migrating existing `aiven_*` resources from SDK to Plugin Framework
- User says: "migrate", "convert", "port", "move to Plugin Framework"
- Modernizing legacy resources

**Use tf-resource-generator skill when:**
- Creating brand new resources from scratch
- Adding resources that don't exist yet

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
# Search for API client calls
grep -A 5 "client\." internal/sdkprovider/service/resource_name.go
```

Then find corresponding OpenAPI operation IDs. See **tf-resource-generator** for OpenAPI search patterns.

### 3. Map SDK Schema to YAML

#### SDK Type → YAML Type

| SDK Type | YAML Type |
|----------|-----------|
| `schema.TypeString` | `type: string` |
| `schema.TypeInt` | `type: integer` |
| `schema.TypeFloat` | `type: number` |
| `schema.TypeBool` | `type: boolean` |
| `schema.TypeList` | `type: arrayOrdered` |
| `schema.TypeSet` | `type: array` (or `arrayOrdered` for performance) |
| `schema.TypeMap` | `additionalProperties: {type: string}` |

#### SDK Attributes → YAML Attributes

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

**Identify in SDK code:**
- `StateUpgraders` → May need state upgrader in Plugin Framework
- `CustomizeDiff` → Use `planModifier: true`
- Flatten/Expand functions → Use `expandModifier: true` / `flattenModifier: true`

For modifier implementation details, see **tf-resource-generator** skill.

### 6. Create YAML Definition

Create `definitions/resource_name.yml`. For complete YAML syntax reference, see **tf-resource-generator** skill.

Focus on migration-specific concerns:
- Match all SDK schema fields exactly
- Preserve ID structure
- Copy descriptions from SDK resource

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

**Example**: See `internal/plugin/service/mysql/database/database_test.go`

### 10. Parity Testing

**CRITICAL**: Verify behavior matches SDK resource exactly.

**Find SDK tests:**
```bash
ls internal/sdkprovider/service/*resource_name*_test.go
```

**Ensure Plugin Framework tests cover:**
- All CRUD operations from SDK tests
- Edge cases
- Error handling
- Import functionality
- Special field behaviors

## State Compatibility Checklist

Before marking migration complete:

- [ ] Resource ID format is identical
- [ ] All schema fields are present (no removals)
- [ ] Field types match exactly
- [ ] Computed fields work the same way
- [ ] Default values match
- [ ] Required/Optional flags match
- [ ] ForceNew behavior matches
- [ ] Import works with existing IDs
- [ ] Existing state can be used without migration
- [ ] Backward compatibility test added using `acc.BackwardCompatibilitySteps()`
- [ ] All SDK test scenarios pass with Plugin version

## Common Migration Issues

| Issue | Solution |
|-------|----------|
| ID format changed accidentally | Verify `idAttributeComposed` matches SDK's ID builder |
| Set ordering causes diffs | Use `arrayOrdered` instead of `array` |
| Computed field becomes required | Keep as `computed: true` if API provides it |
| Custom validation lost | Implement in custom modifier or use schema validation |
| State upgrade needed | Implement state upgrader in Plugin Framework |
| DiffSuppressFunc behavior | Implement plan modifier for custom diff logic |

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

## Key Principles

1. **State compatibility first** - Users should not need to recreate resources
2. **Preserve exact behavior** - Match SDK resource behavior precisely
3. **Test thoroughly** - All SDK test scenarios must pass with Plugin version
4. **Remove SDK version** - Once verified, delete SDK resource to avoid maintenance burden

## After Migration

Once all tests pass and state compatibility is verified:

1. **Remove SDK resource** - Delete from `internal/sdkprovider/` and remove provider registration
2. **Update documentation** - Ensure docs reflect the Plugin Framework version
3. **Add migration notes if needed** - Document any unavoidable behavioral differences

**Do not maintain both versions** - this creates maintenance burden and user confusion.
