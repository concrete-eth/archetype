package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/concrete-eth/archetype/codegen"
	"github.com/concrete-eth/archetype/codegen/gogen"
	"github.com/concrete-eth/archetype/codegen/solgen"
	"github.com/ethereum/go-ethereum/concrete/codegen/datamod"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func logSuccess(name string) {
	green := color.New(color.FgGreen)
	green.Print("[DONE] ")
	fmt.Println(name)
}

func logFail(name string) {
	red := color.New(color.FgRed)
	red.Print("[FAIL] ")
	fmt.Println(name)
}

func logError(err error) {
	red := color.New(color.FgRed)
	fmt.Println("\nError:")
	red.Println(err)
	fmt.Println("\nContext:")
	color.New(color.FgHiBlack).Println(string(debug.Stack()))
	os.Exit(1)
}

func logFatal(err error) {
	logError(err)
	os.Exit(1)
}

func ensureDir(dirName string) error {
	info, err := os.Stat(dirName)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(dirName, 0755)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory: %s", dirName)
	}
	return nil
}

func runCodegen(cmd *cobra.Command, args []string) {
	startTime := time.Now()

	config := getConfig(cmd)
	if err := config.Validate(); err != nil {
		logFatal(err)
	}

	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		actionsSchema, err := loadSchemaFromFile(config.Actions)
		if err != nil {
			logFatal(err)
		}
		tablesSchema, err := loadSchemaFromFile(config.Tables)
		if err != nil {
			logFatal(err)
		}
		printSchemaDescription("Actions", actionsSchema)
		fmt.Println("")
		printSchemaDescription("Tables", tablesSchema)
		fmt.Println("")
	}

	runGogen(cmd, args)
	runSolgen(cmd, args)

	fmt.Println("\nFiles written to:", config.Out)

	color.New(color.FgHiBlack).Printf("\nDone in %v.\n", time.Since(startTime))
}

func loadSchemaFromFile(filePath string) ([]datamod.TableSchema, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return datamod.UnmarshalTableSchemas(data, false)
}

func printSchemaDescription(title string, schema []datamod.TableSchema) {
	var (
		description = codegen.GenerateSchemaDescriptionString(schema)
		clrVal      = color.FgWhite
		clr         = color.New(clrVal)
		bold        = color.New(clrVal, color.Bold)
	)
	bold.Println(title)
	clr.Println(description)
}

func runGogen(cmd *cobra.Command, args []string) {
	taskName := "Go"
	codegenConfig := getConfig(cmd)
	codegenConfig.Out = filepath.Join(codegenConfig.Out, "mod")
	if err := ensureDir(codegenConfig.Out); err != nil {
		logFail(taskName)
		logFatal(err)
	}
	config := gogen.Config{
		Config:  codegenConfig,
		Package: cmd.Flag("pkg").Value.String(),
		Datamod: cmd.Flag("datamod").Value.String(),
	}
	if err := gogen.Codegen(config); err != nil {
		logFail(taskName)
		logFatal(err)
	}
	logSuccess(taskName)
}

func runSolgen(cmd *cobra.Command, args []string) {
	taskName := "Solidity"
	codegenConfig := getConfig(cmd)
	codegenConfig.Out = filepath.Join(codegenConfig.Out, "sol")
	if err := ensureDir(codegenConfig.Out); err != nil {
		logFail(taskName)
		logFatal(err)
	}
	config := solgen.Config{
		Config: codegenConfig,
	}
	if err := solgen.Codegen(config); err != nil {
		logFail(taskName)
		logFatal(err)
	}
	logSuccess(taskName)
}

func getConfig(cmd *cobra.Command) codegen.Config {
	return codegen.Config{
		Actions: cmd.Flag("actions").Value.String(),
		Tables:  cmd.Flag("tables").Value.String(),
		Out:     cmd.Flag("out").Value.String(),
	}
}

func NewRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{Use: "cli"}

	codegenCmd := &cobra.Command{Use: "codegen", Short: "generate code", Run: runCodegen}
	codegenCmd.Flags().StringP("out", "o", "./", "output directory")
	codegenCmd.Flags().StringP("tables", "t", "./tables.json", "table schema")
	codegenCmd.Flags().StringP("actions", "a", "./actions.json", "action schema")
	codegenCmd.Flags().String("datamod", "", "datamod module")
	codegenCmd.Flags().String("pkg", "model", "go package name")
	codegenCmd.Flags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.AddCommand(codegenCmd)

	return rootCmd
}

func main() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		logFatal(err)
	}
}
