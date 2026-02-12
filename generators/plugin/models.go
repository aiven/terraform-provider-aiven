package main

import (
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"
)

const funcIDFields = "idFields"

func genIDFields(avnName string, root *Item) jen.Code {
	fields := make([]string, 0)
	for _, v := range root.GetIDFields() {
		fields = append(fields, v.Name)
	}

	codes := make([]jen.Code, 0)
	for _, v := range fields {
		codes = append(codes, jen.Lit(v))
	}

	return jen.
		Commentf("%s the ID attribute fields, i.e.:", funcIDFields).Line().
		Commentf("terraform import %s.foo %s", avnName, strings.ToUpper(filepath.Join(fields...))).Line().
		Func().Id(funcIDFields).Params().Index().String().
		Block(jen.Return().Index().String().Values(codes...))
}
