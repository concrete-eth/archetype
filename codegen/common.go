package codegen

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/concrete-eth/archetype/params"
	"github.com/ethereum/go-ethereum/concrete/codegen/datamod"
)

type Config struct {
	ActionsJsonPath string
	TablesJsonPath  string
	Out             string
}

func (c Config) Validate() error {
	if c.ActionsJsonPath == "" {
		return errors.New("actions schema is required")
	}
	if err := CheckFile(c.ActionsJsonPath); err != nil {
		return errors.New("error validating actions schema file: " + err.Error())
	}

	if c.TablesJsonPath == "" {
		return errors.New("tables schema is required")
	}
	if err := CheckFile(c.TablesJsonPath); err != nil {
		return errors.New("error validating tables schema file: " + err.Error())
	}

	if c.Out == "" {
		return errors.New("output directory is required")
	}
	if err := CheckDir(c.Out); err != nil {
		return errors.New("error validating output directory: " + err.Error())
	}

	return nil
}

func CheckFile(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("file is a directory")
	}
	return nil
}

func CheckDir(dirPath string) error {
	info, err := os.Stat(dirPath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("directory is a file")
	}
	return nil
}

var DefaultFuncMap = template.FuncMap{
	"_sub": func(a, b int) int { return a - b },
}

// ExecuteTemplate executes a template with the given data and writes the output to a file.
// It will load the JSON schema from the given path and add it to the data map.
func ExecuteTemplate(tplStr string, jsonSchemaPath, outPath string, data map[string]interface{}, funcMap template.FuncMap) error {
	// Load schemas
	if jsonSchemaPath != "" {
		jsonContent, err := os.ReadFile(jsonSchemaPath)
		if err != nil {
			return err
		}
		schemas, err := datamod.UnmarshalTableSchemas(jsonContent, false)
		if err != nil {
			return err
		}
		data["Schemas"] = schemas
		data["Json"] = string(jsonContent)
		data["Comment"] = GenerateSchemaDescriptionString(schemas)
	}

	data["ArchParams"] = params.ValueParams

	// Set funcMap
	if funcMap == nil {
		// Use default functions
		funcMap = DefaultFuncMap
	}
	// Add param functions without overriding existing ones
	for key, value := range params.FunctionParams {
		if _, ok := funcMap[key]; ok {
			continue
		}
		funcMap[key] = value
	}
	// Add default functions without overriding existing ones
	for key, value := range DefaultFuncMap {
		if _, ok := funcMap[key]; ok {
			continue
		}
		funcMap[key] = value
	}

	// Parse template
	tpl, err := template.New("template").Funcs(funcMap).Parse(tplStr)
	if err != nil {
		return err
	}

	// Execute template
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return err
	}

	// Write to file
	if err := os.WriteFile(outPath, buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func GenerateSchemaDescriptionString(schemas []datamod.TableSchema) string {
	sizeData := [][]string{{"Table", "KeySize", "ValueSize"}}
	for _, schema := range schemas {
		keySize := 0
		for _, field := range schema.Keys {
			keySize += field.Type.Size
		}
		valueSize := 0
		for _, field := range schema.Values {
			valueSize += field.Type.Size
		}
		sizeData = append(sizeData, []string{schema.Name, fmt.Sprintf("%d", keySize), fmt.Sprintf("%d", valueSize)})
	}
	return tabWrite(sizeData)
}

func tabWrite(data [][]string) string {
	var buffer bytes.Buffer
	w := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)
	for _, line := range data {
		fmt.Fprintln(w, line[0]+"\t"+line[1]+"\t"+line[2])
	}
	w.Flush()
	return strings.TrimSpace(buffer.String())
}
