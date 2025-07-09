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
		schema, err := genSchema(def.Beta, isResource, root, def.IDAttribute)
		if err != nil {
			return err
		}

		err = os.MkdirAll(def.Location, os.ModePerm)
		if err != nil {
			return fmt.Errorf("could not create directory %s: %w", def.Location, err)
		}

		pkgName := strings.ReplaceAll(filepath.Base(def.Location), "_", "")
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

			// Resources need composeID and expander
			if hasResource {
				codes = append(codes, genComposeID(avnName, def.IDAttribute.Compose))
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

			file := newFile(pkgName)
			for _, c := range codes {
				file.Add(c).Line()
			}

			filePath := genFilePath(def.Location, convertersFileName)
			err = file.Save(filePath)
			if err != nil {
				return fmt.Errorf("could not save file %s: %w", filePath, err)
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

// newFile creates a new file with the given package name, imports, and copyrights.
func newFile(pkgName string, extraImports ...string) *jen.File {
	file := jen.NewFile(strings.ToLower(pkgName))
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

	// Removes the fields that should not be in the schema
	for k, v := range root.Properties {
		if slices.Contains(scope.Definition.Delete, v.JSONPath()) {
			delete(root.Properties, k)
		}
	}

	// Overwrites the fields with the ones from the config
	if scope.CurrentMeta != nil {
		root.Description = or(scope.CurrentMeta.Description, root.Description)
		root.DeprecationMessage = or(scope.CurrentMeta.DeprecationMessage, root.DeprecationMessage)
	}

	// Sets values from the definition.Schema
	err := setEditAttrs(scope, root)
	if err != nil {
		return nil, err
	}

	// Creates new items from the definition.Schema
	for k, item := range scope.Definition.Schema {
		if item.Type == "" {
			// todo: validate property exists
			// The field can't be created
			continue
		}

		// In case of overriding, removes the original field from the schema
		delete(root.Properties, k)
		err := fromDefinition(root, &item, k)
		if err != nil {
			return nil, err
		}
		err = addProperty(root, &item)
		if err != nil {
			return nil, err
		}
	}

	// Marks ID fields
	for i, v := range pkg.IDAttribute.Compose {
		if _, ok := root.Properties[v]; !ok {
			keys := sortedKeys(root.Properties)
			return nil, fmt.Errorf("ID field %q not found in: %s", v, strings.Join(keys, ", "))
		}
		root.Properties[v].InIDAttribute = true
		root.Properties[v].IDAttributePosition = i
	}

	return root, nil
}

func fromDefinition(parent, item *Item, name string) error {
	item.Name = name
	item.JSONName = or(item.JSONName, name)
	item.Parent = parent

	if !item.Required && !item.Computed {
		item.Optional = true
	}

	for k, v := range item.Properties {
		err := fromDefinition(item, v, k)
		if err != nil {
			return err
		}
		err = addProperty(parent, item)
		if err != nil {
			return err
		}
	}

	if item.Items != nil {
		err := fromDefinition(item, item.Items, name)
		if err != nil {
			return err
		}
	}
	return nil
}

// setEditAttrs marks the fields as computed, optional, required or forceNew
func setEditAttrs(scope *Scope, item *Item) error {
	// When a field doesn't appear in the Create request and not in the path
	in := func(a AppearsIn) bool {
		return item.AppearsIn&a > 0
	}

	item.Required = in(CreatePathParameter) || item.Required && in(CreateRequestBody) // Sometimes field required in GET
	item.Optional = item.Optional || !item.Required && in(RequestBody)
	item.Computed = item.Computed || item.Default != nil || !item.Required && !in(RequestBody)
	item.ForceNew = item.ForceNew || (item.Optional || item.Required) && in(CreateHandler) && !in(UpdateRequestBody)

	// Applies patch
	if p, ok := scope.Definition.Schema[item.Path()]; ok {
		err := mergeItems(item, &p, true)
		if err != nil {
			return err
		}
	}

	for _, v := range item.Properties {
		err := setEditAttrs(scope, v)
		if err != nil {
			return err
		}
	}

	if item.Items != nil {
		err := setEditAttrs(scope, item.Items)
		if err != nil {
			return err
		}
	}

	return nil
}

func fromOperationID(scope *Scope, parent *Item, appearsIn AppearsIn, operationID string) error {
	path, err := getOperationID(scope, operationID)
	if err != nil {
		return err
	}

	parent.Description = path.Summary
	err = fromParameter(scope, parent, path, appearsIn|PathParameter)
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
		} else {
			// Often we have fields marked required in Update and other handlers
			request.Required = nil
		}

		if ap&UpdateHandler > 0 {
			ap |= UpdateRequestBody
		}

		err = fromSchema(scope, parent, request, ap)
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
		if ap&CreateHandler > 0 {
			ap |= CreateResponseBody
		}

		// Response should not have required fields.
		// They mess up the schema.
		response.Required = nil

		err = fromSchema(scope, parent, response, ap)
		if err != nil {
			return err
		}
	}
	return nil
}

// reTFName allows only TF acceptable characters
var reTFName = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

func newItem(scope *Scope, parent *Item, name string, s *OASchema, appearsIn AppearsIn, required bool) *Item {
	item := &Item{
		Parent:      parent,
		Name:        name,
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
		// A field appears in the parent, and must inherit its properties.
		// So a ForceNew field won't be Computed at the same time.
		AppearsIn: parent.AppearsIn | appearsIn,
	}

	if v, ok := scope.Definition.Rename[item.JSONPath()]; ok {
		item.Name = v
	} else {
		// Keeps TF names consistent
		item.Name = reTFName.ReplaceAllString(item.Name, "_")
	}
	return item
}

func fromSchema(scope *Scope, item *Item, itemSchema *OASchema, appearsIn AppearsIn) error {
	if p, ok := itemSchema.Properties[scope.Definition.ObjectKey]; ok {
		// The object is hidden on the root level
		if p.Type == SchemaTypeArray {
			return fromSchema(scope, item, p.Items, appearsIn)
		}
		return fromSchema(scope, item, p, appearsIn)
	}

	for _, childName := range sortedKeys(itemSchema.Properties) {
		childSchema := itemSchema.Properties[childName]

		// Some fields are marked as required, because they are required for the Response.
		// Additionally checks if the field is required for the Request.
		// If there is only one field, it is always required
		required := appearsIn&RequestBody > 0 && slices.Contains(itemSchema.Required, childName) || len(itemSchema.Properties) == 1
		childItem := newItem(scope, item, childName, childSchema, appearsIn, required)

		// With the patching workaround we have,
		// it is important to not rely on the schema,
		// but on the type explicitly.
		var elementSchema *OASchema
		switch childSchema.Type {
		case SchemaTypeArray:
			if childSchema.Items != nil {
				elementSchema = childSchema.Items
			}
		case SchemaTypeObject:
			// If additional properties
			if childSchema.AdditionalProperties != nil {
				elementSchema = childSchema.AdditionalProperties
				childSchema.AdditionalProperties = nil
			} else if v, ok := childSchema.Properties[anyField]; ok {
				// ANY is a special case is additional properties
				delete(childSchema.Properties, anyField)
				elementSchema = v
			} else if len(childSchema.Properties) > 0 {
				err := fromSchema(scope, childItem, childSchema, appearsIn)
				if err != nil {
					return err
				}
			} else {
				// A special case for maps with invalid schema
				elementSchema = &OASchema{Type: SchemaTypeString}
			}
		}

		// Either array item or additional properties
		if elementSchema != nil {
			childItem.Items = newItem(scope, childItem, childName, elementSchema, appearsIn, false)
			err := fromSchema(scope, childItem.Items, elementSchema, appearsIn)
			if err != nil {
				return err
			}
		}

		err := addProperty(item, childItem)
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
		prop := newItem(
			scope,
			parent,
			p.Name,
			p.Schema,
			appearsIn,
			p.Required,
		)

		err = addProperty(parent, prop)
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

func addProperty(parent, prop *Item) error {
	// CRUD fields often intersect with each other
	// Merges duplicated fields
	prop.Parent = parent
	other, ok := parent.Properties[prop.Name]
	if ok {
		err := mergeItems(prop, other, false)
		if err != nil {
			return err
		}
	}

	// Definition items have nil properties
	if parent.Properties == nil {
		parent.Properties = make(map[string]*Item)
	}
	parent.Properties[prop.Name] = prop
	return nil
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
