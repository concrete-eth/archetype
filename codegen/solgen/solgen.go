package solgen

import (
	_ "embed"
	"errors"
	"path/filepath"

	"github.com/concrete-eth/archetype/codegen"
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

func (c Config) Validate() error {
	if err := c.Config.Validate(); err != nil {
		return err
	}
	return nil
}

func GenerateActions(config Config) error {
	data := make(map[string]interface{})
	data["Name"] = "IActionExecutor"
	outPath := filepath.Join(config.Out, "IActions.sol")
	return codegen.ExecuteTemplate(actionsTpl, config.Actions, outPath, data, nil)
}

func GenerateTables(config Config) error {
	data := make(map[string]interface{})
	data["Name"] = "ITableGetter"
	outPath := filepath.Join(config.Out, "ITables.sol")
	return codegen.ExecuteTemplate(tablesTpl, config.Tables, outPath, data, nil)
}

func GenerateCore(config Config) error {
	data := make(map[string]interface{})
	data["Name"] = "ICore"
	data["Imports"] = []string{
		"./ITables.sol",
		"./IActions.sol",
	}
	data["Interfaces"] = []string{
		"ITableGetter",
		"IActionExecutor",
	}
	outPath := filepath.Join(config.Out, "ICore.sol")
	return codegen.ExecuteTemplate(coreTpl, "", outPath, data, nil)
}

func GenerateEntrypoint(config Config) error {
	data := make(map[string]interface{})
	data["Name"] = "Entrypoint"
	data["Imports"] = []string{
		"./IActions.sol",
	}
	data["Interfaces"] = []string{
		"IActionExecutor",
	}
	outPath := filepath.Join(config.Out, "Entrypoint.sol")
	return codegen.ExecuteTemplate(entrypointTpl, config.Actions, outPath, data, nil)
}

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
