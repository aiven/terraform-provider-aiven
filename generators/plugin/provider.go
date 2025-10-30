package main

import (
	"path/filepath"

	"github.com/dave/jennifer/jen"
)

func genEntityProvider(file *jen.File, entity entityType, definitions []*Definition) {
	title := entity.Title()
	values := make(jen.Dict)
	for _, def := range definitions {
		if entity == datasourceType && def.Datasource == nil || entity == resourceType && def.Resource == nil {
			continue
		}

		if def.ClientHandler == "" {
			continue
		}

		p := filepath.Join(projectPackagePrefix, def.Location)
		file.ImportName(p, goPkgName(p))

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
