# Aiven Terraform Provider OPA Policies

This directory contains Open Policy Agent (OPA) policies designed to help prevent common configuration issues and conflicts when using the Aiven Terraform provider.

## What Are These Policies?

These policies use [OPA](https://www.openpolicyagent.org/) and [Conftest](https://conftest.dev/) to analyze your Terraform plans **before** they are applied, catching potential issues early in your workflow.

## Available Policies

### Conflicting Resource Policies (`policies/conflicting/`)

**1. Organization Permission Duplicate Prevention**
- Prevents creating duplicate `aiven_organization_permission` resources for the same entity
- Catches both configuration-time and plan-time duplicates

**2. Autoscaler Integration and Service Modification Conflict Prevention** 
- Prevents removing autoscaler integrations while simultaneously modifying the associated service
- Helps avoid "Provider produced inconsistent final plan" errors