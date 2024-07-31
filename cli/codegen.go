package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/naoina/toml"
	"github.com/spf13/viper"

	"github.com/concrete-eth/archetype/codegen"
	"github.com/concrete-eth/archetype/codegen/gogen"
	"github.com/concrete-eth/archetype/codegen/solgen"
	"github.com/concrete-eth/archetype/params"
	"github.com/ethereum/go-ethereum/concrete/codegen/datamod"
	"github.com/spf13/cobra"
)

const (
	GO_BIN              = "go"
	CONCRETE_BIN        = "concrete"
	GOFMT_BIN           = "gofmt"
	NODE_EXEC_BIN       = "npx"
	PRETTIER_BIN        = "prettier"
	PRETTIER_SOL_PLUGIN = "prettier-plugin-solidity"
	FORGE_BIN           = "forge"
	ABIGEN_BIN          = "abigen"
)

/* Logging */

func logTaskSuccess(name string, more ...any) {
	green.Print("[DONE] ")
	fmt.Print(name)
	if len(more) > 0 {
		gray.Print(": ")
		gray.Print(more...)
	}
	fmt.Println()
}

func logTaskFail(name string, err error) {
	red.Print("[FAIL] ")
	fmt.Print(name)
	if err != nil {
		fmt.Print(": ", err)
	}
	fmt.Println()
}

func logTaskSkip(name string, reason string) {
	yellow.Print("[SKIP] ")
	fmt.Print(name)
	fmt.Print(": ", reason)
	fmt.Println()
}

/* Environment utils */

// ensureDir creates a directory if it does not exist.
func ensureDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory: %s", dir)
	}
	return nil
}

// isInstalled checks if a command is installed by attempting to run it with a help flag (-h, --help, help).
func isInstalled(name string, args ...string) bool {
	// Attempt to run the command with a help flag
	for _, flag := range []string{"-h", "--help", "help"} {
		args := append(args, flag)
		if err := exec.Command(name, args...).Run(); err == nil {
			// If the command runs without error, it is installed
			return true
		}
	}
	return false
}

// isInGoModule checks if the current directory is in a go module.
func isInGoModule() bool {
	cmd := exec.Command(GO_BIN, "env", "GOMOD")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return false
	}
	gomod := strings.TrimSpace(out.String())
	return gomod != "" && gomod != "/dev/null"
}

// getGoModule returns the name of the go module in the current directory.
// e.g. github.com/user/repo
func getGoModule() (string, error) {
	if !isInGoModule() {
		return "", fmt.Errorf("not in a go module")
	}
	cmd := exec.Command(GO_BIN, "list", "-m")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// getGoModulePath returns the absolute root path of the go module in the current directory.
// e.g. /Users/user/path/to/module
func getGoModulePath() (string, error) {
	if !isInGoModule() {
		return "", fmt.Errorf("not in a go module")
	}
	cmd := exec.Command(GO_BIN, "list", "-m", "-f", "{{.Dir}}")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// getPackageImportPath returns the import path for a package at the given path.
// e.g. github.com/user/repo/path/to/package
func getPackageImportPath(path string) (string, error) {
	// Get the absolute path of the package
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	var modName, modPath, relDatamodPath string
	// Get the module name
	if modName, err = getGoModule(); err != nil {
		return "", err
	}
	// Get the module absolute root path
	if modPath, err = getGoModulePath(); err != nil {
		return "", err
	}
	// Get the relative path of the package from the module root
	if relDatamodPath, err = filepath.Rel(modPath, absPath); err != nil {
		return "", err
	}
	// Compose and return the import path
	return filepath.Join(modName, relDatamodPath), nil
}

/* Verbose */

// loadSchemasFromFile loads table schemas from a json file.
func loadSchemasFromFile(filePath string) ([]datamod.TableSchema, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return datamod.UnmarshalTableSchemas(data, false)
}

// printSchemasDescription prints a description of table schemas.
func printSchemasDescription(title string, schemas []datamod.TableSchema) {
	description := codegen.GenerateSchemasDescriptionString(schemas)
	bold.Println(title)
	fmt.Println(description)
}

/* Codegen */

// getGogenConfig returns a gogen config from viper settings.
func getGogenConfig() (gogen.Config, error) {
	var (
		actions = viper.GetString("actions")
		tables  = viper.GetString("tables")
		goOut   = viper.GetString("go-out")
		pkg     = viper.GetString("pkg")
		exp     = viper.GetBool("more-experimental")
	)

	var err error

	gogenOut := filepath.Join(goOut, "archmod") // <go-out>/archmod
	datamodOut := getDatamodOut()               // <go-out>/datamod

	datamodImportPath, err := getPackageImportPath(datamodOut)
	if err != nil {
		return gogen.Config{}, err
	}
	contractImportPath, err := getPackageImportPath(getAbigenOut())
	if err != nil {
		return gogen.Config{}, err
	}

	config := gogen.Config{
		Config: codegen.Config{
			ActionsJsonPath: actions,
			TablesJsonPath:  tables,
			Out:             gogenOut,
		},
		PackageName:         pkg,
		DatamodImportPath:   datamodImportPath,
		ContractsImportPath: contractImportPath,
		Experimental:        exp,
	}

	return config, nil
}

// getDatamodOut returns the output directory for the datamod command.
func getDatamodOut() string {
	goOut := viper.GetString("go-out")
	datamodOut := filepath.Join(goOut, "datamod") // <go-out>/datamod
	return datamodOut
}

// getAbigenOut returns the output directory for the abigen command.
func getAbigenOut() string {
	goOut := viper.GetString("go-out")
	abigenOut := filepath.Join(goOut, "abigen") // <go-out>/abigen
	return abigenOut
}

// getSolgenConfig returns the solgen config from the viper settings.
func getSolgenConfig() solgen.Config {
	var (
		actions = viper.GetString("actions")
		tables  = viper.GetString("tables")
		solOut  = viper.GetString("sol-out")
	)
	config := solgen.Config{
		Config: codegen.Config{
			ActionsJsonPath: actions,
			TablesJsonPath:  tables,
			Out:             solOut,
		},
	}
	return config
}

// getForgeBuildOut returns the output directory for the forge build command.
func getForgeBuildOut() string {
	forgeOut := viper.GetString("forge-out")
	return forgeOut
}

// runCodegen runs the full code generation process.
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

	// Get and validate codegen configs
	// Go codegen config
	gogenConfig, err := getGogenConfig()
	if err != nil {
		logFatal(err)
	}
	if err := ensureDir(gogenConfig.Out); err != nil {
		logFatal(err)
	}
	if err := gogenConfig.Validate(); err != nil {
		logFatal(err)
	}
	// Solidity codegen config
	solgenConfig := getSolgenConfig()
	if err := ensureDir(solgenConfig.Out); err != nil {
		logFatal(err)
	}
	if err := solgenConfig.Validate(); err != nil {
		logFatal(err)
	}

	// Preliminary checks
	if !isInstalled(GO_BIN) {
		logFatal(fmt.Errorf("go is not installed (go_bin=%s)", GO_BIN))
	}
	if !isInGoModule() {
		logFatal(fmt.Errorf("not in a go module"))
	}
	if !isInstalled(CONCRETE_BIN) {
		logFatal(fmt.Errorf("concrete cli is not installed (concrete_bin=%s)", CONCRETE_BIN))
	}
	if !isInstalled(FORGE_BIN) {
		logFatal(fmt.Errorf("forge cli is not installed (forge_bin=%s)", FORGE_BIN))
	}
	if !isInstalled(ABIGEN_BIN) {
		logFatal(fmt.Errorf("abigen is not installed (abigen_bin=%s)", ABIGEN_BIN))
	}

	var skipGofmt bool
	var skipPrettier bool
	var hasPreliminaryWarnings bool

	if isInstalled(GOFMT_BIN) {
		skipGofmt = false
	} else {
		logWarning(fmt.Sprintf("gofmt is not installed (gofmt_bin=%s). Install it to format the generated go code.", GOFMT_BIN))
		skipGofmt = true
		hasPreliminaryWarnings = true
	}

	var missing string
	if isInstalled(NODE_EXEC_BIN) {
		if isInstalled(NODE_EXEC_BIN, PRETTIER_BIN) {
			if isInstalled(NODE_EXEC_BIN, PRETTIER_BIN, "--plugin="+PRETTIER_SOL_PLUGIN) {
				skipPrettier = false
			} else {
				missing = PRETTIER_SOL_PLUGIN
			}
		} else {
			missing = PRETTIER_BIN
		}
	} else {
		missing = NODE_EXEC_BIN
	}
	if missing != "" {
		logWarning(fmt.Sprintf("%s is not installed. Install it to format the generated solidity code.", missing))
		skipPrettier = true
		hasPreliminaryWarnings = true
	}

	// Validate schema files
	actionsSchemas, err := loadSchemasFromFile(gogenConfig.ActionsJsonPath)
	if err != nil {
		logFatal(err)
	}
	tablesSchemas, err := loadSchemasFromFile(gogenConfig.TablesJsonPath)
	if err != nil {
		logFatal(err)
	}

	for _, schema := range actionsSchemas {
		if len(schema.Keys) > 0 {
			logWarning(fmt.Sprintf("Action %s has keys defined. Keys are not supported in actions.", schema.Name))
			hasPreliminaryWarnings = true
		}
	}

	if hasPreliminaryWarnings {
		fmt.Println("")
	}

	if verbose {
		// Print schema descriptions
		printSchemasDescription("Actions", actionsSchemas)
		fmt.Println("")
		printSchemasDescription("Tables", tablesSchemas)
		fmt.Println("")
	}

	// Run concrete datamod
	datamodPkg := "datamod"
	datamodOut := getDatamodOut()
	if err := ensureDir(datamodOut); err != nil {
		logFatal(err)
	}
	if err := runDatamod(datamodOut, gogenConfig.TablesJsonPath, datamodPkg, gogenConfig.Experimental); err != nil {
		logFatal(err)
	}
	// Run go and solidity codegen
	if err := runGogen(gogenConfig); err != nil {
		logFatal(err)
	}
	if err := runSolgen(solgenConfig); err != nil {
		logFatal(err)
	}

	// Run gofmt
	if !skipGofmt {
		runGofmt(datamodOut, gogenConfig.Out)
	} else {
		logTaskSkip("gofmt", "gofmt is not installed")
	}

	// Run prettier solidity plugin
	if !skipPrettier {
		runPrettier(solgenConfig.Out + "/**/*.sol")
	} else {
		logTaskSkip("prettier", "prettier is not installed")
	}

	// Run forge build
	forgeBuildOut := getForgeBuildOut()
	if err := runForgeBuild(solgenConfig.Out, forgeBuildOut); err != nil {
		logFatal(err)
	}

	// Run abigen on IActions and ITables
	abigenOut := getAbigenOut()
	for _, contract := range []params.ContractSpecs{params.EntrypointContract, params.ITablesContract} {
		var (
			inPath   = filepath.Join(forgeBuildOut, contract.FileName)
			dirName  = contract.PackageName
			fileName = contract.PackageName + ".go"
			outDir   = filepath.Join(abigenOut, dirName)
			outPath  = filepath.Join(outDir, fileName)
		)
		if err := ensureDir(outDir); err != nil {
			logFatal(err)
		}
		if err := runAbigen(contract.ContractName, inPath, outPath); err != nil {
			logFatal(err)
		}
	}

	// Done
	logInfo("\nCode generation completed successfully.")
	logInfo("Files written to: %s, %s", gogenConfig.Out, solgenConfig.Out)
	logDebug("\nDone in %v", time.Since(startTime))
}

// runCommand runs a command and logs it as a task, returning an error if the command fails.
func runCommand(name string, cmd *exec.Cmd) error {
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("%s failed: %w", name, err)
		logTaskFail(name, err)
		logDebug(strings.Join(cmd.Args, " "))
		logDebug(stdErr.String())
		return err
	}
	if viper.GetBool("verbose") {
		logTaskSuccess(name, strings.Join(cmd.Args, " "))
	} else {
		logTaskSuccess(name)
	}
	return nil
}

// runDatamod runs the concrete datamod command.
// Datamod generates type safe go wrappers for datastore structures from a JSON specification.
func runDatamod(outDir, tables, pkg string, experimental bool) error {
	taskName := "Concrete datamod"
	args := []string{"datamod", tables, "--pkg", pkg, "--out", outDir}
	if experimental {
		args = append(args, "--more-experimental")
	}
	cmd := exec.Command(CONCRETE_BIN, args...)
	return runCommand(taskName, cmd)
}

// runGogen runs the gogen codegen.
// config is assumed to be valid.
func runGogen(config gogen.Config) error {
	taskName := "Go"
	if err := gogen.Codegen(config); err != nil {
		err = fmt.Errorf("gogen failed: %w", err)
		logTaskFail(taskName, nil)
		return err
	}
	logTaskSuccess(taskName)
	return nil
}

// Run solgen codegen
// config is assumed to be valid
func runSolgen(config solgen.Config) error {
	taskName := "Solidity"
	if err := solgen.Codegen(config); err != nil {
		err = fmt.Errorf("solgen failed: %w", err)
		logTaskFail(taskName, nil)
		return err
	}
	logTaskSuccess(taskName)
	return nil
}

// runForgeBuild runs the forge build command.
func runForgeBuild(inDir, outDir string) error {
	taskName := "Forge build"
	cmd := exec.Command(
		FORGE_BIN, "build",
		"--contracts", inDir,
		"--out", outDir,
		"--extra-output-files", "bin", "abi",
	)
	return runCommand(taskName, cmd)
}

// runAbigen runs the abigen command.
func runAbigen(contractName, inPath, outPath string) error {
	var (
		pgk     = "contract"
		binPath = filepath.Join(inPath, contractName+".bin")
		abiDir  = filepath.Join(inPath, contractName+".abi.json")
	)
	taskName := "abigen: " + contractName
	cmd := exec.Command(ABIGEN_BIN, "--bin", binPath, "--abi", abiDir, "--pkg", pgk, "--out", outPath)
	return runCommand(taskName, cmd)
}

// runGofmt runs gofmt on the given directories.
func runGofmt(dirs ...string) error {
	taskName := "gofmt"
	args := append([]string{"-w"}, dirs...)
	cmd := exec.Command(GOFMT_BIN, args...)
	return runCommand(taskName, cmd)
}

// runPrettier runs prettier on the given patterns.
func runPrettier(patterns ...string) error {
	taskName := "prettier"
	args := []string{PRETTIER_BIN, "--plugin=" + PRETTIER_SOL_PLUGIN, "--write"}
	args = append(args, patterns...)
	cmd := exec.Command(NODE_EXEC_BIN, args...)
	return runCommand(taskName, cmd)
}

/* CLI */

// AddCodegenCommand
func AddCodegenCommand(parent *cobra.Command) {
	// Codegen command
	var cfgFile string
	codegenCmd := &cobra.Command{
		Use:   "codegen",
		Short: "Generate Golang definitions and Solidity interfaces for Archetype tables and actions from the given JSON specifications",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initConfig(cfgFile)
		},
		Run: runCodegen,
	}

	// Codegen flags
	codegenCmd.Flags().StringP("go-out", "g", "./codegen", "output directory")
	codegenCmd.Flags().StringP("sol-out", "s", "./codegen/sol", "output directory")
	codegenCmd.Flags().StringP("forge-out", "f", "./out", "forge output directory")
	codegenCmd.Flags().StringP("tables", "t", "./tables.json", "table schema file")
	codegenCmd.Flags().StringP("actions", "a", "./actions.json", "action schema file")
	codegenCmd.Flags().String("pkg", "archmod", "go package name")
	codegenCmd.Flags().BoolP("verbose", "v", false, "verbose output")
	codegenCmd.Flags().Bool("more-experimental", false, "enable experimental features")

	// Bind flags to viper
	viper.BindPFlag("go-out", codegenCmd.Flags().Lookup("go-out"))
	viper.BindPFlag("sol-out", codegenCmd.Flags().Lookup("sol-out"))
	viper.BindPFlag("forge-out", codegenCmd.Flags().Lookup("forge-out"))
	viper.BindPFlag("tables", codegenCmd.Flags().Lookup("tables"))
	viper.BindPFlag("actions", codegenCmd.Flags().Lookup("actions"))
	viper.BindPFlag("pkg", codegenCmd.Flags().Lookup("pkg"))
	viper.BindPFlag("verbose", codegenCmd.Flags().Lookup("verbose"))
	viper.BindPFlag("more-experimental", codegenCmd.Flags().Lookup("more-experimental"))

	parent.AddCommand(codegenCmd)
}

// initConfig loads the viper configuration from the given or default file and the environment.
// See https://github.com/spf13/viper?tab=readme-ov-file#why-viper for precedence order.
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
		// Log the config file used
		configFileAbsPath := viper.ConfigFileUsed()
		wd, err := os.Getwd()
		if err != nil {
			logFatal(err)
		}
		configFileRelPath, err := filepath.Rel(wd, configFileAbsPath)
		if err != nil {
			logFatal(err)
		}
		var configFilePathToPrint string
		if strings.HasPrefix(configFileRelPath, "..") {
			// If the config file is outside the working directory, print the absolute path
			configFilePathToPrint = configFileAbsPath
		} else {
			// Otherwise, print the relative path
			configFilePathToPrint = "./" + configFileRelPath
		}
		logDebug("Using config file: %s", configFilePathToPrint)
		fmt.Println("")
	} else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		logFatal(err)
	}
}
