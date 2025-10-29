package main

import (
	"path/filepath"

	"github.com/dave/jennifer/jen"
)

func genEntityProvider(file *jen.File, entity entityType, definitions []*Definition) {
	m := "resources"
	if entity == datasourceType {
		m = "dataSources"
	}

	title := entity.Title()
	values := make(jen.Dict)
	valuesBeta := make(jen.Dict)
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
		if def.Beta {
			valuesBeta[jen.Lit(def.typeName)] = c
		} else {
			values[jen.Lit(def.typeName)] = c
		}
	}

	entityPkg := entity.Import(entityPackage)
	returnType := jen.Map(jen.String()).Func().Params().Qual(entityPkg, title)

	block := []jen.Code{
		jen.Id(m).Op(":=").Add(returnType.Clone()).Values(values),
	}

	if len(valuesBeta) > 0 {
		block = append(block,
			jen.If(jen.Qual(utilPackage, "IsBeta").Call()).BlockFunc(func(group *jen.Group) {
				for k, v := range valuesBeta {
					group.Id(m).Index(k).Op("=").Add(v)
				}
			}),
		)
	}

	block = append(block, jen.Return(jen.Id(m)))
	file.
		Func().Id(title + "s").Params().
		Add(returnType.Clone()).
		Block(block...).Line()
}
