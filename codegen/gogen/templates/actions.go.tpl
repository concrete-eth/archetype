/* Autogenerated file. Do not edit manually. */

package {{.Package}}

import (
    "reflect"

    archtypes "github.com/concrete-eth/archetype/types"
    "github.com/ethereum/go-ethereum/common"

	{{ range .Imports }}
	"{{.}}"
	{{ end }}
)

var (
	_ = common.Big1
)

{{ if .Schemas }}
const (
    {{- range $index, $element := .Schemas }}
    ActionId_{{.Name}}{{ if eq $index 0 }} uint8 = iota{{ end }}
    {{- end }}
)
{{ end }}

var Actions = map[uint8]archtypes.ActionMetadata{
    {{- range .Schemas }}
    ActionId_{{.Name}}: {
        Id: ActionId_{{.Name}},
        Name: "{{.Name}}",
        MethodName: "{{_lowerFirstChar .Name}}",
        Type: reflect.TypeOf(ActionData_{{.Name}}{}),
    },
    {{- end }}
}

/*
var ActionIdsByMethodName = map[string]uint8{
    {{- range .Schemas }}
    "{{_lowerFirstChar .Name}}": ActionId_{{.Name}},
    {{- end }}
}
*/

{{ range $schema := .Schemas }}
type ActionData_{{.Name}} struct{
    {{- range .Values }}
    {{.PascalCase}} {{.Type.GoType}} `json:"{{.Name}}"`
    {{- end }}
}
{{ range .Values }}
func (action *ActionData_{{$schema.Name}}) Get{{.PascalCase}}() {{.Type.GoType}} {
    return action.{{.PascalCase}}
}
{{ end }}
{{ end }}

/*
func ActionIdFromAction(action interface{}) (uint8, bool) {
	switch action.(type) {
    {{- range .Schemas }}
    case *ActionData_{{.Name}}:
        return ActionId_{{.Name}}, true
    {{- end }}
    default:
        return 0, false
    }
}
*/
