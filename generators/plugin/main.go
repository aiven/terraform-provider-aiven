package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/dave/jennifer/jen"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const (
	openAPIPath      = "openapi.json"
	definitionsPath  = "./definitions"
	openAPIPatchPath = definitionsPath + "/.openapi_patch.yml"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: ""})

	err := exec()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func exec() error {
	// Loads OpenAPIDoc with the patch
	doc, err := readOpenAPIDoc(openAPIPath, openAPIPatchPath)
	if err != nil {
		return err
	}

	// Loads the definitions (yaml files)
	definitions, err := listDefinitionFiles(definitionsPath)
	switch {
	case err != nil:
		return fmt.Errorf("could not list definitions: %w", err)
	case len(definitions) == 0:
		return fmt.Errorf("no definitions found in %s", definitionsPath)
	}

	// Generates files
	for _, fileName := range definitions {
		log.Info().Str("file", fileName).Msg("generating")
		fullPath := filepath.Join(definitionsPath, fileName)
		err := genDefinition(doc, fullPath)
		if err != nil {
			return fmt.Errorf("could not generate package %q: %w", fullPath, err)
		}
	}

	return nil
}

const (
	// aivenNameConst resource and datasource name, e.g. aivenName = "aiven_foo_bar"
	aivenNameConst = "aivenName"

	// aivenNamePrefix is a prefix for the resource and datasource names: aiven_pg, aiven_kafka, etc.
	aivenNamePrefix = "aiven_"

	// convertersFileName contains expandData and flattenData functions
	convertersFileName = "converters"
	viewFileName       = "view"
)

// genDefinition generates zz_datasource.go, zz_resource.go, and zz_converters.go files
func genDefinition(doc *OpenAPIDoc, defPath string) error {
	defFile, err := os.Open(defPath)
	if err != nil {
		return fmt.Errorf("could not open definition file %q: %w", defPath, err)
	}

	// Marshals yaml files with a little validation
	def := new(Definition)
	dec := yaml.NewDecoder(defFile)
	dec.KnownFields(true)
	err = dec.Decode(def)
	if err != nil {
		return err
	}

	resName := strings.TrimSuffix(filepath.Base(defPath), filepath.Ext(defPath))
	if def.Resource == nil && def.Datasource == nil {
		return fmt.Errorf("resource %q has no resource or datasource defined", resName)
	}

	// Stores datasource and resource functions in separate files because they have different imports
	files := map[entityType]*SchemaMeta{
		resourceType:   def.Resource,
		datasourceType: def.Datasource,
	}

	// Some operations must be done once: when either resource or datasource is generated
	doOnce := true
	for _, entity := range sortedKeys(files) {
		resDat := files[entity]
		if resDat == nil {
			// Resource or datasource is not defined
			continue
		}

		scope := &Scope{
			OpenAPI:     doc,
			Definition:  def,
			CurrentMeta: resDat,
		}

		root, err := createRootItem(scope)
		if err != nil {
			return err
		}

		root.Name = resName
		isResource := entity == resourceType
		schema, err := genSchema(isResource, root, def)
		if err != nil {
			return err
		}

		err = os.MkdirAll(def.Location, os.ModePerm)
		if err != nil {
			return fmt.Errorf("could not create directory %s: %w", def.Location, err)
		}

		fileName := filepath.Base(def.Location)
		pkgName := goPkgName(fileName)
		if doOnce {
			doOnce = false
			hasResource := def.Resource != nil
			var codes []jen.Code

			// Renders the resourceName constant
			avnName := aivenNamePrefix + resName
			codes = append(codes, jen.Const().Id(aivenNameConst).Op("=").Lit(avnName))

			// Generates TF and API models
			models, err := genModels(root)
			if err != nil {
				return fmt.Errorf("could not generate model: %w", err)
			}
			codes = append(codes, models...)

			// Resources need idFields and expander
			if hasResource {
				codes = append(codes, genIDFields(avnName, def.IDAttribute.Fields))
				expand, err := genExpand(root)
				if err != nil {
					return fmt.Errorf("could not generate expand: %w", err)
				}
				codes = append(codes, expand...)
			}

			// Flatten is only needed for both resource and datasource
			flatten, err := genFlatten(root)
			if err != nil {
				return fmt.Errorf("could not generate flatten: %w", err)
			}
			codes = append(codes, flatten...)

			convertersFile := newFile(pkgName)
			for _, c := range codes {
				convertersFile.Add(c).Line()
			}

			convertersPath := genFilePath(def.Location, convertersFileName)
			err = convertersFile.Save(convertersPath)
			if err != nil {
				return fmt.Errorf("could not save file %s: %w", convertersPath, err)
			}

			// Generates views
			if def.ClientHandler != "" {
				codes, err := genViews(root, def)
				if err != nil {
					return fmt.Errorf("could not generate views: %w", err)
				}

				viewFile := newFile(pkgName, avnGenHandlerPackage+def.ClientHandler)
				for _, v := range codes {
					viewFile.Add(v).Line()
				}

				viewFilePath := genFilePath(def.Location, viewFileName)
				err = viewFile.Save(viewFilePath)
				if err != nil {
					return fmt.Errorf("could not save file %s: %w", viewFilePath, err)
				}
			}
		}

		filePath := genFilePath(def.Location, entity)
		file := newFile(
			pkgName,
			entityImport(isResource, schemaPackageFmt),
			entityImport(isResource, planmodifierPackageFmt),
			entityImport(isResource, timeoutsPackageFmt),
		)
		file.Add(genTFModel(isResource, root)...)
		file.Add(schema)
		err = file.Save(filePath)
		if err != nil {
			return fmt.Errorf("could not save file %s: %w", filePath, err)
		}
	}
	return nil
}

// generatedFileFormat file prefix similar to Kubebuilder for Kubernetes.
// This pushes them down in file listings and makes them easier to identify visually.
const generatedFileFormat = "zz_%s.go"

// genFilePath generates prefixed file name
func genFilePath[T ~string](location string, name T) string {
	return filepath.Join(location, fmt.Sprintf(generatedFileFormat, name))
}

var reNonAlphaNum = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func goPkgName(s string) string {
	return reNonAlphaNum.ReplaceAllString(strings.ToLower(s), "")
}

// newFile creates a new file with the given package name, imports, and copyrights.
func newFile(pkgName string, extraImports ...string) *jen.File {
	file := jen.NewFile(pkgName)
	file.HeaderComment(fmt.Sprintf("Copyright (c) %d Aiven, Helsinki, Finland. https://aiven.io/", time.Now().Year()))
	file.HeaderComment("Code generated by user config generator. DO NOT EDIT.")

	// Generic imports
	imports := getUntypedImports()
	imports = append(imports, extraImports...)

	// Go-imports will remove unused imports
	for _, imp := range getTypedImports() {
		for k := range typingMapping() {
			imports = append(imports, getTypedImport(k, imp))
		}
	}

	// This is unnecessary, but it removes aliases in the imports
	for _, p := range imports {
		file.ImportName(p, filepath.Base(p))
	}

	// avngen is alias we use for go-client-codegen
	file.ImportAlias(avnGenPackage, "avngen")
	return file
}

func createRootItem(scope *Scope) (*Item, error) {
	pkg := scope.Definition
	root := &Item{
		Properties: make(map[string]*Item),
		Required:   true,
		Type:       SchemaTypeObject,
	}

	operationAppearsIn := map[Operation]AppearsIn{
		OperationCreate: CreateHandler,
		OperationRead:   ReadHandler,
		OperationUpdate: UpdateHandler,
		OperationDelete: DeleteHandler,
		OperationUpsert: CreateHandler | UpdateHandler,
	}

	// Initializes the map
	for _, operationID := range sortedKeys(pkg.Operations) {
		operation := pkg.Operations[operationID]
		appearsIn := operationAppearsIn[operation]
		if appearsIn == 0 {
			return nil, fmt.Errorf("unknown operation %q: %q", operationID, operation)
		}

		err := fromOperationID(scope, root, appearsIn, operationID)
		if err != nil {
			return nil, fmt.Errorf("operationID %s: %w", operationID, err)
		}
	}

	// Overwrites the fields with the ones from the config
	if scope.CurrentMeta != nil {
		root.Description = or(scope.CurrentMeta.Description, root.Description)
		root.DeprecationMessage = or(scope.CurrentMeta.DeprecationMessage, root.DeprecationMessage)
	}

	// Patches the schema and creates new properties
	var err error
	for k, v := range scope.Definition.Schema {
		err = fillOptionalFields(root, v, k)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", k, err)
		}
		root.Properties[k], err = mergeItem(root, root.Properties[k], v)
		if err != nil {
			return nil, fmt.Errorf("can't create item %q: %w", k, err)
		}
	}

	err = recalcDeep(root)
	if err != nil {
		return nil, err
	}

	// Marks ID fields
	for i, v := range pkg.IDAttribute.Fields {
		if _, ok := root.Properties[v]; !ok {
			keys := sortedKeys(root.Properties)
			return nil, fmt.Errorf("ID field %q not found in: %s", v, strings.Join(keys, ", "))
		}
		root.Properties[v].InIDAttribute = true
		root.Properties[v].IDAttributePosition = i
	}

	return root, nil
}

func fromOperationID(scope *Scope, root *Item, appearsIn AppearsIn, operationID string) error {
	path, err := getOperationID(scope, operationID)
	if err != nil {
		return err
	}

	root.Description = path.Summary
	err = fromParameter(scope, root, path, appearsIn|PathParameter)
	if err != nil {
		return err
	}

	if path.Request.Content.ApplicationJSON.Schema.Ref != "" {
		request, err := scope.OpenAPI.getComponentRef(path.Request.Content.ApplicationJSON.Schema.Ref)
		if err != nil {
			return fmt.Errorf("request schema: %w", err)
		}

		ap := appearsIn | RequestBody
		if ap&CreateHandler > 0 {
			ap |= CreateRequestBody
		}

		if ap&UpdateHandler > 0 {
			ap |= UpdateRequestBody
		}

		err = fromSchema(scope, root, request, ap)
		if err != nil {
			return err
		}
	}

	if path.Response.OK.Content.ApplicationJSON.Schema.Ref != "" {
		response, err := scope.OpenAPI.getComponentRef(path.Response.OK.Content.ApplicationJSON.Schema.Ref)
		if err != nil {
			return fmt.Errorf("response schema: %w", err)
		}

		// These fields are reserved for errors
		delete(response.Properties, "message")
		delete(response.Properties, "errors")

		ap := appearsIn | ResponseBody
		switch {
		case ap&CreateHandler > 0:
			ap |= CreateResponseBody
		case ap&DeleteHandler > 0:
			ap |= DeleteResponseBody
		case ap&ReadHandler > 0:
			ap |= ReadResponseBody
		case ap&UpdateHandler > 0:
			ap |= UpdateResponseBody
		}

		err = fromSchema(scope, root, response, ap)
		if err != nil {
			return err
		}
	}
	return nil
}

// reTFName allows only TF acceptable characters
var reTFName = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

func newItem(scope *Scope, parent *Item, name string, s *OASchema, required bool, appearsIn AppearsIn) *Item {
	item := &Item{
		Parent: parent,
		// Adds "AppearsIn" from the parent to the item.
		// A field appears in the parent, and must inherit its properties.
		// So a ForceNew field won't be Computed at the same time.
		AppearsIn: parent.AppearsIn | appearsIn,

		// Keeps TF names consistent
		Name:        reTFName.ReplaceAllString(name, "_"),
		JSONName:    name,
		Type:        s.Type,
		Description: s.Description,
		Enum:        s.Enum,
		Default:     s.Default,
		MinItems:    s.MinItems,
		MaxItems:    s.MaxItems,
		MinLength:   s.MinLength,
		MaxLength:   s.MaxLength,
		Minimum:     s.Minimum,
		Maximum:     s.Maximum,
		ForceNew:    s.CreateOnly,
		Sensitive:   s.Sensitive,
		Nullable:    s.Nullable,
		Required:    required,
		Properties:  make(map[string]*Item),
	}

	// Renames it to match the definition.
	newName, ok := scope.Definition.Rename[item.JSONPath()]
	if ok {
		item.Name = newName
	}

	return item
}

func fromSchema(scope *Scope, parent *Item, parentSchema *OASchema, appearsIn AppearsIn) error {
	if p, ok := parentSchema.Properties[scope.Definition.ObjectKey]; ok {
		// The object is hidden on the root level
		if p.Type == SchemaTypeArray {
			return fromSchema(scope, parent, p.Items, appearsIn)
		}
		return fromSchema(scope, parent, p, appearsIn)
	}

	for _, thisName := range sortedKeys(parentSchema.Properties) {
		thisSchema := parentSchema.Properties[thisName]

		// Some fields are marked as required, because they are required for the Response.
		// Additionally, checks if the field is required for the Request.
		// If there is only one field, it is always required
		required := appearsIn&RequestBody > 0 && slices.Contains(parentSchema.Required, thisName) || len(parentSchema.Properties) == 1
		thisItem := newItem(scope, parent, thisName, thisSchema, required, appearsIn)

		// With the patching workaround we have,
		// it is important to not rely on the schema,
		// but on the type explicitly.
		var thisItemsSchema *OASchema
		switch thisSchema.Type {
		case SchemaTypeArray, SchemaTypeArrayOrdered:
			if thisSchema.Items != nil {
				thisItemsSchema = thisSchema.Items
			}
		case SchemaTypeObject:
			if thisSchema.AdditionalProperties != nil {
				// If additional properties for maps
				thisItemsSchema = thisSchema.AdditionalProperties
			} else if v, ok := thisSchema.Properties[anyField]; ok {
				// ANY is a special case is additional properties
				thisItemsSchema = v
			} else if len(thisSchema.Properties) > 0 {
				// A regular object with properties
				err := fromSchema(scope, thisItem, thisSchema, appearsIn)
				if err != nil {
					return err
				}
			} else {
				// A special case for maps with invalid schema
				thisItemsSchema = &OASchema{Type: SchemaTypeString}
			}
		}

		// Either array item or additional properties
		if thisItemsSchema != nil {
			// Array or Map items
			items := newItem(scope, thisItem, thisItem.Name, thisItemsSchema, thisItem.Required, appearsIn)
			err := fromSchema(scope, items, thisItemsSchema, appearsIn)
			if err != nil {
				return err
			}
			thisItem.Items = items
		}

		err := addProperty(scope, parent, thisItem)
		if err != nil {
			return err
		}
	}

	return nil
}

func fromParameter(scope *Scope, parent *Item, path *OAPath, appearsIn AppearsIn) error {
	if appearsIn&CreateHandler > 0 {
		appearsIn |= CreatePathParameter
	}

	for _, param := range path.Parameters {
		p, err := scope.OpenAPI.getParameterRef(param.Ref)
		if err != nil {
			return err
		}

		p.Schema.Description = p.Description
		prop := newItem(scope, parent, p.Name, p.Schema, p.Required, appearsIn)
		err = addProperty(scope, parent, prop)
		if err != nil {
			return err
		}
	}
	return nil
}

func getOperationID(scope *Scope, operationID string) (*OAPath, error) {
	for _, pathKey := range sortedKeys(scope.OpenAPI.Paths) {
		path := scope.OpenAPI.Paths[pathKey]
		for _, methodKey := range sortedKeys(path) {
			method := path[methodKey]
			if method.OperationID == operationID {
				return method, nil
			}
		}
	}

	return nil, fmt.Errorf("operationID %s not found", operationID)
}

func addProperty(scope *Scope, parent, prop *Item) error {
	if slices.Contains(scope.Definition.Remove, prop.JSONPath()) {
		return nil
	}

	// Definition items have nil properties
	if parent.Properties == nil {
		parent.Properties = make(map[string]*Item)
	}

	// CRUD fields often intersect with each other
	// Merges duplicated fields
	var err error
	parent.Properties[prop.Name], err = mergeItem(parent, parent.Properties[prop.Name], prop)
	return err
}

// listDefinitionFiles returns go file names
func listDefinitionFiles(dir string) ([]string, error) {
	pattern := regexp.MustCompile(`^[a-z_]+\.yml$`)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	list := make([]string, 0, len(files))
	for _, file := range files {
		name := file.Name()
		if file.IsDir() || !pattern.MatchString(name) {
			continue
		}
		list = append(list, name)
	}
	return list, nil
}

func mergeItem(parent, a, b *Item) (*Item, error) {
	switch {
	case a == nil && b == nil:
		return nil, nil // Both are nil, nothing to merge
	case b == nil:
		a.Parent = parent
		return a, nil
	case a == nil:
		// For creating items from the definition
		a = &Item{
			Name:       b.Name,
			JSONName:   b.JSONName,
			Type:       b.Type,
			Properties: make(map[string]*Item),
		}
	}

	var err error
	for k, v := range b.Properties {
		a.Properties[k], err = mergeItem(a, a.Properties[k], v)
		if err != nil {
			return nil, err
		}
	}

	a.Items, err = mergeItem(a, a.Items, b.Items)
	if err != nil {
		return nil, err
	}

	a.Parent = parent
	a.AppearsIn |= b.AppearsIn
	a.Type = or(b.Type, a.Type) // In case user overrides the type
	a.Required = a.Required || b.Required
	a.Optional = a.Optional || b.Optional
	a.Computed = a.Computed || b.Computed
	a.ForceNew = a.ForceNew || b.ForceNew
	a.Sensitive = a.Sensitive || b.Sensitive
	a.Description = orLonger(a.Description, b.Description)
	a.DeprecationMessage = orLonger(a.DeprecationMessage, b.DeprecationMessage)
	a.Enum = mergeSlices(a.Enum, b.Enum)
	a.Default = or(b.Default, a.Default)
	a.MinItems = max(a.MinItems, b.MinItems)
	a.MaxItems = max(a.MaxItems, b.MaxItems)
	a.MinLength = max(a.MinLength, b.MinLength)
	a.MaxLength = max(a.MaxLength, b.MaxLength)
	a.Minimum = max(a.Minimum, b.Minimum)
	a.Maximum = max(a.Maximum, b.Maximum)

	// Validators
	a.ConflictsWith = mergeSlices(a.ConflictsWith, b.ConflictsWith)
	a.ExactlyOneOf = mergeSlices(a.ExactlyOneOf, b.ExactlyOneOf)
	a.AtLeastOneOf = mergeSlices(a.AtLeastOneOf, b.AtLeastOneOf)
	a.AlsoRequires = mergeSlices(a.AlsoRequires, b.AlsoRequires)

	// User overrides that can set "false" values
	a.OverrideSensitive = or(b.OverrideSensitive, a.OverrideSensitive)
	a.OverrideRequired = or(b.OverrideRequired, a.OverrideRequired)
	a.OverrideComputed = or(b.OverrideComputed, a.OverrideComputed)
	a.OverrideOptional = or(b.OverrideOptional, a.OverrideOptional)
	a.OverrideForceNew = or(b.OverrideForceNew, a.OverrideForceNew)
	a.UseStateForUnknown = a.UseStateForUnknown || b.UseStateForUnknown

	if a.Name == "" {
		return nil, fmt.Errorf("node doesn't have name set, parent: %q", parent.Name)
	}

	if a.Type == "" {
		return nil, fmt.Errorf("node doesn't have type set, parent: %q", parent.Name)
	}

	return a, nil
}

// recalcDeep sets Computed, Required, Optional, ForceNew, and Sensitive fields
// WARNING: never re-inherit Item.AppearsIn - a parent might exist both in Request and Response.
// But the child might appear only in the Response (as computed field).
// It is OK to inherit AppearsIn from the parent when you create a child node.
// But shouldn't do that after all nodes are merged.
func recalcDeep(item *Item) error {
	in := func(a AppearsIn) bool {
		return item.AppearsIn&a > 0
	}

	item.Required = in(CreatePathParameter) || item.Required && in(CreateRequestBody) // Sometimes field required in GET
	item.Optional = item.Optional || !item.Required && in(RequestBody)
	item.Computed = item.Computed || item.Default != nil || !item.Required && !in(RequestBody) && in(ResponseBody)
	item.ForceNew = item.ForceNew || (item.Optional || item.Required) && in(CreateHandler) && !in(UpdateRequestBody)

	// Overrides that come from the user config
	item.Optional = ptrOrDefault(item.OverrideOptional, item.Optional)
	item.Required = ptrOrDefault(item.OverrideRequired, item.Required)

	// Optional field can't be required
	// If user has set both Required and Optional, TF will complain about it.
	if item.Optional && item.OverrideRequired == nil {
		item.Required = false
	}

	if item.Required {
		// Required field can't be optional
		if item.OverrideOptional == nil {
			item.Optional = false
		}

		// Required field can't be computed
		if item.OverrideComputed == nil {
			item.Computed = false
		}
	}

	item.Sensitive = ptrOrDefault(item.OverrideSensitive, item.Sensitive)
	item.Computed = ptrOrDefault(item.OverrideComputed, item.Computed)
	item.ForceNew = ptrOrDefault(item.OverrideForceNew, item.ForceNew)

	for k, v := range item.Properties {
		// If parent is computed, child must be computed too
		v.Computed = v.Computed || item.Computed
		v.UseStateForUnknown = v.UseStateForUnknown || item.UseStateForUnknown
		err := recalcDeep(v)
		if err != nil {
			return fmt.Errorf("%s property: %w", k, err)
		}
	}

	if item.Items != nil {
		// If parent is computed, child must be computed too
		item.Items.Computed = item.Items.Computed || item.Computed
		item.Items.UseStateForUnknown = item.Items.UseStateForUnknown || item.UseStateForUnknown
		err := recalcDeep(item.Items)
		if err != nil {
			return fmt.Errorf("%s items: %w", item.Name, err)
		}
	}

	return nil
}

// fillOptionalFields the definition files have most of the fields optional.
// Fills the gaps.
func fillOptionalFields(parent *Item, item *Item, name string) error {
	item.Parent = parent
	item.Name = or(item.Name, name)
	item.JSONName = or(item.JSONName, name)

	for k, v := range item.Properties {
		err := fillOptionalFields(item, v, k)
		if err != nil {
			return fmt.Errorf("%s: %w", k, err)
		}
	}

	if item.Items != nil {
		err := fillOptionalFields(item, item.Items, name)
		if err != nil {
			return fmt.Errorf("%s items: %w", item.Name, err)
		}
	}
	return nil
}
