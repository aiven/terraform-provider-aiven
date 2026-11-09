package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadHandlerPackages(t *testing.T) {
	packages, err := readHandlerPackages()
	require.NoError(t, err)
	assert.Equal(t, avnGenHandlerPackage+"flinkapplication", packages["ServiceFlinkCreateApplication"])
	assert.Equal(t, avnGenHandlerPackage+"organizationbilling", packages["OrganizationBillingGroupList"])
}

func TestSetHandlerPackages(t *testing.T) {
	packages := map[OperationID]string{"FooGet": "handler/foo"}
	def := &Definition{fileName: "aiven_foo.yml", Operations: Operations{
		{ID: "FooGet", Type: OperationRead},
		{Fn: "check", Type: OperationUpdate},
	}}
	require.NoError(t, setHandlerPackages([]*Definition{def}, packages))
	assert.Equal(t, "handler/foo", def.Operations[0].Package)
	assert.Empty(t, def.Operations[1].Package)

	def.Operations = append(def.Operations, &Operation{ID: "BarGet", Type: OperationRead})
	require.EqualError(t, setHandlerPackages([]*Definition{def}, packages), "aiven_foo.yml: operation BarGet not found in go-client-codegen")
}
