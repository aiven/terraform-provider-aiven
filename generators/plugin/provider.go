package main

import (
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"
)

func genEntityProvider(file *jen.File, entity entityType, definitions []*Definition) {
	title := entity.Title()
	values := make(jen.Dict)
	for _, def := range definitions {
		if entity.IsDataSource() && def.Datasource == nil || entity == resourceType && def.Resource == nil {
			continue
		}

		if def.ClientHandler == "" {
			continue
		}

		p := filepath.Join(projectPackagePrefix, def.Location)
		// Use a stable, unique alias derived from the resource's Terraform type
		// name (e.g. "aiven_opensearch_user_list" -> "opensearchuserlist"). This
		// keeps zz_provider.go readable: without it, packages that share a base
		// name (say every ".../userlist") collide and goimports picks numeric
		// suffixes like userlist1/userlist2/userlist3 that carry no meaning.
		file.ImportAlias(p, providerImportAlias(def.typeName))

		c := jen.Qual(adapterPackage, "NewLazy"+title).Call(jen.Qual(p, title+optionsSuffix))
		values[jen.Lit(def.typeName)] = c
	}

	entityPkg := entity.Import(entityPackage)
	returnType := jen.Map(jen.String()).Func().Params().Qual(entityPkg, title)
	file.
		Func().Id(title + "s").Params().
		Add(returnType.Clone()).
		Block(jen.Return(returnType.Clone().Values(values))).Line()
}

// providerImportAlias derives the import alias used in zz_provider.go from a
// Terraform resource/datasource type name by trimming the "aiven_" prefix and
// stripping the remaining underscores, matching the way the docs refer to
// each resource.
func providerImportAlias(typeName string) string {
	return strings.ReplaceAll(strings.TrimPrefix(typeName, typeNamePrefix), "_", "")
}
