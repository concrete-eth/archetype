package solgen

import (
	_ "embed"
	"errors"
	"path/filepath"

	"github.com/concrete-eth/archetype/codegen"
	"github.com/concrete-eth/archetype/params"
)

//go:embed templates/tables.sol.tpl
var tablesTpl string

//go:embed templates/actions.sol.tpl
var actionsTpl string

//go:embed templates/core.sol.tpl
var coreTpl string

//go:embed templates/entrypoint.sol.tpl
var entrypointTpl string

type Config struct {
	codegen.Config
}

// GenerateActions generates the solidity interface from the actions schema.
func GenerateActions(config Config) error {
	data := make(map[string]interface{})
	data["Name"] = params.IActionsContract.ContractName
	outPath := filepath.Join(config.Out, params.IActionsContract.FileName)
	return codegen.ExecuteTemplate(actionsTpl, config.ActionsJsonPath, outPath, data, nil)
}

// GenerateTables generates the solidity interface from the tables schema.
func GenerateTables(config Config) error {
	data := make(map[string]interface{})
	data["Name"] = params.ITablesContract.ContractName
	outPath := filepath.Join(config.Out, params.ITablesContract.FileName)
	return codegen.ExecuteTemplate(tablesTpl, config.TablesJsonPath, outPath, data, nil)
}

// GenerateCore generates the core solidity interface.
func GenerateCore(config Config) error {
	data := make(map[string]interface{})
	data["Name"] = params.ICoreContract.ContractName
	data["Imports"] = []string{
		"./" + params.IActionsContract.FileName,
		"./" + params.ITablesContract.FileName,
	}
	data["Interfaces"] = []string{
		params.IActionsContract.ContractName,
		params.ITablesContract.ContractName,
	}
	outPath := filepath.Join(config.Out, params.ICoreContract.FileName)
	return codegen.ExecuteTemplate(coreTpl, "", outPath, data, nil)
}

// GenerateEntrypoint generates the entrypoint solidity abstract contract.
func GenerateEntrypoint(config Config) error {
	data := make(map[string]interface{})
	data["Name"] = params.EntrypointContract.ContractName
	data["Imports"] = []string{"./" + params.IActionsContract.FileName}
	data["Interfaces"] = []string{params.IActionsContract.ContractName}
	outPath := filepath.Join(config.Out, params.EntrypointContract.FileName)
	return codegen.ExecuteTemplate(entrypointTpl, config.ActionsJsonPath, outPath, data, nil)
}

// Codegen generates the solidity code from the given config.
func Codegen(config Config) error {
	if err := config.Validate(); err != nil {
		return errors.New("error validating config for solidity code generation: " + err.Error())
	}
	if err := GenerateActions(config); err != nil {
		return errors.New("error generating solidity actions interface: " + err.Error())
	}
	if err := GenerateTables(config); err != nil {
		return errors.New("error generating solidity tables interface: " + err.Error())
	}
	if err := GenerateCore(config); err != nil {
		return errors.New("error generating solidity core interface: " + err.Error())
	}
	if err := GenerateEntrypoint(config); err != nil {
		return errors.New("error generating solidity entrypoint: " + err.Error())
	}
	return nil
}
