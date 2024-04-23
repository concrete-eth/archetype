package params

import (
	"unicode"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/crypto"
)

// ValueParams holds value parameters.
var ValueParams = map[string]interface{}{
	"ActionExecutedEventName": ActionExecutedEventName,
	"MultiActionMethodName":   MultiActionMethodName,
	"IActionsContract":        IActionsContract,
	"ITablesContract":         ITablesContract,
	"ICoreContract":           ICoreContract,
	"EntrypointContract":      EntrypointContract,
	"TickActionName":          TickActionName,
	"TickActionIdHex":         TickActionIdHex,
}

// FunctionParams holds function parameters.
var FunctionParams = map[string]interface{}{
	"GoActionMethodNameFn":       GoActionMethodName,
	"GoActionStructNameFn":       GoActionStructName,
	"GoTableMethodNameFn":        GoTableMethodName,
	"GoTableStructNameFn":        GoTableStructName,
	"SolidityActionMethodNameFn": SolidityActionMethodName,
	"SolidityActionStructNameFn": SolidityActionStructName,
	"SolidityTableMethodNameFn":  SolidityTableMethodName,
	"SolidityTableStructNameFn":  SolidityTableStructName,
}

const (
	ActionExecutedEventName = "ActionExecuted"
	ActionEventSignature    = "ActionExecuted(bytes4,bytes)"
	MultiActionMethodName   = "executeMultipleActions"
)

var (
	ActionExecutedEventID = crypto.Keccak256Hash([]byte(ActionEventSignature))
)

func lowerFirstChar(s string) string {
	if s == "" {
		return ""
	}
	return string(unicode.ToLower(rune(s[0]))) + s[1:]
}

func upperFirstChar(s string) string {
	if s == "" {
		return ""
	}
	return string(unicode.ToUpper(rune(s[0]))) + s[1:]
}

func actionMethodName(name string) string {
	return name
}

func actionStructName(name string) string {
	return "ActionData_" + upperFirstChar(name)
}

func tableMethodName(name string) string {
	return "get" + upperFirstChar(name) + "Row"
}

func tableStructName(name string) string {
	return "RowData_" + upperFirstChar(name)
}

func GoActionMethodName(name string) string {
	return upperFirstChar(actionMethodName(name))
}

func GoActionStructName(name string) string {
	return actionStructName(name)
}

func GoTableMethodName(name string) string {
	return upperFirstChar(tableMethodName(name))
}

func GoTableStructName(name string) string {
	return tableStructName(name)
}

func SolidityActionMethodName(name string) string {
	return lowerFirstChar(actionMethodName(name))
}

func SolidityActionStructName(name string) string {
	return actionStructName(name)
}

func SolidityTableMethodName(name string) string {
	return lowerFirstChar(tableMethodName(name))
}

func SolidityTableStructName(name string) string {
	return tableStructName(name)
}

type ContractSpecs struct {
	FileName     string
	ContractName string
	PackageName  string
}

var IActionsContract = ContractSpecs{
	FileName:     "IActions.sol",
	ContractName: "IActions",
	PackageName:  "actions",
}

var ITablesContract = ContractSpecs{
	FileName:     "ITables.sol",
	ContractName: "ITables",
	PackageName:  "tables",
}

var ICoreContract = ContractSpecs{
	FileName:     "ICore.sol",
	ContractName: "ICore",
	PackageName:  "core",
}

var EntrypointContract = ContractSpecs{
	FileName:     "Entrypoint.sol",
	ContractName: "Entrypoint",
	PackageName:  "entrypoint",
}

var (
	TickActionName  = "Tick"
	TickActionId    = crypto.Keccak256([]byte(SolidityActionMethodName(TickActionName) + "()"))[:4]
	TickActionIdHex = "0x" + common.Bytes2Hex(TickActionId)
)
