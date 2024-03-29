package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aiven/go-api-schemas/pkg/dist"
	"github.com/dave/jennifer/jen"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"golang.org/x/tools/imports"
	"gopkg.in/yaml.v3"
)

const (
	destPath         = "./internal/sdkprovider/userconfig/"
	localPrefix      = "github.com/aiven/terraform-provider-aiven"
	importSchema     = "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	importDiff       = "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/userconfig/diff"
	importValidation = "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	codeGenerated    = "Code generated by user config generator. DO NOT EDIT."
)

func main() {
	var serviceList, integrationList, endpointList string
	flag.StringVar(&serviceList, "excludeServices", "", "Comma separated list of names to exclude from generation")
	flag.StringVar(&integrationList, "excludeIntegrations", "", "Comma separated list of names to exclude from generation")
	flag.StringVar(&endpointList, "excludeEndpoints", "", "Comma separated list of names to exclude from generation")
	flag.Parse()

	err := generate("service", dist.ServiceTypes, strings.Split(serviceList, ","))
	if err != nil {
		log.Fatalf("generating services: %s", err)
	}

	err = generate("serviceintegration", dist.IntegrationTypes, strings.Split(integrationList, ","))
	if err != nil {
		log.Fatalf("generating integrations: %s", err)
	}

	err = generate("serviceintegrationendpoint", dist.IntegrationEndpointTypes, strings.Split(endpointList, ","))
	if err != nil {
		log.Fatalf("generating integration endpoints: %s", err)
	}
}

func generate(kind string, data []byte, exclude []string) error {
	// Fixes imports order
	imports.LocalPrefix = localPrefix

	var root map[string]*object
	err := yaml.Unmarshal(data, &root)
	if err != nil {
		return err
	}

	dirPath := filepath.Join(destPath, kind)
	err = os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return err
	}

	keys := maps.Keys(root)
	slices.Sort(keys)
	doneKeys := make([]string, 0, len(keys))
	doneNames := make([]string, 0, len(keys))

	for _, key := range keys {
		if slices.Contains(exclude, key) {
			log.Printf("skipping %q: in exclude list", key)
			continue
		}

		o, ok := root[key]
		if !ok {
			return fmt.Errorf("key %q not found in spec", key)
		}

		if len(o.Properties) == 0 {
			log.Printf("skipping %q: empty root object", key)
			continue
		}

		o.isRoot = true
		o.init(key + "_user_config")
		if o.Description == "" {
			o.Description = toUpperFirst(o.camelName) + " user configurable settings"
		}

		doneKeys = append(doneKeys, key)
		doneNames = append(doneNames, o.camelName)

		f := jen.NewFile(kind)
		f.HeaderComment(codeGenerated)
		f.ImportName(importSchema, "schema")
		f.ImportName(importDiff, "diff")
		f.ImportName(importValidation, "validation")
		err := genSchema(f, o)
		if err != nil {
			return fmt.Errorf("error generating %q: %w", key, err)
		}

		// Sorts imports
		b, err := imports.Process("", []byte(f.GoString()), nil)
		if err != nil {
			return err
		}

		// Saves file
		err = os.WriteFile(filepath.Join(dirPath, key+".go"), b, 0644)
		if err != nil {
			return err
		}
	}

	cases := make([]jen.Code, 0, len(keys)+1)
	for i, k := range doneKeys {
		cases = append(cases, jen.Case(jen.Lit(k)).Block(
			jen.Return(jen.Id(doneNames[i]).Call()),
		))
	}

	// Panics if unknown kind requested
	cases = append(cases, jen.Default().Block(jen.Return(jen.Nil())))

	f := jen.NewFile(kind)
	f.HeaderComment(codeGenerated)
	f.ImportName(importSchema, "schema")
	f.Func().Id("GetUserConfig").Params(jen.Id("kind").String()).Op("*").Qual(importSchema, "Schema").Block(
		jen.Switch(jen.Id("kind")).Block(cases...),
	)

	configTypes := make([]jen.Code, 0)
	for _, v := range doneKeys {
		configTypes = append(configTypes, jen.Lit(v))
	}
	f.Func().Id("UserConfigTypes").Params().Index().String().Block(
		jen.Return(jen.Index().String().Values(configTypes...)),
	)
	return f.Save(filepath.Join(dirPath, kind+".go"))
}

func genSchema(f *jen.File, o *object) error {
	values, err := getSchemaValues(o)
	if err != nil {
		return err
	}

	f.Func().Id(o.camelName).Params().Op("*").Qual(importSchema, "Schema").Block(
		jen.Return(jen.Op("&").Qual(importSchema, "Schema").Values(values)),
	)
	return nil
}

func getSchemaValues(o *object) (jen.Dict, error) {
	values := make(jen.Dict)

	if d := getDescription(o); d != "" {
		for old, n := range replaceDescriptionSubStrings {
			d = strings.ReplaceAll(d, old, n)
		}
		values[jen.Id("Description")] = jen.Lit(d)
	}

	var t string
	switch o.Type {
	case objectTypeObject, objectTypeArray:
		if o.isSchemaless() {
			// todo: handle schemaless if this happens
			return nil, fmt.Errorf("schemaless is not implemented: %q", o.jsonName)
		}

		t = "List"
		if isTypeSet(o) {
			t = "Set"
		}

		if o.MinItems != nil {
			values[jen.Id("MinItems")] = jen.Lit(*o.MinItems)
		}
		if o.MaxItems != nil {
			values[jen.Id("MaxItems")] = jen.Lit(*o.MaxItems)
		}
	case objectTypeBoolean:
		t = "Bool"
	case objectTypeString:
		t = "String"
	case objectTypeInteger:
		t = "Int"
	case objectTypeNumber:
		t = "Float"
	default:
		return nil, fmt.Errorf("unknown type %q for %q", o.Type, o.jsonName)
	}

	values[jen.Id("Type")] = jen.Qual(importSchema, "Type"+t)
	if o.IsDeprecated {
		if o.DeprecationNotice == "" {
			return nil, fmt.Errorf("missing deprecation notice for %q", o.jsonName)
		}
		values[jen.Id("Deprecated")] = jen.Lit(o.DeprecationNotice)
	}

	if o.CreateOnly {
		values[jen.Id("ForceNew")] = jen.True()
	}

	// Doesn't mark with required or optional scalar elements of arrays
	if !(o.isScalar() && o.parent.isArray()) {
		if o.Required {
			values[jen.Id("Required")] = jen.True()
		} else {
			values[jen.Id("Optional")] = jen.True()
		}
	}

	if o.isScalar() {
		if strings.Contains(o.jsonName, "api_key") || strings.Contains(o.jsonName, "password") {
			values[jen.Id("Sensitive")] = jen.True()
		}

		if o.Enum != nil {
			args := make([]jen.Code, 0)
			for _, v := range o.Enum {
				args = append(args, scalarLit(o, v.Value))
			}

			call, err := scalarArrayLit(o, args)
			if err != nil {
				return nil, err
			}

			// There are no other types functions.
			// Bool and number won't compile
			switch o.Type {
			case objectTypeString:
				values[jen.Id("ValidateFunc")] = jen.Qual(importValidation, "StringInSlice").Call(call, jen.False())
			case objectTypeInteger:
				values[jen.Id("ValidateFunc")] = jen.Qual(importValidation, "IntInSlice").Call(call)
			}
		}

		return values, nil
	}

	if o.isArray() {
		if o.ArrayItems.isScalar() {
			fields, err := getSchemaValues(o.ArrayItems)
			if err != nil {
				return nil, err
			}

			values[jen.Id("Elem")] = jen.Op("&").Qual(importSchema, "Schema").Values(fields)
			return values, nil
		}

		// Renders the array as an object
		o = o.ArrayItems
	}

	if o.isRoot {
		values[jen.Id("DiffSuppressFunc")] = jen.Qual(importDiff, "SuppressUnchanged")
	}

	fields := make(jen.Dict)
	for _, p := range o.properties {
		vals, err := getSchemaValues(p)
		if err != nil {
			return nil, err
		}

		fields[jen.Lit(p.tfName)] = jen.Values(vals)
	}

	values[jen.Id("Elem")] = jen.Op("&").Qual(importSchema, "Resource").Values(jen.Dict{
		jen.Id("Schema"): jen.Map(jen.String()).Op("*").Qual(importSchema, "Schema").Values(fields),
	})

	return values, nil
}

func getDescription(o *object) string {
	desc := make([]string, 0)
	d := o.Description
	if len(d) < len(o.Title) {
		d = o.Title
	}

	// Comes from the schema, quite confusing
	d = strings.TrimSuffix(d, "The default value is `map[]`.")
	if d != "" {
		desc = append(desc, addDot(d))
	}

	if o.isScalar() && o.Default != nil {
		desc = append(desc, fmt.Sprintf("The default value is `%v`.", o.Default))
	}

	// Trims dot from description, so it doesn't look weird with link to nested schema
	// Example: Databases to expose[dot] (see [below for nested schema]...)
	if len(desc) == 1 && o.isNestedBlock() {
		return strings.Trim(desc[0], ".")
	}

	return strings.Join(desc, " ")
}

func addDot(s string) string {
	if s != "" {
		switch s[len(s)-1:] {
		case ".", "!", "?":
		default:
			s += "."
		}
	}
	return s
}

var replaceDescriptionSubStrings = map[string]string{
	"UserConfig":                   "",
	"DEPRECATED: ":                 "",
	"This setting is deprecated. ": "",
	"[seconds]":                    "(seconds)",
}

func scalarLit(o *object, value any) *jen.Statement {
	switch o.Type {
	case objectTypeString:
		return jen.Lit(value.(string))
	case objectTypeBoolean:
		return jen.Lit(value.(bool))
	case objectTypeInteger:
		n, _ := strconv.Atoi(value.(string))
		return jen.Lit(n)
	case objectTypeNumber:
		return jen.Lit(value.(float64))
	}
	log.Fatalf("unknown scalar %v", o)
	return nil
}

func scalarArrayLit(o *object, args []jen.Code) (*jen.Statement, error) {
	switch o.Type {
	case objectTypeString:
		return jen.Index().String().Values(args...), nil
	case objectTypeInteger:
		return jen.Index().Int().Values(args...), nil
	}
	return nil, fmt.Errorf("unexpected element type of array for default value: %q", o.Type)
}

func isTypeSet(o *object) bool {
	// Allowlist for set types
	// Warning: test each type you add!
	switch o.path() {
	case "ip_filter", "ip_filter_string", "ip_filter_object":
		return true
	}
	return false
}
