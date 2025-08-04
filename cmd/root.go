package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"transformer/internal/tui"
)

var (
	cfgFile string
	region  string
	output  string
	verbose bool
	tuiMode bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "transformer",
	Short: "AWS to OpenTofu Infrastructure as Code Transformer",
	Long: `A CLI tool to transform existing AWS infrastructure into OpenTofu (formerly Terraform) 
Infrastructure as Code (IaC) scripts. This tool discovers AWS resources and generates 
corresponding OpenTofu configuration files.`,
	RunE: runRoot,
}

// runRoot handles the root command execution
func runRoot(cmd *cobra.Command, args []string) error {
	if tuiMode {
		fmt.Println("Starting interactive TUI mode...")
		fmt.Println("Use arrow keys to navigate, Enter to select, Ctrl+C to quit")
		fmt.Println()
		return tui.RunTUI()
	}
	
	// If no subcommand is provided and TUI is not enabled, show help
	return cmd.Help()
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.transformer.yaml)")
	rootCmd.PersistentFlags().StringVar(&region, "region", "us-east-1", "AWS region")
	rootCmd.PersistentFlags().StringVar(&output, "output", "./infrastructure", "output directory for generated files")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&tuiMode, "tui", false, "start interactive terminal user interface")

	// Bind flags to viper
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("tui", rootCmd.PersistentFlags().Lookup("tui"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".transformer" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".transformer")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
} 