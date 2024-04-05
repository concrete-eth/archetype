package gogen

import (
	_ "embed"
	"errors"
	"path/filepath"

	"github.com/concrete-eth/archetype/codegen"
)

//go:embed templates/actions.go.tpl
var actionsTpl string

//go:embed templates/tables.go.tpl
var tablesTpl string

type Config struct {
	codegen.Config
	Package      string
	Datamod      string
	Experimental bool
}

func (c Config) Validate() error {
	if err := c.Config.Validate(); err != nil {
		return err
	}
	if c.Package == "" {
		return errors.New("package is required")
	}
	if c.Datamod == "" {
		return errors.New("datamod is required")
	}
	return nil
}

func GenerateActions(config Config) error {
	data := make(map[string]interface{})
	data["Package"] = config.Package
	data["Experimental"] = config.Experimental
	outPath := filepath.Join(config.Out, "actions.go")
	return codegen.ExecuteTemplate(actionsTpl, config.Actions, outPath, data, nil)
}

func GenerateTables(config Config) error {
	data := make(map[string]interface{})
	data["Package"] = config.Package
	data["Imports"] = []string{config.Datamod}
	data["Experimental"] = config.Experimental
	outPath := filepath.Join(config.Out, "tables.go")
	return codegen.ExecuteTemplate(tablesTpl, config.Tables, outPath, data, nil)
}

func Codegen(config Config) error {
	if err := config.Validate(); err != nil {
		return errors.New("error validating config for go code generation: " + err.Error())
	}
	if err := GenerateActions(config); err != nil {
		return errors.New("error generating go actions binding: " + err.Error())
	}
	if err := GenerateTables(config); err != nil {
		return errors.New("error generating go tables binding: " + err.Error())
	}
	return nil
}
