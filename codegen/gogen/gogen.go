package gogen

import (
	_ "embed"
	"errors"
	"html/template"
	"path/filepath"

	"github.com/concrete-eth/archetype/codegen"
	"github.com/concrete-eth/archetype/params"
)

//go:embed templates/types.go.tpl
var typesTpl string

//go:embed templates/actions.go.tpl
var actionsTpl string

//go:embed templates/tables.go.tpl
var tablesTpl string

type importSpecs struct {
	Name string
	Path string
}

type Config struct {
	codegen.Config
	PackageName         string
	ContractsImportPath string
	DatamodImportPath   string
	Experimental        bool
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	if err := c.Config.Validate(); err != nil {
		return err
	}
	if c.PackageName == "" {
		return errors.New("package name is required")
	}
	if c.DatamodImportPath == "" {
		return errors.New("datamod import path is required")
	}
	return nil
}

// GenerateActionTypes generates the go code for the action types.
func GenerateActionTypes(config Config) error {
	data := make(map[string]interface{})
	data["Package"] = config.PackageName
	funcMap := make(template.FuncMap)
	funcMap["StructNameFn"] = params.ActionStructName
	outPath := filepath.Join(config.Out, "action_types.go")
	return codegen.ExecuteTemplate(typesTpl, config.ActionsJsonPath, outPath, data, funcMap)
}

// GenerateActions generates the go code for the ActionSpecs.
func GenerateActions(config Config) error {
	data := make(map[string]interface{})
	data["Package"] = config.PackageName
	data["Imports"] = []importSpecs{
		{"contract", filepath.Join(config.ContractsImportPath, params.IActionsContract.PackageName)},
	}
	data["Experimental"] = config.Experimental
	outPath := filepath.Join(config.Out, "actions.go")
	return codegen.ExecuteTemplate(actionsTpl, config.ActionsJsonPath, outPath, data, nil)
}

// GenerateTableTypes generates the go code for the table types.
func GenerateTableTypes(config Config) error {
	data := make(map[string]interface{})
	data["Package"] = config.PackageName
	funcMap := make(template.FuncMap)
	funcMap["StructNameFn"] = params.TableStructName
	outPath := filepath.Join(config.Out, "table_types.go")
	return codegen.ExecuteTemplate(typesTpl, config.TablesJsonPath, outPath, data, funcMap)
}

// GenerateTables generates the go code for the TableSpecs.
func GenerateTables(config Config) error {
	data := make(map[string]interface{})
	data["Package"] = config.PackageName
	data["Imports"] = []importSpecs{
		{"", config.DatamodImportPath},
		{"contract", filepath.Join(config.ContractsImportPath, params.ITablesContract.PackageName)},
	}
	data["Experimental"] = config.Experimental
	outPath := filepath.Join(config.Out, "tables.go")
	return codegen.ExecuteTemplate(tablesTpl, config.TablesJsonPath, outPath, data, nil)
}

// Codegen generates the go code from the given config.
func Codegen(config Config) error {
	if err := config.Validate(); err != nil {
		return errors.New("error validating config for go code generation: " + err.Error())
	}
	if err := GenerateActionTypes(config); err != nil {
		return errors.New("error generating go action types binding: " + err.Error())
	}
	if err := GenerateActions(config); err != nil {
		return errors.New("error generating go actions binding: " + err.Error())
	}
	if err := GenerateTableTypes(config); err != nil {
		return errors.New("error generating go table types binding: " + err.Error())
	}
	if err := GenerateTables(config); err != nil {
		return errors.New("error generating go tables binding: " + err.Error())
	}
	return nil
}
