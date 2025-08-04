package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"transformer/internal/aws"
	"transformer/internal/generator"
)

var (
	importResources string
	importAll       bool
	importFile      string
	importState     string
	generateState   bool
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Generate OpenTofu import statements for AWS resources",
	Long: `Generate OpenTofu import statements for existing AWS infrastructure.
This command discovers AWS resources and generates corresponding import statements
that can be used to import existing resources into OpenTofu state.

Features:
- Generate import statements and resource definitions
- Create automated import scripts
- Generate state file templates (like Terraformer)
- Provide comprehensive import guides`,
	RunE: runImportCommand,
}

func init() {
	rootCmd.AddCommand(importCmd)

	// Local flags for import command
	importCmd.Flags().StringVar(&importResources, "resources", "", "comma-separated list of resource types to import (e.g., vpc,ec2,iam,rds)")
	importCmd.Flags().BoolVar(&importAll, "all", false, "import all supported resource types")
	importCmd.Flags().StringVar(&importFile, "file", "import.tf", "output file for import statements")
	importCmd.Flags().StringVar(&importState, "state", "", "OpenTofu state file path (optional)")
	importCmd.Flags().BoolVar(&generateState, "state-file", false, "generate a complete state file (like Terraformer)")
}

func runImportCommand(cmd *cobra.Command, args []string) error {
	// Validate flags
	if !importAll && importResources == "" {
		return fmt.Errorf("either --all or --resources must be specified")
	}

	if importAll && importResources != "" {
		return fmt.Errorf("cannot use both --all and --resources flags together")
	}

	// Get configuration
	region := viper.GetString("region")
	outputDir := viper.GetString("output")
	verbose := viper.GetBool("verbose")

	if verbose {
		fmt.Println("Starting AWS resource import generation...")
		fmt.Printf("Region: %s\n", region)
		if importAll {
			fmt.Println("Resources: all")
		} else {
			fmt.Printf("Resources: %s\n", importResources)
		}
		fmt.Printf("Output directory: %s\n", outputDir)
		fmt.Printf("Import file: %s\n", importFile)
		if importState != "" {
			fmt.Printf("State file: %s\n", importState)
		}
		if generateState {
			fmt.Println("State file generation: enabled")
		}
	}

	// Initialize AWS client
	client, err := aws.NewClient(region)
	if err != nil {
		return fmt.Errorf("failed to initialize AWS client: %w", err)
	}

	// Determine resource types to discover
	var resourceTypes []string
	if importAll {
		resourceTypes = aws.GetAllSupportedResources()
	} else {
		resourceTypes = strings.Split(importResources, ",")
		// Clean up resource types
		for i, rt := range resourceTypes {
			resourceTypes[i] = strings.TrimSpace(rt)
		}
	}

	if verbose {
		fmt.Println("Discovering AWS resources...")
	}

	// Discover resources
	var allResources []aws.Resource
	for _, resourceType := range resourceTypes {
		if verbose {
			fmt.Printf("Discovering %s resources...\n", resourceType)
		}

		resources, err := client.DiscoverResources([]string{resourceType})
		if err != nil {
			if verbose {
				fmt.Printf("Warning: Failed to discover %s resources: %v\n", resourceType, err)
			}
			continue
		}

		allResources = append(allResources, resources...)
	}

	if verbose {
		fmt.Printf("Discovered %d resources\n", len(allResources))
		fmt.Println("Generating import statements...")
	}

	// Generate import statements
	importGen := generator.NewImportGenerator(outputDir, verbose)
	
	// If state file generation is requested, set the state file path
	stateFilePath := importState
	if generateState && importState == "" {
		stateFilePath = "terraform.tfstate"
	}
	
	err = importGen.GenerateImportStatements(allResources, region, importFile, stateFilePath)
	if err != nil {
		return fmt.Errorf("failed to generate import statements: %w", err)
	}

	if verbose {
		fmt.Println("Import statement generation completed successfully!")
		fmt.Println()
		fmt.Println("✅ AWS to OpenTofu import generation completed successfully!")
		fmt.Println()
		fmt.Printf("📊 Summary:\n")
		fmt.Printf("   • Total resources discovered: %d\n", len(allResources))
		
		// Count resources by type
		resourceCounts := make(map[string]int)
		for _, resource := range allResources {
			resourceCounts[resource.GetType()]++
		}
		
		fmt.Printf("   • Resource types: %d\n", len(resourceCounts))
		fmt.Printf("   • Output directory: %s\n", outputDir)
		fmt.Printf("   • Import file: %s\n", importFile)
		if stateFilePath != "" {
			fmt.Printf("   • State file: %s\n", stateFilePath)
		}
		fmt.Println()
		fmt.Printf("🔍 Resources by type:\n")
		for resourceType, count := range resourceCounts {
			fmt.Printf("   • %s: %d\n", strings.ToUpper(resourceType), count)
		}
		fmt.Println()
		fmt.Printf("📁 Generated files:\n")
		fmt.Printf("   • %s - Import statements and resource definitions\n", importFile)
		fmt.Printf("   • import.sh - Automated import script\n")
		fmt.Printf("   • README.md - Import guide\n")
		if stateFilePath != "" {
			fmt.Printf("   • %s - State file template\n", stateFilePath)
		}
		fmt.Println()
		
		if generateState {
			fmt.Printf("🚀 Terraformer-like State File Generation:\n")
			fmt.Printf("   ✅ Complete state file generated: %s\n", stateFilePath)
			fmt.Printf("   ✅ Resource definitions included\n")
			fmt.Printf("   ✅ Import commands ready to execute\n")
			fmt.Printf("   ✅ Automated script provided\n")
			fmt.Println()
			fmt.Printf("🎯 Easy Import Process:\n")
			fmt.Printf("   1. Copy %s to terraform.tfstate\n", stateFilePath)
			fmt.Printf("   2. Run: tofu init\n")
			fmt.Printf("   3. Run: ./import.sh (automated import)\n")
			fmt.Printf("   4. Run: tofu plan (verify no changes)\n")
			fmt.Printf("   5. Run: tofu apply (finalize)\n")
		} else {
			fmt.Printf("🚀 Next steps:\n")
			fmt.Printf("   1. Review the generated import statements\n")
			fmt.Printf("   2. Run 'tofu init' to initialize OpenTofu\n")
			fmt.Printf("   3. Run 'tofu import' commands from the generated file\n")
			fmt.Printf("   4. Run 'tofu plan' to verify the import\n")
			fmt.Printf("   5. Run 'tofu apply' to finalize the import\n")
		}
		fmt.Println()
		fmt.Printf("⚠️  Important notes:\n")
		fmt.Printf("   • Always backup your existing state before importing\n")
		fmt.Printf("   • Test import process in a non-production environment\n")
		fmt.Printf("   • Verify resource configurations after import\n")
		fmt.Printf("   • Some resources may require manual configuration\n")
		fmt.Println()
	}

	return nil
} 