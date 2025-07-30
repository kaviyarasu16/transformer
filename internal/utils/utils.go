package utils

import (
	"fmt"
	"strings"

	"github.com/kaviyarasu16/transformer/internal/aws"
)

// GenerateSummary creates a summary of the transformation process
func GenerateSummary(resources []aws.Resource, outputDir string) string {
	// Count resources by type
	resourceCounts := make(map[string]int)
	for _, resource := range resources {
		resourceType := resource.GetType()
		resourceCounts[resourceType]++
	}

	// Build summary
	var summary strings.Builder
	summary.WriteString("✅ AWS to OpenTofu transformation completed successfully!\n\n")
	summary.WriteString("📊 Summary:\n")
	summary.WriteString(fmt.Sprintf("   • Total resources discovered: %d\n", len(resources)))
	summary.WriteString(fmt.Sprintf("   • Resource types: %d\n", len(resourceCounts)))
	summary.WriteString(fmt.Sprintf("   • Output directory: %s\n\n", outputDir))

	summary.WriteString("🔍 Resources by type:\n")
	for resourceType, count := range resourceCounts {
		summary.WriteString(fmt.Sprintf("   • %s: %d\n", strings.ToUpper(resourceType), count))
	}

	summary.WriteString("\n📁 Generated files:\n")
	summary.WriteString("   • main.tf - Main OpenTofu configuration\n")
	summary.WriteString("   • variables.tf - Variable definitions\n")
	summary.WriteString("   • outputs.tf - Output definitions\n")
	summary.WriteString("   • versions.tf - Provider version constraints\n")
	summary.WriteString("   • README.md - Documentation and usage guide\n")
	summary.WriteString("   • modules/ - Resource-specific modules\n")

	summary.WriteString("\n🚀 Next steps:\n")
	summary.WriteString("   1. Review the generated configuration\n")
	summary.WriteString("   2. Update sensitive values (passwords, keys)\n")
	summary.WriteString("   3. Test in a non-production environment\n")
	summary.WriteString("   4. Run 'tofu init' and 'tofu plan'\n")
	summary.WriteString("   5. Apply the configuration with 'tofu apply'\n")

	summary.WriteString("\n⚠️  Important notes:\n")
	summary.WriteString("   • Always review configurations before applying\n")
	summary.WriteString("   • Backup existing infrastructure before migration\n")
	summary.WriteString("   • Test thoroughly in a safe environment\n")
	summary.WriteString("   • Update any hardcoded values or references\n")

	return summary.String()
}

// ValidateRegion checks if the provided region is valid
func ValidateRegion(region string) error {
	validRegions := []string{
		"us-east-1", "us-east-2", "us-west-1", "us-west-2",
		"af-south-1", "ap-east-1", "ap-south-1", "ap-northeast-1",
		"ap-northeast-2", "ap-northeast-3", "ap-southeast-1", "ap-southeast-2",
		"ap-southeast-3", "ap-southeast-4", "ca-central-1", "eu-central-1",
		"eu-west-1", "eu-west-2", "eu-west-3", "eu-north-1", "eu-south-1",
		"eu-south-2", "me-south-1", "me-central-1", "sa-east-1",
		"us-gov-east-1", "us-gov-west-1",
	}

	for _, validRegion := range validRegions {
		if region == validRegion {
			return nil
		}
	}

	return fmt.Errorf("invalid region: %s. Valid regions: %s", region, strings.Join(validRegions, ", "))
}

// SanitizeResourceName converts a resource name to a valid OpenTofu resource name
func SanitizeResourceName(name string) string {
	// Replace invalid characters with underscores
	invalidChars := []string{" ", "-", ".", "/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Ensure it starts with a letter
	if len(result) > 0 && (result[0] < 'a' || result[0] > 'z') && (result[0] < 'A' || result[0] > 'Z') {
		result = "resource_" + result
	}

	// Limit length
	if len(result) > 63 {
		result = result[:63]
	}

	return result
}

// FormatTags formats tags for OpenTofu configuration
func FormatTags(tags map[string]string) string {
	if len(tags) == 0 {
		return "  tags = {}\n"
	}

	var result strings.Builder
	result.WriteString("  tags = {\n")
	
	// Sort keys for consistent output
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	
	for _, key := range keys {
		value := tags[key]
		result.WriteString(fmt.Sprintf("    %s = \"%s\"\n", key, value))
	}
	
	result.WriteString("  }\n")
	return result.String()
}

// EscapeString escapes special characters in strings for OpenTofu
func EscapeString(s string) string {
	// Replace backslashes and quotes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// GenerateResourceID creates a unique resource ID
func GenerateResourceID(resourceType, name string) string {
	return fmt.Sprintf("%s_%s", resourceType, SanitizeResourceName(name))
}

// IsValidResourceType checks if a resource type is supported
func IsValidResourceType(resourceType string) bool {
	validTypes := aws.GetAllSupportedResources()
	for _, validType := range validTypes {
		if resourceType == validType {
			return true
		}
	}
	return false
}

// GetResourceTypeDisplayName returns a human-readable name for a resource type
func GetResourceTypeDisplayName(resourceType string) string {
	displayNames := map[string]string{
		"vpc":           "VPC",
		"ec2":           "EC2 Instance",
		"iam":           "IAM Role",
		"rds":           "RDS Database",
		"s3":            "S3 Bucket",
		"alb":           "Application Load Balancer",
		"elb":           "Elastic Load Balancer",
		"asg":           "Auto Scaling Group",
		"cloudwatch":    "CloudWatch",
		"cloudtrail":    "CloudTrail",
		"ecs":           "ECS Service",
		"eks":           "EKS Cluster",
		"lambda":        "Lambda Function",
		"sqs":           "SQS Queue",
		"sns":           "SNS Topic",
		"elasticache":   "ElastiCache",
		"redshift":      "Redshift Cluster",
		"route53":       "Route53",
		"cloudfront":    "CloudFront",
		"apigateway":    "API Gateway",
		"dynamodb":      "DynamoDB",
		"elasticsearch": "Elasticsearch",
		"opensearch":    "OpenSearch",
		"neptune":       "Neptune",
		"docdb":         "DocumentDB",
		"elasticbeanstalk": "Elastic Beanstalk",
		"ecr":           "ECR Repository",
		"codecommit":    "CodeCommit",
		"codebuild":     "CodeBuild",
		"codedeploy":    "CodeDeploy",
		"codepipeline":  "CodePipeline",
		"cloudformation": "CloudFormation",
		"ssm":           "Systems Manager",
		"secretsmanager": "Secrets Manager",
		"kms":           "KMS",
		"guardduty":     "GuardDuty",
		"config":        "Config",
		"backup":        "Backup",
		"glacier":       "Glacier",
		"glue":          "Glue",
		"athena":        "Athena",
		"quicksight":    "QuickSight",
		"workspaces":    "WorkSpaces",
		"directoryservice": "Directory Service",
		"fsx":           "FSx",
		"storagegateway": "Storage Gateway",
		"transfer":      "Transfer",
		"mq":            "MQ",
		"kinesis":       "Kinesis",
		"firehose":      "Firehose",
		"mediastore":    "MediaStore",
		"mediaconvert":  "MediaConvert",
		"medialive":     "MediaLive",
		"mediatailor":   "MediaTailor",
		"iot":           "IoT",
		"greengrass":    "Greengrass",
		"greengrassv2":  "Greengrass V2",
		"iotanalytics":  "IoT Analytics",
		"iotevents":     "IoT Events",
		"iotsitewise":   "IoT SiteWise",
		"iotthingsgraph": "IoT Things Graph",
		"iotwireless":   "IoT Wireless",
		"iotdeviceadvisor": "IoT Device Advisor",
		"iotfleethub":   "IoT Fleet Hub",
		"iotsecuretunneling": "IoT Secure Tunneling",

	}

	if displayName, exists := displayNames[resourceType]; exists {
		return displayName
	}

	// Fallback: capitalize and replace underscores
	return strings.ReplaceAll(strings.Title(resourceType), "_", " ")
}

// FormatResourceCount formats a resource count with proper pluralization
func FormatResourceCount(count int, resourceType string) string {
	if count == 1 {
		return fmt.Sprintf("1 %s", GetResourceTypeDisplayName(resourceType))
	}
	return fmt.Sprintf("%d %s resources", count, GetResourceTypeDisplayName(resourceType))
}

// ValidateAWSResourceName validates if a resource name follows AWS naming conventions
func ValidateAWSResourceName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("resource name cannot be empty")
	}
	
	if len(name) > 128 {
		return fmt.Errorf("resource name cannot exceed 128 characters")
	}
	
	// Check for valid characters (alphanumeric, hyphens, underscores)
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '-' || char == '_') {
			return fmt.Errorf("resource name contains invalid character: %c", char)
		}
	}
	
	// Check if it starts with a letter or number
	if name[0] == '-' || name[0] == '_' {
		return fmt.Errorf("resource name cannot start with '-' or '_'")
	}
	
	return nil
}

// GenerateResourceReference creates a proper OpenTofu resource reference
func GenerateResourceReference(resourceType, name string) string {
	sanitizedName := SanitizeResourceName(name)
	return fmt.Sprintf("aws_%s.%s", resourceType, sanitizedName)
}

// ExtractResourceNameFromARN extracts the resource name from an AWS ARN
func ExtractResourceNameFromARN(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) >= 6 {
		// For most resources, the name is the last part
		lastPart := parts[len(parts)-1]
		// Remove any additional path components
		if strings.Contains(lastPart, "/") {
			pathParts := strings.Split(lastPart, "/")
			return pathParts[len(pathParts)-1]
		}
		return lastPart
	}
	return arn
}

// IsGlobalService checks if an AWS service is global (not region-specific)
func IsGlobalService(service string) bool {
	globalServices := []string{
		"iam", "route53", "cloudfront", "waf", "wafv2", "shield",
		"organizations", "billing", "budgets", "ce", "cur",
	}
	
	for _, globalService := range globalServices {
		if service == globalService {
			return true
		}
	}
	return false
}

// GenerateModuleName generates a valid module name from a resource type
func GenerateModuleName(resourceType string) string {
	// Replace any special characters with underscores
	moduleName := strings.ReplaceAll(resourceType, "-", "_")
	moduleName = strings.ReplaceAll(moduleName, ".", "_")
	
	// Ensure it starts with a letter
	if len(moduleName) > 0 && (moduleName[0] < 'a' || moduleName[0] > 'z') && (moduleName[0] < 'A' || moduleName[0] > 'Z') {
		moduleName = "module_" + moduleName
	}
	
	return moduleName
}

// FormatOpenTofuBlock formats a block of OpenTofu configuration
func FormatOpenTofuBlock(blockType, name string, attributes map[string]interface{}, indent int) string {
	var result strings.Builder
	
	// Add indentation
	indentStr := strings.Repeat("  ", indent)
	
	// Start block
	result.WriteString(fmt.Sprintf("%s%s \"%s\" {\n", indentStr, blockType, name))
	
	// Add attributes
	for key, value := range attributes {
		switch v := value.(type) {
		case string:
			result.WriteString(fmt.Sprintf("%s  %s = \"%s\"\n", indentStr, key, EscapeString(v)))
		case bool:
			result.WriteString(fmt.Sprintf("%s  %s = %t\n", indentStr, key, v))
		case int:
			result.WriteString(fmt.Sprintf("%s  %s = %d\n", indentStr, key, v))
		case float64:
			result.WriteString(fmt.Sprintf("%s  %s = %f\n", indentStr, key, v))
		case []string:
			if len(v) > 0 {
				result.WriteString(fmt.Sprintf("%s  %s = [\n", indentStr, key))
				for _, item := range v {
					result.WriteString(fmt.Sprintf("%s    \"%s\",\n", indentStr, EscapeString(item)))
				}
				result.WriteString(fmt.Sprintf("%s  ]\n", indentStr))
			} else {
				result.WriteString(fmt.Sprintf("%s  %s = []\n", indentStr, key))
			}
		case map[string]string:
			if len(v) > 0 {
				result.WriteString(fmt.Sprintf("%s  %s = {\n", indentStr, key))
				for k, val := range v {
					result.WriteString(fmt.Sprintf("%s    %s = \"%s\"\n", indentStr, k, EscapeString(val)))
				}
				result.WriteString(fmt.Sprintf("%s  }\n", indentStr))
			} else {
				result.WriteString(fmt.Sprintf("%s  %s = {}\n", indentStr, key))
			}
		default:
			result.WriteString(fmt.Sprintf("%s  %s = %v\n", indentStr, key, v))
		}
	}
	
	// End block
	result.WriteString(fmt.Sprintf("%s}\n", indentStr))
	
	return result.String()
}

// ValidateOpenTofuConfiguration validates basic OpenTofu configuration syntax
func ValidateOpenTofuConfiguration(config string) error {
	// Basic validation - check for common syntax errors
	if strings.Contains(config, "{{") || strings.Contains(config, "}}") {
		return fmt.Errorf("configuration contains template syntax that should be escaped")
	}
	
	if strings.Contains(config, "\\n") {
		return fmt.Errorf("configuration contains escaped newlines that should be properly formatted")
	}
	
	// Check for balanced braces
	openBraces := strings.Count(config, "{")
	closeBraces := strings.Count(config, "}")
	if openBraces != closeBraces {
		return fmt.Errorf("unbalanced braces in configuration")
	}
	
	return nil
} 