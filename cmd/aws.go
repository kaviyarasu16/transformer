package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/kaviyarasu16/transformer/internal/aws"
	"github.com/kaviyarasu16/transformer/internal/generator"
)

var (
	resources string
	all       bool
)

// awsCmd represents the aws command
var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Transform AWS infrastructure to OpenTofu",
	Long: `Transform existing AWS infrastructure into OpenTofu (formerly Terraform) 
Infrastructure as Code (IaC) scripts. This command discovers AWS resources 
and generates corresponding OpenTofu configuration files.`,
	RunE: runAWSCommand,
}

func init() {
	rootCmd.AddCommand(awsCmd)

	// Local flags for aws command
	awsCmd.Flags().StringVar(&resources, "resources", "", "comma-separated list of resource types to discover (e.g., vpc,ec2,iam,rds)")
	awsCmd.Flags().BoolVar(&all, "all", false, "discover all supported resource types")
}

func runAWSCommand(cmd *cobra.Command, args []string) error {
	// Validate flags
	if !all && resources == "" {
		return fmt.Errorf("either --all or --resources must be specified")
	}

	if all && resources != "" {
		return fmt.Errorf("cannot use both --all and --resources flags together")
	}

	// Get configuration
	region := viper.GetString("region")
	outputDir := viper.GetString("output")
	verbose := viper.GetBool("verbose")

	if verbose {
		fmt.Println("Starting AWS to OpenTofu transformation...")
		fmt.Printf("Region: %s\n", region)
		if all {
			fmt.Println("Resources: all")
		} else {
			fmt.Printf("Resources: %s\n", resources)
		}
		fmt.Printf("Output directory: %s\n", outputDir)
	}

	// Initialize AWS client
	client, err := aws.NewClient(region)
	if err != nil {
		return fmt.Errorf("failed to initialize AWS client: %w", err)
	}

	// Determine resource types to discover
	var resourceTypes []string
	if all {
		resourceTypes = aws.GetAllSupportedResources()
	} else {
		resourceTypes = strings.Split(resources, ",")
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
		fmt.Println("Generating OpenTofu configuration...")
	}

	// Generate OpenTofu configuration
	gen := generator.NewGenerator(outputDir, verbose)
	err = gen.Generate(allResources, region)
	if err != nil {
		return fmt.Errorf("failed to generate OpenTofu configuration: %w", err)
	}

	if verbose {
		fmt.Println("OpenTofu configuration generation completed successfully!")
		fmt.Println()
		fmt.Println("✅ AWS to OpenTofu transformation completed successfully!")
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
		fmt.Println()
		fmt.Printf("🔍 Resources by type:\n")
		for resourceType, count := range resourceCounts {
			fmt.Printf("   • %s: %d\n", strings.ToUpper(resourceType), count)
		}
		fmt.Println()
		fmt.Printf("📁 Generated files:\n")
		fmt.Printf("   • main.tf - Main OpenTofu configuration\n")
		fmt.Printf("   • variables.tf - Variable definitions\n")
		fmt.Printf("   • outputs.tf - Output definitions\n")
		fmt.Printf("   • versions.tf - Provider version constraints\n")
		fmt.Printf("   • README.md - Documentation and usage guide\n")
		fmt.Printf("   • modules/ - Resource-specific modules\n")
		fmt.Println()
		fmt.Printf("🚀 Next steps:\n")
		fmt.Printf("   1. Review the generated configuration\n")
		fmt.Printf("   2. Update sensitive values (passwords, keys)\n")
		fmt.Printf("   3. Test in a non-production environment\n")
		fmt.Printf("   4. Run 'tofu init' and 'tofu plan'\n")
		fmt.Printf("   5. Apply the configuration with 'tofu apply'\n")
		fmt.Println()
		fmt.Printf("⚠️  Important notes:\n")
		fmt.Printf("   • Always review configurations before applying\n")
		fmt.Printf("   • Backup existing infrastructure before migration\n")
		fmt.Printf("   • Test thoroughly in a safe environment\n")
		fmt.Printf("   • Update any hardcoded values or references\n")
		fmt.Println()
	}

	return nil
} 