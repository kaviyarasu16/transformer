package aws

import (
	"fmt"
	"strings"
)

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

	// Limit length but preserve uniqueness by keeping the end
	if len(result) > 63 {
		// Keep the first 50 characters and last 13 characters to preserve uniqueness
		prefix := result[:50]
		suffix := result[len(result)-13:]
		result = prefix + "_" + suffix
	}

	return result
}

// Resource interface defines methods that all AWS resources must implement
type Resource interface {
	GetType() string
	GetID() string
	GetName() string
	GetRegion() string
	GetTags() map[string]string
	ToOpenTofu() (string, error)
	GetDependencies() []string
}

// BaseResource contains common fields for all AWS resources
type BaseResource struct {
	Type         string            `json:"type"`
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Region       string            `json:"region"`
	Tags         map[string]string `json:"tags"`
	Dependencies []string          `json:"dependencies"`
}

func (r *BaseResource) GetType() string {
	return r.Type
}

func (r *BaseResource) GetID() string {
	return r.ID
}

func (r *BaseResource) GetName() string {
	return r.Name
}

func (r *BaseResource) GetRegion() string {
	return r.Region
}

func (r *BaseResource) GetTags() map[string]string {
	return r.Tags
}

func (r *BaseResource) GetDependencies() []string {
	return r.Dependencies
}

// VPCResource represents an AWS VPC
type VPCResource struct {
	BaseResource
	CIDRBlock           string   `json:"cidr_block"`
	EnableDNSHostnames  bool     `json:"enable_dns_hostnames"`
	EnableDNSSupport    bool     `json:"enable_dns_support"`
	InstanceTenancy     string   `json:"instance_tenancy"`
	IPv6CIDRBlock       string   `json:"ipv6_cidr_block,omitempty"`
	AssignGeneratedIPv6 bool     `json:"assign_generated_ipv6,omitempty"`
	Subnets             []string `json:"subnets,omitempty"`
}

func (r *VPCResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	// Build the base template
	content := fmt.Sprintf(`resource "aws_vpc" "%s" {
  cidr_block           = "%s"
  enable_dns_hostnames = %t
  enable_dns_support   = %t
  instance_tenancy     = "%s"
  tags = {`,
		resourceName,
		r.CIDRBlock,
		r.EnableDNSHostnames,
		r.EnableDNSSupport,
		r.InstanceTenancy)

	// Add tags
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
			// Quote tag keys that contain special characters
			quotedKey := k
			if strings.ContainsAny(k, ":-") {
				quotedKey = fmt.Sprintf(`"%s"`, k)
			}
			content += fmt.Sprintf(`
    %s = "%s"`, quotedKey, v)
		}
	}
	
	content += `
  }
}`

	// Add IPv6 configuration if present
	if r.IPv6CIDRBlock != "" {
		content += fmt.Sprintf(`

resource "aws_vpc_ipv6_cidr_block_association" "%s_ipv6" {
  vpc_id = aws_vpc.%s.id
  ipv6_cidr_block = "%s"
  assign_generated_ipv6_cidr_block = %t
}`, resourceName, resourceName, r.IPv6CIDRBlock, r.AssignGeneratedIPv6)
	}

	return content, nil
}

// EC2InstanceResource represents an AWS EC2 instance
type EC2InstanceResource struct {
	BaseResource
	InstanceType     string            `json:"instance_type"`
	AMI              string            `json:"ami"`
	KeyName          string            `json:"key_name,omitempty"`
	SubnetID         string            `json:"subnet_id"`
	SecurityGroups   []string          `json:"security_groups"`
	UserData         string            `json:"user_data,omitempty"`
	IAMInstanceProfile string          `json:"iam_instance_profile,omitempty"`
	RootBlockDevice  *BlockDeviceMapping `json:"root_block_device,omitempty"`
	EBSBlockDevices  []*BlockDeviceMapping `json:"ebs_block_devices,omitempty"`
}

type BlockDeviceMapping struct {
	DeviceName string `json:"device_name"`
	VolumeSize int    `json:"volume_size"`
	VolumeType string `json:"volume_type"`
	Encrypted  bool   `json:"encrypted"`
	KMSKeyID   string `json:"kms_key_id,omitempty"`
}

func (r *EC2InstanceResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	// Build the base template
	content := fmt.Sprintf(`resource "aws_instance" "%s" {
  ami           = "%s"
  instance_type = "%s"
  subnet_id     = "%s"`,
		resourceName,
		r.AMI,
		r.InstanceType,
		r.SubnetID)

	// Add optional fields
	if r.KeyName != "" {
		content += fmt.Sprintf(`
  key_name = "%s"`, r.KeyName)
	}

	if r.IAMInstanceProfile != "" {
		content += fmt.Sprintf(`
  iam_instance_profile = "%s"`, r.IAMInstanceProfile)
	}

	if r.UserData != "" {
		content += fmt.Sprintf(`
  user_data = "%s"`, r.UserData)
	}

	// Add security groups
	if len(r.SecurityGroups) > 0 {
		content += `
  vpc_security_group_ids = [`
		for i, sg := range r.SecurityGroups {
			if i > 0 {
				content += ", "
			}
			content += fmt.Sprintf(`"%s"`, sg)
		}
		content += `]`
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
			// Quote tag keys that contain special characters
			quotedKey := k
			if strings.ContainsAny(k, ":-") {
				quotedKey = fmt.Sprintf(`"%s"`, k)
			}
			content += fmt.Sprintf(`
    %s = "%s"`, quotedKey, v)
		}
	}
	content += `
  }`

	// Add root block device
	if r.RootBlockDevice != nil {
		content += fmt.Sprintf(`
  root_block_device {
    volume_size = %d
    volume_type = "%s"
    encrypted    = %t`,
			r.RootBlockDevice.VolumeSize,
			r.RootBlockDevice.VolumeType,
			r.RootBlockDevice.Encrypted)
		
		if r.RootBlockDevice.KMSKeyID != "" {
			content += fmt.Sprintf(`
    kms_key_id = "%s"`, r.RootBlockDevice.KMSKeyID)
		}
		content += `
  }`
	}

	// Add EBS block devices
	if len(r.EBSBlockDevices) > 0 {
		for _, device := range r.EBSBlockDevices {
			content += fmt.Sprintf(`
  ebs_block_device {
    device_name = "%s"
    volume_size = %d
    volume_type = "%s"
    encrypted    = %t`,
				device.DeviceName,
				device.VolumeSize,
				device.VolumeType,
				device.Encrypted)
			
			if device.KMSKeyID != "" {
				content += fmt.Sprintf(`
    kms_key_id = "%s"`, device.KMSKeyID)
			}
			content += `
  }`
		}
	}

	content += `
}`

	return content, nil
}

// IAMRoleResource represents an AWS IAM role
type IAMRoleResource struct {
	BaseResource
	AssumeRolePolicy string   `json:"assume_role_policy"`
	ManagedPolicyARNs []string `json:"managed_policy_arns"`
	InlinePolicies    map[string]string `json:"inline_policies"`
	Path              string   `json:"path"`
	Description       string   `json:"description"`
}

func (r *IAMRoleResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	// Build the main role resource
	content := fmt.Sprintf(`resource "aws_iam_role" "%s" {
  name = "%s"
  path = "%s"
  description = "%s"
  assume_role_policy = %s
  tags = {`,
		resourceName,
		r.Name,
		r.Path,
		r.Description,
		r.AssumeRolePolicy)

	// Add tags
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
			// Quote tag keys that contain special characters
			quotedKey := k
			if strings.ContainsAny(k, ":-") {
				quotedKey = fmt.Sprintf(`"%s"`, k)
			}
			content += fmt.Sprintf(`
    %s = "%s"`, quotedKey, v)
		}
	}
	
	content += `
  }
}`

	// Add managed policy attachments
	for _, policyARN := range r.ManagedPolicyARNs {
		policyName := strings.ReplaceAll(policyARN, ":", "_")
		policyName = strings.ReplaceAll(policyName, "/", "_")
		content += fmt.Sprintf(`

resource "aws_iam_role_policy_attachment" "%s_%s" {
  role       = aws_iam_role.%s.name
  policy_arn = "%s"
}`, resourceName, policyName, resourceName, policyARN)
	}

	// Add inline policies
	for policyName, policyDocument := range r.InlinePolicies {
		inlinePolicyName := strings.ReplaceAll(policyName, "-", "_")
		content += fmt.Sprintf(`

resource "aws_iam_role_policy" "%s_%s" {
  name = "%s"
  role = aws_iam_role.%s.id
  policy = %s
}`, resourceName, inlinePolicyName, policyName, resourceName, policyDocument)
	}

	return content, nil
}

// S3BucketResource represents an AWS S3 bucket
type S3BucketResource struct {
	BaseResource
	VersioningEnabled bool              `json:"versioning_enabled"`
	EncryptionEnabled bool              `json:"encryption_enabled"`
	KMSKeyID          string            `json:"kms_key_id,omitempty"`
	PublicAccessBlock *PublicAccessBlock `json:"public_access_block,omitempty"`
	LifecycleRules    []*LifecycleRule   `json:"lifecycle_rules,omitempty"`
}

type PublicAccessBlock struct {
	BlockPublicAcls       bool `json:"block_public_acls"`
	BlockPublicPolicy     bool `json:"block_public_policy"`
	IgnorePublicAcls      bool `json:"ignore_public_acls"`
	RestrictPublicBuckets bool `json:"restrict_public_buckets"`
}

type LifecycleRule struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Status  string `json:"status"`
	Expiration *Expiration `json:"expiration,omitempty"`
}

type Expiration struct {
	Days int `json:"days"`
}

func (r *S3BucketResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	// Build the main bucket resource
	content := fmt.Sprintf(`resource "aws_s3_bucket" "%s" {
  bucket = "%s"
  tags = {`,
		resourceName,
		r.Name)

	// Add tags
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
			// Quote tag keys that contain special characters
			quotedKey := k
			if strings.ContainsAny(k, ":-") {
				quotedKey = fmt.Sprintf(`"%s"`, k)
			}
			content += fmt.Sprintf(`
    %s = "%s"`, quotedKey, v)
		}
	}
	
	content += `
  }
}`

	mainResource := content

	// Add versioning
	if r.VersioningEnabled {
		mainResource += fmt.Sprintf(`
resource "aws_s3_bucket_versioning" "%s" {
  bucket = aws_s3_bucket.%s.id
  versioning_configuration {
    status = "Enabled"
  }
}`, strings.ReplaceAll(r.Name, "-", "_"),
		   strings.ReplaceAll(r.Name, "-", "_"))
	}

	// Add encryption
	if r.EncryptionEnabled {
		encryptionConfig := `  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"`
		if r.KMSKeyID != "" {
			encryptionConfig += fmt.Sprintf(`
        kms_master_key_id = "%s"`, r.KMSKeyID)
		}
		encryptionConfig += `
      }
    }
  }`

		mainResource += fmt.Sprintf(`
resource "aws_s3_bucket_server_side_encryption_configuration" "%s" {
  bucket = aws_s3_bucket.%s.id
%s
}`, strings.ReplaceAll(r.Name, "-", "_"),
		   strings.ReplaceAll(r.Name, "-", "_"),
		   encryptionConfig)
	}

	return mainResource, nil
}

// RDSInstanceResource represents an AWS RDS instance
type RDSInstanceResource struct {
	BaseResource
	Engine               string `json:"engine"`
	EngineVersion        string `json:"engine_version"`
	InstanceClass        string `json:"instance_class"`
	AllocatedStorage     int    `json:"allocated_storage"`
	StorageType          string `json:"storage_type"`
	StorageEncrypted     bool   `json:"storage_encrypted"`
	KMSKeyID             string `json:"kms_key_id,omitempty"`
	DBName               string `json:"db_name"`
	Username             string `json:"username"`
	Password             string `json:"password"`
	VpcSecurityGroupIDs  []string `json:"vpc_security_group_ids"`
	DBSubnetGroupName    string `json:"db_subnet_group_name"`
	BackupRetentionPeriod int   `json:"backup_retention_period"`
	MultiAZ              bool   `json:"multi_az"`
	PubliclyAccessible   bool   `json:"publicly_accessible"`
	SkipFinalSnapshot    bool   `json:"skip_final_snapshot"`
}

func (r *RDSInstanceResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	// Build the base template
	content := fmt.Sprintf(`resource "aws_db_instance" "%s" {
  identifier = "%s"
  engine     = "%s"
  engine_version = "%s"
  instance_class = "%s"
  allocated_storage = %d
  storage_type = "%s"
  storage_encrypted = %t`,
		resourceName,
		r.Name,
		r.Engine,
		r.EngineVersion,
		r.InstanceClass,
		r.AllocatedStorage,
		r.StorageType,
		r.StorageEncrypted)

	// Add KMS key if present
	if r.KMSKeyID != "" {
		content += fmt.Sprintf(`
  kms_key_id = "%s"`, r.KMSKeyID)
	}

	// Add remaining fields
	content += fmt.Sprintf(`
  db_name = "%s"
  username = "%s"
  password = var.rds_password_%s
  vpc_security_group_ids = [`,
		r.DBName,
		r.Username,
		resourceName)

	// Add security groups
	if len(r.VpcSecurityGroupIDs) > 0 {
		for i, sg := range r.VpcSecurityGroupIDs {
			if i > 0 {
				content += ", "
			}
			content += fmt.Sprintf(`"%s"`, sg)
		}
	}
	
	content += fmt.Sprintf(`]
  db_subnet_group_name = "%s"
  backup_retention_period = %d
  multi_az = %t
  publicly_accessible = %t
  skip_final_snapshot = %t
  tags = {`,
		r.DBSubnetGroupName,
		r.BackupRetentionPeriod,
		r.MultiAZ,
		r.PubliclyAccessible,
		r.SkipFinalSnapshot)

	// Add tags
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
			// Quote tag keys that contain special characters
			quotedKey := k
			if strings.ContainsAny(k, ":-") {
				quotedKey = fmt.Sprintf(`"%s"`, k)
			}
			content += fmt.Sprintf(`
    %s = "%s"`, quotedKey, v)
		}
	}
	
	content += `
  }
}`

	return content, nil
}

// LoadBalancerResource represents an AWS Application Load Balancer
type LoadBalancerResource struct {
	BaseResource
	Internal           bool     `json:"internal"`
	LoadBalancerType   string   `json:"load_balancer_type"`
	SecurityGroups     []string `json:"security_groups"`
	Subnets            []string `json:"subnets"`
	EnableDeletionProtection bool `json:"enable_deletion_protection"`
	IdleTimeout        int      `json:"idle_timeout"`
	TargetGroups       []string `json:"target_groups,omitempty"`
}

func (r *LoadBalancerResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	// Build the base template
	content := fmt.Sprintf(`resource "aws_lb" "%s" {
  name               = "%s"
  internal           = %t
  load_balancer_type = "%s"
  security_groups    = [`,
		resourceName,
		r.Name,
		r.Internal,
		r.LoadBalancerType)

	// Add security groups
	if len(r.SecurityGroups) > 0 {
		for i, sg := range r.SecurityGroups {
			if i > 0 {
				content += ", "
			}
			content += fmt.Sprintf(`"%s"`, sg)
		}
	}
	
	content += `]
  subnets            = [`

	// Add subnets
	if len(r.Subnets) > 0 {
		for i, subnet := range r.Subnets {
			if i > 0 {
				content += ", "
			}
			content += fmt.Sprintf(`"%s"`, subnet)
		}
	}
	
	content += fmt.Sprintf(`]
  enable_deletion_protection = %t
  idle_timeout       = %d
  tags = {`,
		r.EnableDeletionProtection,
		r.IdleTimeout)

	// Add tags
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
			// Quote tag keys that contain special characters
			quotedKey := k
			if strings.ContainsAny(k, ":-") {
				quotedKey = fmt.Sprintf(`"%s"`, k)
			}
			content += fmt.Sprintf(`
    %s = "%s"`, quotedKey, v)
		}
	}
	
	content += `
  }
}`

	return content, nil
}

// LambdaFunctionResource represents an AWS Lambda function
type LambdaFunctionResource struct {
	BaseResource
	Runtime            string            `json:"runtime"`
	Handler            string            `json:"handler"`
	Role               string            `json:"role"`
	Code               string            `json:"code"`
	Timeout            int               `json:"timeout"`
	MemorySize         int               `json:"memory_size"`
	Environment        map[string]string `json:"environment"`
	ReservedConcurrencyLimit int         `json:"reserved_concurrency_limit,omitempty"`
}

func (r *LambdaFunctionResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	// Build the base template
	content := fmt.Sprintf(`resource "aws_lambda_function" "%s" {
  filename         = "%s"
  function_name    = "%s"
  role            = "%s"
  handler         = "%s"
  runtime         = "%s"
  timeout         = %d
  memory_size     = %d`,
		resourceName,
		r.Code,
		r.Name,
		r.Role,
		r.Handler,
		r.Runtime,
		r.Timeout,
		r.MemorySize)

	// Add reserved concurrency if specified
	if r.ReservedConcurrencyLimit > 0 {
		content += fmt.Sprintf(`
  reserved_concurrent_executions = %d`, r.ReservedConcurrencyLimit)
	}

	// Add environment variables if present
	if len(r.Environment) > 0 {
		content += `
  environment {
    variables = {`
		for k, v := range r.Environment {
			content += fmt.Sprintf(`
      %s = "%s"`, k, v)
		}
		content += `
    }
  }`
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
			// Quote tag keys that contain special characters
			quotedKey := k
			if strings.ContainsAny(k, ":-") {
				quotedKey = fmt.Sprintf(`"%s"`, k)
			}
			content += fmt.Sprintf(`
    %s = "%s"`, quotedKey, v)
		}
	}
	content += `
  }
}`

	return content, nil
}

// GenericResource represents a generic AWS resource
type GenericResource struct {
	BaseResource
	ResourceType string                 `json:"resource_type"`
	Attributes   map[string]interface{} `json:"attributes"`
}

func (r *GenericResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`# Generic resource: %s
# Resource type: %s
# This is a placeholder for %s resource
# Manual configuration may be required

resource "%s" "%s" {
  # Add your configuration here
  # Example attributes:`, r.Name, r.ResourceType, r.ResourceType, r.ResourceType, resourceName)

	// Add some example attributes if available
	if len(r.Attributes) > 0 {
		for k, v := range r.Attributes {
			content += fmt.Sprintf(`
  # %s = %v`, k, v)
		}
	}

	content += `
}`

	return content, nil
} 