/* Autogenerated file. Do not edit manually. */

package {{.Package}}

import (
    "reflect"

	archtypes "github.com/concrete-eth/archetype/types"

	{{ range .Imports }}
	{{- if .Name }}{{ .Name }} "{{ .Path }}"
    {{- else }}"{{ .Path }}"{{ end }}
	{{ end }}
)

var ActionsABIJson = contract.ContractABI

var ActionsSchemaJson = `{{.Json}}`

var ActionSpecs archtypes.ActionSpecs

func init() {
    types := map[string]reflect.Type{
        {{- range .Schemas }}
        "{{.Name}}": reflect.TypeOf({{ActionStructNameFn .Name}}{}),
        {{- end }}
    }
    var err error
	if ActionSpecs, err = archtypes.NewActionSpecsFromRaw(ActionsABIJson, ActionsSchemaJson, types); err != nil {
		panic(err)
	}
}
