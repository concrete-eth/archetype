/* Autogenerated file. Do not edit manually. */

package archmod

import (
	"reflect"

	"github.com/concrete-eth/archetype/arch"

	contract "github.com/concrete-eth/archetype/example/gogen/abigen/tables"
	"github.com/concrete-eth/archetype/example/gogen/datamod"
)

var TablesABIJson = contract.ContractABI

var TablesSchemaJson = `{
    "meta": {
        "schema": {
            "maxBodyCount": "uint8",
            "bodyCount": "uint8"
        }
    },
    "bodies": {
        "keySchema": {
            "bodyId": "uint8"
        },
        "schema": {
            "x": "int32",
            "y": "int32",
            "r": "uint32",
            "vx": "int32",
            "vy": "int32"
        }
    }
}`

var TableSpecs arch.TableSpecs

func init() {
	types := map[string]reflect.Type{
		"Meta":   reflect.TypeOf(RowData_Meta{}),
		"Bodies": reflect.TypeOf(RowData_Bodies{}),
	}
	getters := map[string]interface{}{
		"Meta":   datamod.NewMeta,
		"Bodies": datamod.NewBodies,
	}
	var err error
	if TableSpecs, err = arch.NewTableSpecsFromRaw(TablesABIJson, TablesSchemaJson, types, getters); err != nil {
		panic(err)
	}
}
