---
name: tf-byoc-resource
description: Add or change BYOC (Bring Your Own Cloud) custom cloud environment resources, e.g. aiven_byoc_<cloud>_entity and aiven_byoc_<cloud>_provision for a new cloud. Use when working on aiven_byoc_* resources, CustomCloudEnvironment* operations, or questions about the BYOC environment lifecycle.
allowed-tools: Read, Grep, Glob, Bash, Edit, Write, Agent
---

# BYOC Resource Skill

How the custom cloud environment API behaves, and how the `aiven_byoc_*` resources map onto it. Read
it before designing a new BYOC resource instead of guessing the flow. The YAML, generator and adapter
mechanics are in [tf-resource-generator](../tf-resource-generator/SKILL.md); this skill covers only
what is specific to BYOC.

**Reference**: `definitions/aiven_byoc_aws_*.yml`, `definitions/aiven_byoc_azure_*.yml`,
`internal/plugin/byoc/provision.go`.

## Where the facts come from

The lifecycle below was established by black-box testing the public API and from the public docs:

- [BYOC concepts](https://aiven.io/docs/platform/concepts/byoc), the
  [AWS](https://aiven.io/docs/platform/howto/byoc/create-cloud/create-aws-custom-cloud) and
  [Azure](https://aiven.io/docs/platform/howto/byoc/create-cloud/create-azure-custom-cloud) setup
  guides, [cloud status](https://aiven.io/docs/platform/howto/byoc/view-custom-cloud-status),
  [deleting a cloud](https://aiven.io/docs/platform/howto/byoc/delete-custom-cloud) and the
  [`avn byoc` CLI](https://aiven.io/docs/tools/cli/byoc).
- The OpenAPI spec (`CustomCloudEnvironment*` operations). The endpoints are marked experimental,
  so the resources are `beta: true`.

Backend sources are private. Never quote, paraphrase, or name their code, files, functions, tables
or internal comments in code, comments, docs, commit messages or PRs. Describe findings as API
behavior ("black-box tested: Provision returns 409 unless ..."). When a fact isn't covered here,
test the API or ask the user; don't infer it.

Prefer letting the API decide over copying its state machine: don't encode "which state allows
which call" in the provider. Retry and poll, and let the API's answer drive the outcome.

## Environment lifecycle

```
Create ──> draft ──Provision──> validating ──> creating ──> active
                                                   └──────> creation_failed
active ──> disconnected ──Provision──> reconnecting ──> active
Delete ──> deleting ──> (gone: GET returns 404)
               └──────> deletion_failed
```

| Call | Black-box tested behavior |
|---|---|
| Create | Makes a `draft` environment. The endpoint is polymorphic: `cloud_provider` picks the cloud. Nothing exists in the customer account yet. |
| Provision | Accepted in `draft` and `creation_failed` (a retry); `disconnected` starts reconnecting. Any other state returns 409 `Unexpected custom cloud environment state '...'`. It stores the identity fields it gets, even when it answers 409. |
| Update | Stores the identity fields (`aws_iam_role_arn`, `azure_subscription_id`, `azure_tenant_id`) in any state, with no validation. On an `active` environment that silently points Aiven at another account. |
| Delete | Accepted in `draft`, `active`, `creation_failed` and `deletion_failed`; rejected while provisioning, reconnecting or deleting. Needs the customer-side access still in place, because it removes the bastion. |
| Get | A deleted environment returns 404; it is never returned with state `deleted`. |

From the public docs:

- `creation_failed` is "an unrecoverable error": the guidance is to delete the cloud and re-create
  it. The API accepts Provision again, but don't promise users that a retry succeeds.
- Provisioning "might take a few minutes", and can run past the create timeout.
- A cloud can't be deleted while it hosts services. The customer-side Terraform resources aren't
  removed by the delete: destroy them after the environment.
- The CLI can't change the role ARN, subscription or tenant after creation.

## Resource split

Each cloud gets two resources, because the customer-side infrastructure has to be built between
the two steps from the entity's outputs:

| Resource | Operations | Role |
|---|---|---|
| `aiven_byoc_<cloud>_entity` | Create, Get, Update, Delete | Creates the `draft` environment and exposes the values the customer-side infrastructure needs. |
| `aiven_byoc_<cloud>_provision` | Provision (as create), Get, Update, Delete | Hands Aiven the customer account, waits for `active`, and deletes the environment on destroy. |

`aiven_byoc_permissions` is shared by all clouds and is created after provision.

Terraform order on apply: entity, then customer-side infrastructure, then provision. On destroy
it's the reverse, so provision deletes the environment while the customer-side access still
exists. The provision resource documents `depends_on` for the infrastructure it doesn't reference.

## Entity definition

- `fixedValues: {cloud_provider: <cloud>}` on the create operation, and `cloud_provider` in
  `remove`: the user can't pick another cloud.
- `removeMissing: true`: destroying provision deletes the environment, so the entity's next
  refresh must drop it instead of failing.
- Remove what changes during provisioning (`state`, `errors`, `update_time`) and what provision
  owns (`custom_cloud_names`, the identity fields). A changing entity attribute would cascade
  updates into every resource that references it.
- The generator prunes other clouds' fields by name (`pruneCrossGroupFields`). Names without a
  cloud token slip through and need an explicit `remove`, e.g. the Google-only
  `aiven_object_storage_credentials_creator_user`.
- `deployment_model`: override the enum with the models the cloud supports. The spec lists every
  cloud's models (Azure supports only `standard` and `standard_public`).
- `forceNew` on what can't change after creation: `cloud_region`, `deployment_model`,
  `reserved_cidr`.

## Provision definition

```yaml
resource:
  alreadyExists: adopt # Provision returns 409 while provisioning or after it, the next apply waits for the outcome
  modifyPlan: true
  stateAttribute: state
  refreshState:
    desired: [active]
    failed: [creation_failed]
  deleteState: true
  removeMissing: true
```

Why each key:

- `alreadyExists: adopt`: Provision is idempotent from Terraform's view. Re-running it after a
  timeout gets a 409 and resumes waiting. Adopt also saves no state when the refresh fails: a tainted
  provision would be replaced on the next apply, and that replacement deletes the environment.
- `refreshState`: wait for `active`, fail fast on `creation_failed`.
- `deleteState: true`: Delete is retried while the API rejects it (mid-provisioning), then Read is
  polled until 404. Don't wait for a `deleted` state: Get never returns it.
- `removeMissing: true`: an environment deleted outside Terraform plans a create instead of failing.
- `modifyPlan: true` with `byoc.RejectIdentityChange(<identity fields>...)`: Update would store
  changed identity fields on an `active` environment, and replacing provision deletes the
  environment. So a change after provisioning fails the plan and tells the user to replace the
  entity. It reads the environment ID from the config: while the entity is replaced, `d.Get`
  returns the old ID from the state. Known gap: an identity value unknown at plan time (e.g. the
  IAM role is replaced in the same apply) isn't checked, and apply stores it. Rejecting unknown
  values would block the common case where the new value is the same. The descriptions say so.
- `forceNew` on `organization_id` and `custom_cloud_environment_id`; `required` on the identity
  fields.
- Remove every entity input and environment attribute exposed by the entity, so each value has one
  owner.

The description must say: the apply order and `depends_on`; that a timed-out create resumes on the
next apply; that `creation_failed` and identity changes need `-replace` on the entity; and that
destroying the resource deletes the environment, after provisioning finishes if it's running.

## Adding a cloud

1. Copy the AWS or Azure definitions and swap the cloud-specific fields: identity fields, outputs,
   `deployment_model` enum, `fixedValues`, `remove`.
2. Check the public setup guide for that cloud: what the customer creates, what Aiven needs back,
   and which deployment models and limits apply.
3. `task generate`, then read the generated `zz_resource.go`: no field from another cloud may be
   left, and every field the user can't set must be computed or removed.
4. Add `var modifyPlan = byoc.RejectIdentityChange(...)` in the provision package's `provision.go`.
5. Add the example (`examples/resources/.../resource.tf`, `import.sh`), acceptance tests next to the
   AWS and Azure ones, and the changelog entries. Update `aiven_byoc_permissions`' description.
6. `task build`, `task test-unit`, `task lint`.

Acceptance tests need a real customer account and are skipped by `acc.SkipBYOC`. Say in the PR
which flows were tested by hand: create, a timed-out create applied again, identity change, entity
replacement and destroy.
