package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// Entity represents a resource or datasource.
type Entity struct {
	// Type is either "resource" or "datasource".
	Type string
	// Name is the name of the entity.
	Name string
	// Path is the location of the package that contains the entity.
	Path string
}

// EntityProvider is an interface that provides a list of entities.
type EntityProvider interface {
	// GetEntities returns a list of entities.
	GetEntities() []Entity
}

// SDKv2EntityProvider is an EntityProvider that provides entities from the SDKv2 implementation of the provider.
type SDKv2EntityProvider struct{}

// PluginFrameworkEntityProvider is an EntityProvider that provides entities from the Plugin Framework implementation
// of the provider.
type PluginFrameworkEntityProvider struct{}

func main() {
	fileName := "../../internal/sdkprovider/provider/provider.go"
	entities, err := parseFileForEntities(fileName)
	if err != nil {
		fmt.Println("Error parsing file:", err)
		return
	}

	printEntities(entities)
}

func parseFileForEntities(fileName string) ([]Entity, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, fileName, nil, 0)
	if err != nil {
		return nil, err
	}

	imports := extractImports(node)
	return extractEntities(node, imports), nil
}

func extractImports(node *ast.File) map[string]string {
	imports := make(map[string]string)
	ast.Inspect(node, func(n ast.Node) bool {
		if importSpec, ok := n.(*ast.ImportSpec); ok {
			pkgName, path := parseImportSpec(importSpec)
			imports[pkgName] = path
		}
		return true
	})
	return imports
}

func parseImportSpec(importSpec *ast.ImportSpec) (pkgName, path string) {
	if importSpec.Name != nil {
		pkgName = importSpec.Name.Name
	} else {
		pathTrimmed := strings.Trim(importSpec.Path.Value, `"`)
		pathParts := strings.Split(pathTrimmed, "/")
		pkgName = pathParts[len(pathParts)-1]
	}
	path = strings.Trim(importSpec.Path.Value, `"`)
	return
}

func extractEntities(node *ast.File, imports map[string]string) []Entity {
	var entities []Entity
	ast.Inspect(node, func(n ast.Node) bool {
		if compositeLit, ok := n.(*ast.CompositeLit); ok {
			entities = append(entities, parseCompositeLit(compositeLit, imports)...)
		}
		return true
	})
	return entities
}

func parseCompositeLit(compositeLit *ast.CompositeLit, imports map[string]string) []Entity {
	var entities []Entity
	for _, elt := range compositeLit.Elts {
		if keyValueExpr, ok := elt.(*ast.KeyValueExpr); ok {
			entities = append(entities, parseKeyValueExpr(keyValueExpr, imports)...)
		}
	}
	return entities
}

func parseKeyValueExpr(keyValueExpr *ast.KeyValueExpr, imports map[string]string) []Entity {
	var entities []Entity
	keyIdent, ok := keyValueExpr.Key.(*ast.Ident)
	if !ok || (keyIdent.Name != "ResourcesMap" && keyIdent.Name != "DataSourcesMap") {
		return nil
	}

	compLit, ok := keyValueExpr.Value.(*ast.CompositeLit)
	if !ok {
		return nil
	}

	for _, innerElt := range compLit.Elts {
		if innerKv, ok := innerElt.(*ast.KeyValueExpr); ok {
			if lit, ok := innerKv.Key.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				if callExpr, ok := innerKv.Value.(*ast.CallExpr); ok {
					if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
						if pkgIdent, ok := selExpr.X.(*ast.Ident); ok {
							entityType := determineEntityType(keyIdent.Name)
							entity := Entity{
								Type: entityType,
								Name: strings.Trim(lit.Value, `"`),
								Path: imports[pkgIdent.Name],
							}
							entities = append(entities, entity)
						}
					}
				}
			}
		}
	}
	return entities
}

func determineEntityType(keyIdentName string) string {
	if keyIdentName == "DataSourcesMap" {
		return "datasource"
	}
	return "resource"
}

func printEntities(entities []Entity) {
	for _, entity := range entities {
		fmt.Printf("%+v\n", entity)
	}
}
