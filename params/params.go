package params

import "unicode"

var ValueParams = map[string]interface{}{
	"ActionExecutedEventName": ActionExecutedEventName,
	"MultiActionMethodName":   MultiActionMethodName,
	"IActionsContract":        IActionsContract,
	"ITablesContract":         ITablesContract,
	"ICoreContract":           ICoreContract,
	"EntrypointContract":      EntrypointContract,
}

var FunctionParams = map[string]interface{}{
	"ActionMethodNameFn": ActionMethodName,
	"ActionStructNameFn": ActionStructName,
	"TableMethodNameFn":  TableMethodName,
	"TableStructNameFn":  TableStructName,
}

const (
	ActionExecutedEventName = "ActionExecuted"
	MultiActionMethodName   = "executeMultipleActions"
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

func ActionMethodName(name string) string {
	return lowerFirstChar(name)
}

func ActionStructName(name string) string {
	return "ActionData_" + upperFirstChar(name)
}

func TableMethodName(name string) string {
	return "get" + upperFirstChar(name)
}

func TableStructName(name string) string {
	return "RowData_" + upperFirstChar(name)
}

type ContractSpecs struct {
	FileName     string
	ContractName string
}

var IActionsContract = ContractSpecs{
	FileName:     "IActions.sol",
	ContractName: "IActions",
}

var ITablesContract = ContractSpecs{
	FileName:     "ITables.sol",
	ContractName: "ITables",
}

var ICoreContract = ContractSpecs{
	FileName:     "ICore.sol",
	ContractName: "ICore",
}

var EntrypointContract = ContractSpecs{
	FileName:     "Entrypoint.sol",
	ContractName: "Entrypoint",
}
