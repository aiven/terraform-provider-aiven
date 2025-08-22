# Open Policy Agent policies for Aiven Terraform Provider

This directory contains Open Policy Agent (OPA) policies designed to help prevent common configuration issues and conflicts when using the Aiven Terraform Provider.

## Table of contents
- [What Are These Policies?](#what-are-these-policies)
- [Available Policies](#available-policies)
- [Local Evaluation of Terraform Plans](#evaluate-your-terraform-plans-locally)
- [CI/CD Integration with GitHub Actions](#enforce-policies-in-your-cicd-pipeline-with-github-actions)

## What are these policies?

OPA is an open-source policy engine that allows you to define and enforce policies as code.
[OPA](https://www.openpolicyagent.org/) analyzes your Terraform plans **before** they are applied, catching potential issues early in your workflow.

> **Note**: These policies are version-specific to the Aiven Terraform provider. Always use policies matching your provider version for accurate validation.

## Available Policies

### Conflicting Resource Policies (`policies/conflicting/`)

**1. Organization Permission Duplicate Prevention**
- **When it triggers**: When multiple `aiven_organization_permission` resources target the same user/group
- **Why it matters**: Prevents Terraform conflicts and API errors
- **Example scenario**: Granting the same user permissions twice in different configurations

**2. Autoscaler Integration and Service Modification Conflict Prevention**
- **When it triggers**: When removing autoscaler integrations while simultaneously modifying the associated service
- **Why it matters**: Helps avoid "Provider produced inconsistent final plan" errors
- **Example scenario**: Deleting an autoscaler integration while updating service configuration in the same plan

**3. ClickHouse Grant Duplicate Prevention**
- **When it triggers**: When multiple `aiven_clickhouse_grant` resources target the same role/user within a service
- **Why it matters**: Prevents grant conflicts
- **Example scenario**: Granting database access to the same user through multiple grant resources

## Evaluate your Terraform plans locally

### Prerequisites
- [Conftest](https://conftest.dev/) installed. Conftest evaluates a JSON representation of your Terraform plan against the OPA policies.

### Steps

1. Define the Aiven Terraform Provider version you are using by running:
```bash
export AIVEN_PROVIDER_VERSION="4.44.0"
```

2. Download the policy bundle:
```bash
curl -L -o aiven-tf-policies.tar.gz "https://github.com/aiven/terraform-provider-aiven/releases/download/v${AIVEN_PROVIDER_VERSION}/aiven-terraform-provider-policies-${AIVEN_PROVIDER_VERSION}.tar.gz"
```

3. Extract the policies into a directory named 'policies':
```bash
mkdir -p policies
tar -xvzf aiven-tf-policies.tar.gz -C policies
```

4. Generate a Terraform plan in JSON format by running:
```bash
terraform plan -out=tfplan.out
terraform show -json tfplan.out > tfplan.json
```

5. To validate your Terraform plan using Conftest, run:
```bash
conftest test --policy policies --namespace aiven.main tfplan.json
```

### Expected output

When policies pass, the output is similar to the following:
```bash
PASS - 0 tests, 0 failures
```

When policies fail, the output is similar to the following:
```bash
FAIL - tfplan.json - aiven.main - Organization permission conflict detected for user 'john@example.com'
FAIL - tfplan.json - aiven.main - ClickHouse grant duplicate detected for role 'analytics_role'
```

## Enforce policies in your CI/CD pipeline with GitHub Actions
Integrating these policy checks into your CI/CD pipeline is the best way to automate enforcement. The following is an example of a GitHub Actions workflow.

```yaml
name: 'Terraform OPA Policy Check'

on:
  pull_request:
    paths:
      - '**.tf'
      - '**.tfvars'

jobs:
  validate:
    name: 'Validate Terraform Plan with OPA'
    runs-on: ubuntu-latest

    # Specify the version of the Aiven provider you are using.
    env:
      AIVEN_PROVIDER_VERSION: '4.44.0'

    steps:
      - name: 'Checkout Code'
        uses: actions/checkout@v4

      - name: 'Set up Terraform'
        uses: hashicorp/setup-terraform@v3

      - name: 'Set up Conftest'
        run: |
          CONFTEST_VERSION="0.62.0"
          curl -sL -o conftest.tar.gz "https://github.com/open-policy-agent/conftest/releases/download/v${CONFTEST_VERSION}/conftest_${CONFTEST_VERSION}_linux_amd64.tar.gz"
          tar -xzf conftest.tar.gz
          sudo mv conftest /usr/local/bin/

      - name: 'Download Aiven OPA Policies'
        run: |
          curl -sL -o aiven-tf-policies.tar.gz "https://github.com/aiven/terraform-provider-aiven/releases/download/v${AIVEN_PROVIDER_VERSION}/aiven-terraform-provider-policies-${AIVEN_PROVIDER_VERSION}.tar.gz"
          mkdir -p opa/policies
          tar -xzf aiven-tf-policies.tar.gz -C opa/policies

      - name: 'Terraform Init'
        id: init
        run: terraform init

      - name: 'Terraform Plan'
        id: plan
        run: terraform plan -no-color -out=tfplan.out

      - name: 'Convert Plan to JSON'
        id: json
        run: terraform show -json tfplan.out > tfplan.json

      - name: 'Run Conftest Policy Check'
        id: conftest
        run: conftest test --policy opa/policies --namespace aiven.main tfplan.json
```
