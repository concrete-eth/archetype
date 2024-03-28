package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/naoina/toml"
	"github.com/spf13/viper"

	"github.com/concrete-eth/archetype/codegen"
	"github.com/concrete-eth/archetype/codegen/gogen"
	"github.com/concrete-eth/archetype/codegen/solgen"
	"github.com/ethereum/go-ethereum/concrete/codegen/datamod"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	CONCRETE_BIN = "concrete"
	GOFMT_BIN    = "gofmt"
)

/* Logging */

func logTaskSuccess(name string) {
	green := color.New(color.FgGreen)
	green.Print("[DONE] ")
	fmt.Println(name)
}

func logTaskFail(name string, err error) {
	red := color.New(color.FgRed)
	red.Print("[FAIL] ")
	fmt.Print(name)
	if err != nil {
		fmt.Println(": ", err)
	} else {
		fmt.Println()
	}
}

func logInfo(a ...any) {
	fmt.Println(a...)
}

func logDebug(a ...any) {
	gray := color.New(color.FgHiBlack)
	gray.Println(a...)
}

func logWarning(warning string) {
	yellow := color.New(color.FgYellow)
	yellow.Println("\nWarning:")
	fmt.Println(warning)
}

func logError(err error) {
	red := color.New(color.FgRed)
	fmt.Println("\nError:")
	red.Println(err)
	fmt.Println("\nContext:")
	logDebug(string(debug.Stack()))
	os.Exit(1)
}

func logFatal(err error) {
	logError(err)
	os.Exit(1)
}

/* Directory and PATH checks */

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

func isInstalled(cmd string) bool {
	err := exec.Command(cmd, "-h").Run()
	return err == nil
}

/* Verbose */

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

/* Codegen */

func getConfig(cmd *cobra.Command) codegen.Config {
	return codegen.Config{
		Actions: viper.GetString("actions"),
		Tables:  viper.GetString("tables"),
		Out:     viper.GetString("out"),
	}
}

func runCodegen(cmd *cobra.Command, args []string) {
	startTime := time.Now()

	verbose := viper.GetBool("verbose")

	if verbose {
		// Print settings
		allSettings := viper.AllSettings()
		settingsToml, err := toml.Marshal(allSettings)
		if err != nil {
			logFatal(err)
		}
		logDebug(string(settingsToml))
	}

	// Get basic config
	config := getConfig(cmd)
	if err := ensureDir(config.Out); err != nil {
		logFatal(err)
	}
	if err := config.Validate(); err != nil {
		logFatal(err)
	}

	if verbose {
		// Print schema descriptions
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

	// Run concrete datamod
	if isInstalled(CONCRETE_BIN) {
		runConcrete(cmd, args)
	} else {
		logFatal(fmt.Errorf("concrete cli is not installed"))
	}

	// Run go and solidity codegen
	runGogen(cmd, args)
	runSolgen(cmd, args)

	// Run gofmt
	if isInstalled(GOFMT_BIN) {
		runGofmt(config.Out)
	} else {
		logWarning("gofmt is not installed. Install it to format the generated code.")
	}

	// Done
	logInfo("\nCode generation completed successfully.")
	logInfo("Files written to:", config.Out)
	logDebug(fmt.Sprintf("\nDone in %v", time.Since(startTime)))
}

func runGogen(cmd *cobra.Command, args []string) {
	taskName := "Go"
	codegenConfig := getConfig(cmd)
	codegenConfig.Out = filepath.Join(codegenConfig.Out, "mod")
	if err := ensureDir(codegenConfig.Out); err != nil {
		logTaskFail(taskName, nil)
		logFatal(err)
	}
	config := gogen.Config{
		Config:  codegenConfig,
		Package: viper.GetString("pkg"),
		Datamod: viper.GetString("datamod"),
	}
	if err := gogen.Codegen(config); err != nil {
		logTaskFail(taskName, nil)
		logFatal(err)
	}
	logTaskSuccess(taskName)
}

func runSolgen(cmd *cobra.Command, args []string) {
	taskName := "Solidity"
	codegenConfig := getConfig(cmd)
	codegenConfig.Out = filepath.Join(codegenConfig.Out, "sol")
	if err := ensureDir(codegenConfig.Out); err != nil {
		logTaskFail(taskName, nil)
		logFatal(err)
	}
	config := solgen.Config{
		Config: codegenConfig,
	}
	if err := solgen.Codegen(config); err != nil {
		logTaskFail(taskName, nil)
		logFatal(err)
	}
	logTaskSuccess(taskName)
}

func runConcrete(cmd *cobra.Command, args []string) {
	taskName := "Concrete datamod"
	config := getConfig(cmd)
	outDir := filepath.Join(config.Out, "datamod")
	if err := ensureDir(outDir); err != nil {
		logTaskFail(taskName, nil)
		logFatal(err)
	}
	concreteCmd := exec.Command("concrete", "datamod", config.Tables, "--pkg", "datamod", "--out", outDir)
	var out bytes.Buffer
	concreteCmd.Stdout = &out
	if err := concreteCmd.Run(); err != nil {
		err = fmt.Errorf("concrete datamod failed: %w", err)
		logTaskFail(taskName, nil)
		logDebug(">", strings.Join(concreteCmd.Args, " "))
		logDebug(out.String())
		logFatal(err)
		return
	}
	logTaskSuccess(taskName)
}

func runGofmt(outDir string) {
	taskName := "gofmt"
	if err := exec.Command("gofmt", "-w", outDir).Run(); err != nil {
		err = fmt.Errorf("gofmt failed: %w", err)
		logTaskFail(taskName, err)
		return
	}
	logTaskSuccess(taskName)
}

func NewRootCmd() *cobra.Command {
	var cfgFile string
	var rootCmd = &cobra.Command{
		Use: "cli",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initConfig(cfgFile)
		},
	}
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.cli.yaml)")

	codegenCmd := &cobra.Command{Use: "codegen", Short: "generate code", Run: runCodegen}

	codegenCmd.Flags().StringP("out", "o", "./", "output directory")
	codegenCmd.Flags().StringP("tables", "t", "./tables.json", "table schema")
	codegenCmd.Flags().StringP("actions", "a", "./actions.json", "action schema")
	codegenCmd.Flags().String("datamod", "", "datamod module")
	codegenCmd.Flags().String("pkg", "model", "go package name")
	codegenCmd.Flags().BoolP("verbose", "v", false, "verbose output")

	viper.BindPFlag("out", codegenCmd.Flags().Lookup("out"))
	viper.BindPFlag("tables", codegenCmd.Flags().Lookup("tables"))
	viper.BindPFlag("actions", codegenCmd.Flags().Lookup("actions"))
	viper.BindPFlag("datamod", codegenCmd.Flags().Lookup("datamod"))
	viper.BindPFlag("pkg", codegenCmd.Flags().Lookup("pkg"))
	viper.BindPFlag("verbose", codegenCmd.Flags().Lookup("verbose"))

	rootCmd.AddCommand(codegenCmd)

	return rootCmd
}

func initConfig(cfgFile string) {
	// Get config from file
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in the working directory
		viper.AddConfigPath(".")
		viper.SetConfigName("arch")
	}

	// Get config from environment
	viper.SetEnvPrefix("ARCH")
	viper.AutomaticEnv()

	// Read config
	if err := viper.ReadInConfig(); err == nil {
		logDebug("Using config file:", viper.ConfigFileUsed())
		fmt.Println("")
	} else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		logError(err)
		os.Exit(1)
	}
}

func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		logFatal(err)
	}
}
