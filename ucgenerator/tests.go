package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// genJSONSample generates sample JSON for a test
// If not allFields provided, creates a smaller json, which helps to test nil values (missing)
func genJSONSample(b *strings.Builder, o *object, allFields bool) string {
	switch o.Type {
	case objectTypeObject:
		b.WriteString("{")
		for i, p := range o.properties {
			// Either field required or all fields printed
			if !(p.Required || allFields || !p.CreateOnly) {
				continue
			}

			b.WriteString(fmt.Sprintf("%q:", p.jsonName))
			genJSONSample(b, p, allFields)
			if i+1 != len(o.properties) {
				b.WriteString(",")
			}
		}
		b.WriteString("}")
	case objectTypeArray:
		b.WriteString("[")
		genJSONSample(b, o.ArrayItems, allFields)
		b.WriteString("]")
	case objectTypeString:
		b.WriteString(`"foo"`)
	case objectTypeBoolean:
		b.WriteString("true")
	case objectTypeInteger:
		b.WriteString("1")
	case objectTypeNumber:
		b.WriteString("1")
	}
	return b.String()
}

func genTestFile(pkg string, o *object) (string, error) {
	allFields, err := indentJSON(genJSONSample(new(strings.Builder), o, true))
	if err != nil {
		return "", err
	}

	updateOnlyFields, err := indentJSON(genJSONSample(new(strings.Builder), o, false))
	if err != nil {
		return "", err
	}

	file := fmt.Sprintf(
		testFile,
		codeGenerated,
		pkg,
		o.camelName,
		fmt.Sprintf("`%s`", allFields),
		fmt.Sprintf("`%s`", updateOnlyFields),
	)

	return strings.TrimSpace(file), nil
}

func indentJSON(s string) (string, error) {
	s = strings.ReplaceAll(s, ",}", "}") // fixes trailing comma when not all fields are generated
	m := make(map[string]any)
	err := json.Unmarshal([]byte(s), &m)
	if err != nil {
		return "", err
	}

	b, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

const testFile = `
// %[1]s

package %[2]s

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

const allFields = %[4]s
const updateOnlyFields = %[5]s

func Test%[3]s(t *testing.T) {
	cases := []struct{
		name string
		source string
		expect string
		create bool
	}{
		{
			name: "fields to create resource",
			source: allFields,
			expect: allFields,
			create: true,
		},
		{
			name: "only fields to update resource",
			source: allFields,
			expect: updateOnlyFields, // usually, fewer fields
			create: false,
		},
	}

	ctx := context.Background()
	diags := make(diag.Diagnostics, 0)
	for _, opt := range cases {
		t.Run(opt.name, func(t *testing.T) {
			dto := new(dto%[3]s)
			err := json.Unmarshal([]byte(opt.source), dto)
			require.NoError(t, err)

			// From json to TF
			tfo := flatten%[3]s(ctx, diags, dto)
			require.Empty(t, diags)

			// From TF to json
			config := expand%[3]s(ctx, diags, tfo)
			require.Empty(t, diags)
			
			// Run specific marshal (create or update resource)
			dtoConfig, err := schemautil.MarshalUserConfig(config, opt.create)
			require.NoError(t, err)

			// Compares that output is strictly equal to the input
			// If so, the flow is valid
			b, err := json.MarshalIndent(dtoConfig, "", "    ")
			require.NoError(t, err)
			require.Empty(t, cmp.Diff(opt.expect, string(b)))
		})
	}
}
`
