package aws

import (
	"fmt"
	"strings"
)

// SanitizeResourceName converts a resource name to a valid OpenTofu resource name
func SanitizeResourceName(name string) string {
	// Replace invalid characters with underscores
	invalidChars := []string{" ", "-", ".", "/", "\\", ":", "*", "?", "\"", "<", ">", "|", "!", "@", "#", "$", "%", "^", "&", "(", ")", "+", "=", "[", "]", "{", "}", "|", "\\", ";", "'", ",", "?"}
	result := name
	
	// Replace invalid characters
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

	// Remove any double underscores that might have been created
	result = strings.ReplaceAll(result, "__", "_")
	
	// Remove leading/trailing underscores
	result = strings.Trim(result, "_")

	return result
}

// SanitizeSecretsManagerName specifically handles SecretsManager names which have different requirements
func SanitizeSecretsManagerName(name string) string {
	// Replace invalid characters including forward slashes
	result := strings.ReplaceAll(name, "!", "_")
	result = strings.ReplaceAll(result, " ", "_")
	result = strings.ReplaceAll(result, "-", "_")
	result = strings.ReplaceAll(result, "/", "_")
	result = strings.ReplaceAll(result, "\\", "_")
	result = strings.ReplaceAll(result, ".", "_")
	result = strings.ReplaceAll(result, ":", "_")
	
	// Ensure it starts with a letter
	if len(result) > 0 && (result[0] < 'a' || result[0] > 'z') && (result[0] < 'A' || result[0] > 'Z') {
		result = "secret_" + result
	}

	// Limit length
	if len(result) > 63 {
		prefix := result[:50]
		suffix := result[len(result)-13:]
		result = prefix + "_" + suffix
	}

	// Remove any double underscores
	result = strings.ReplaceAll(result, "__", "_")
	
	// Remove leading/trailing underscores
	result = strings.Trim(result, "_")

	return result
}

// Resource interface defines methods that all AWS resources must implement
type Resource interface {
	GetType() string
	GetID() string
	GetName() string
	GetRegion() string
	GetTags() map[string]string
	GetARN() string
	GetTagsJSON() string
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

func (r *BaseResource) GetARN() string {
	// This is a default implementation
	// Individual resource types should override this if they have specific ARN logic
	return ""
}

// GetTagsJSON returns tags as a JSON string
func (r *BaseResource) GetTagsJSON() string {
	if len(r.Tags) == 0 {
		return "{}"
	}
	
	// Convert map to JSON string
	tagsJSON := "{"
	first := true
	for key, value := range r.Tags {
		if !first {
			tagsJSON += ","
		}
		tagsJSON += fmt.Sprintf(`"%s":"%s"`, key, value)
		first = false
	}
	tagsJSON += "}"
	return tagsJSON
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
		fmt.Sprintf(`<<EOF
%s
EOF`, r.AssumeRolePolicy))

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
}`, resourceName, inlinePolicyName, policyName, resourceName, fmt.Sprintf(`<<EOF
%s
EOF`, policyDocument))
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

	// Add encryption - temporarily disabled due to OpenTofu syntax issues
	// TODO: Fix encryption configuration generation
	/*
	if r.EncryptionEnabled {
		encryptionConfig := `    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"`
		if r.KMSKeyID != "" {
			encryptionConfig += fmt.Sprintf(`
        kms_master_key_id = "%s"`, r.KMSKeyID)
		}
		encryptionConfig += `
      }
    }`

		mainResource += fmt.Sprintf(`
resource "aws_s3_bucket_server_side_encryption_configuration" "%s" {
  bucket = aws_s3_bucket.%s.id
  server_side_encryption_configuration {
%s
  }
}`, strings.ReplaceAll(r.Name, "-", "_"),
		   strings.ReplaceAll(r.Name, "-", "_"),
		   encryptionConfig)
	}
	*/

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

// ASGResource represents an AWS Auto Scaling Group
type ASGResource struct {
	BaseResource
	MaxSize                 int32    `json:"max_size"`
	MinSize                 int32    `json:"min_size"`
	DesiredCapacity         int32    `json:"desired_capacity"`
	HealthCheckType         string   `json:"health_check_type"`
	HealthCheckGracePeriod  int32    `json:"health_check_grace_period"`
	VPCZoneIdentifier       []string `json:"vpc_zone_identifier,omitempty"`
	LaunchTemplate          *LaunchTemplateSpecification `json:"launch_template,omitempty"`
	MixedInstancesPolicy    *MixedInstancesPolicy       `json:"mixed_instances_policy,omitempty"`
	TargetGroupARNs         []string `json:"target_group_arns,omitempty"`
	LoadBalancerNames       []string `json:"load_balancer_names,omitempty"`
	ServiceLinkedRoleARN    string   `json:"service_linked_role_arn,omitempty"`
	MaxInstanceLifetime     int32    `json:"max_instance_lifetime,omitempty"`
	CapacityRebalance       bool     `json:"capacity_rebalance,omitempty"`
	ProtectFromScaleIn      bool     `json:"protect_from_scale_in,omitempty"`
}

type LaunchTemplateSpecification struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

type MixedInstancesPolicy struct {
	LaunchTemplate *LaunchTemplateSpecification `json:"launch_template,omitempty"`
	InstancesDistribution *InstancesDistribution `json:"instances_distribution,omitempty"`
}

type InstancesDistribution struct {
	OnDemandBaseCapacity                  int32   `json:"on_demand_base_capacity,omitempty"`
	OnDemandPercentageAboveBaseCapacity   int32   `json:"on_demand_percentage_above_base_capacity,omitempty"`
	SpotAllocationStrategy                string  `json:"spot_allocation_strategy,omitempty"`
	SpotInstancePools                     int32   `json:"spot_instance_pools,omitempty"`
	SpotMaxPrice                          string  `json:"spot_max_price,omitempty"`
}

func (r *ASGResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_autoscaling_group" "%s" {
  name                = "%s"
  max_size            = %d
  min_size            = %d
  desired_capacity    = %d
  health_check_type   = "%s"`,
		resourceName,
		r.Name,
		r.MaxSize,
		r.MinSize,
		r.DesiredCapacity,
		r.HealthCheckType)

	// Add health check grace period if specified
	if r.HealthCheckGracePeriod > 0 {
		content += fmt.Sprintf(`
  health_check_grace_period = %d`, r.HealthCheckGracePeriod)
	}

	// Add VPC zone identifier if specified
	if len(r.VPCZoneIdentifier) > 0 {
		content += fmt.Sprintf(`
  vpc_zone_identifier = [`)
		for i, subnet := range r.VPCZoneIdentifier {
			if i > 0 {
				content += ","
			}
			content += fmt.Sprintf(`"%s"`, subnet)
		}
		content += `]`
	}

	// Add launch template if specified
	if r.LaunchTemplate != nil {
		content += `
  launch_template {`
		if r.LaunchTemplate.ID != "" {
			content += fmt.Sprintf(`
    id = "%s"`, r.LaunchTemplate.ID)
		}
		if r.LaunchTemplate.Name != "" {
			content += fmt.Sprintf(`
    name = "%s"`, r.LaunchTemplate.Name)
		}
		if r.LaunchTemplate.Version != "" {
			content += fmt.Sprintf(`
    version = "%s"`, r.LaunchTemplate.Version)
		}
		content += `
  }`
	}

	// Add mixed instances policy if specified
	if r.MixedInstancesPolicy != nil {
		content += `
  mixed_instances_policy {`
		if r.MixedInstancesPolicy.LaunchTemplate != nil {
			content += `
    launch_template {`
			if r.MixedInstancesPolicy.LaunchTemplate.ID != "" {
				content += fmt.Sprintf(`
      id = "%s"`, r.MixedInstancesPolicy.LaunchTemplate.ID)
			}
			if r.MixedInstancesPolicy.LaunchTemplate.Name != "" {
				content += fmt.Sprintf(`
      name = "%s"`, r.MixedInstancesPolicy.LaunchTemplate.Name)
			}
			if r.MixedInstancesPolicy.LaunchTemplate.Version != "" {
				content += fmt.Sprintf(`
      version = "%s"`, r.MixedInstancesPolicy.LaunchTemplate.Version)
			}
			content += `
    }`
		}
		if r.MixedInstancesPolicy.InstancesDistribution != nil {
			content += `
    instances_distribution {`
			if r.MixedInstancesPolicy.InstancesDistribution.OnDemandBaseCapacity > 0 {
				content += fmt.Sprintf(`
      on_demand_base_capacity = %d`, r.MixedInstancesPolicy.InstancesDistribution.OnDemandBaseCapacity)
			}
			if r.MixedInstancesPolicy.InstancesDistribution.OnDemandPercentageAboveBaseCapacity > 0 {
				content += fmt.Sprintf(`
      on_demand_percentage_above_base_capacity = %d`, r.MixedInstancesPolicy.InstancesDistribution.OnDemandPercentageAboveBaseCapacity)
			}
			if r.MixedInstancesPolicy.InstancesDistribution.SpotAllocationStrategy != "" {
				content += fmt.Sprintf(`
      spot_allocation_strategy = "%s"`, r.MixedInstancesPolicy.InstancesDistribution.SpotAllocationStrategy)
			}
			if r.MixedInstancesPolicy.InstancesDistribution.SpotInstancePools > 0 {
				content += fmt.Sprintf(`
      spot_instance_pools = %d`, r.MixedInstancesPolicy.InstancesDistribution.SpotInstancePools)
			}
			if r.MixedInstancesPolicy.InstancesDistribution.SpotMaxPrice != "" {
				content += fmt.Sprintf(`
      spot_max_price = "%s"`, r.MixedInstancesPolicy.InstancesDistribution.SpotMaxPrice)
			}
			content += `
    }`
		}
		content += `
  }`
	}

	// Add target group ARNs if specified
	if len(r.TargetGroupARNs) > 0 {
		content += `
  target_group_arns = [`
		for i, arn := range r.TargetGroupARNs {
			if i > 0 {
				content += ","
			}
			content += fmt.Sprintf(`"%s"`, arn)
		}
		content += `]`
	}

	// Add load balancer names if specified
	if len(r.LoadBalancerNames) > 0 {
		content += `
  load_balancers = [`
		for i, name := range r.LoadBalancerNames {
			if i > 0 {
				content += ","
			}
			content += fmt.Sprintf(`"%s"`, name)
		}
		content += `]`
	}

	// Add service linked role ARN if specified
	if r.ServiceLinkedRoleARN != "" {
		content += fmt.Sprintf(`
  service_linked_role_arn = "%s"`, r.ServiceLinkedRoleARN)
	}

	// Add max instance lifetime if specified
	if r.MaxInstanceLifetime > 0 {
		content += fmt.Sprintf(`
  max_instance_lifetime = %d`, r.MaxInstanceLifetime)
	}

	// Add capacity rebalance if specified
	if r.CapacityRebalance {
		content += `
  capacity_rebalance = true`
	}

	// Add protect from scale in if specified
	if r.ProtectFromScaleIn {
		content += `
  protect_from_scale_in = true`
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

// ElastiCacheResource represents an AWS ElastiCache cluster
type ElastiCacheResource struct {
	BaseResource
	Engine               string `json:"engine"`
	NodeType             string `json:"node_type"`
	NumCacheNodes        int32  `json:"num_cache_nodes"`
	Port                 int32  `json:"port"`
	SubnetGroupName      string `json:"subnet_group_name"`
	SecurityGroupIDs     []string `json:"security_group_ids"`
	ParameterGroupName   string `json:"parameter_group_name"`
	EngineVersion        string `json:"engine_version"`
	MultiAZEnabled       bool   `json:"multi_az_enabled"`
	AutomaticFailoverEnabled bool `json:"automatic_failover_enabled"`
	AtRestEncryptionEnabled bool `json:"at_rest_encryption_enabled"`
	TransitEncryptionEnabled bool `json:"transit_encryption_enabled"`
	KMSKeyID             string `json:"kms_key_id,omitempty"`
	SnapshotRetentionLimit int32 `json:"snapshot_retention_limit"`
	SnapshotWindow       string `json:"snapshot_window"`
	MaintenanceWindow    string `json:"maintenance_window"`
	NotificationTopicARN string `json:"notification_topic_arn,omitempty"`
	LogDeliveryConfiguration []*LogDeliveryConfiguration `json:"log_delivery_configuration,omitempty"`
}

type LogDeliveryConfiguration struct {
	LogType                string `json:"log_type"`
	DestinationType        string `json:"destination_type"`
	DestinationDetails     *DestinationDetails `json:"destination_details"`
	LogFormat             string `json:"log_format"`
	Enabled               bool   `json:"enabled"`
}

type DestinationDetails struct {
	CloudWatchLogsDetails *CloudWatchLogsDetails `json:"cloud_watch_logs_details,omitempty"`
	KinesisFirehoseDetails *KinesisFirehoseDetails `json:"kinesis_firehose_details,omitempty"`
}

type CloudWatchLogsDetails struct {
	LogGroup string `json:"log_group"`
}

type KinesisFirehoseDetails struct {
	DeliveryStream string `json:"delivery_stream"`
}

func (r *ElastiCacheResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_elasticache_cluster" "%s" {
  cluster_id           = "%s"
  engine               = "%s"
  node_type            = "%s"
  num_cache_nodes      = %d
  port                 = %d`,
		resourceName,
		r.Name,
		r.Engine,
		r.NodeType,
		r.NumCacheNodes,
		r.Port)

	// Add subnet group if specified
	if r.SubnetGroupName != "" {
		content += fmt.Sprintf(`
  subnet_group_name    = "%s"`, r.SubnetGroupName)
	}

	// Add security groups if specified
	if len(r.SecurityGroupIDs) > 0 {
		content += fmt.Sprintf(`
  security_group_ids   = [`)
		for i, sg := range r.SecurityGroupIDs {
			if i > 0 {
				content += ","
			}
			content += fmt.Sprintf(`"%s"`, sg)
		}
		content += `]`
	}

	// Add parameter group if specified
	if r.ParameterGroupName != "" {
		content += fmt.Sprintf(`
  parameter_group_name = "%s"`, r.ParameterGroupName)
	}

	// Add engine version if specified
	if r.EngineVersion != "" {
		// Fix Redis version format - remove patch version for Redis v6+
		engineVersion := r.EngineVersion
		if r.Engine == "redis" && strings.HasPrefix(engineVersion, "7.") {
			// For Redis 7.x, use major.minor format
			parts := strings.Split(engineVersion, ".")
			if len(parts) >= 2 {
				engineVersion = fmt.Sprintf("%s.%s", parts[0], parts[1])
			}
		}
		content += fmt.Sprintf(`
  engine_version       = "%s"`, engineVersion)
	}

	// Add multi-AZ if enabled
	if r.MultiAZEnabled {
		content += `
  multi_az_enabled    = true`
	}

	// Add automatic failover if enabled
	if r.AutomaticFailoverEnabled {
		content += `
  automatic_failover_enabled = true`
	}

	// Add encryption settings
	if r.AtRestEncryptionEnabled {
		content += `
  at_rest_encryption_enabled = true`
	}

	if r.TransitEncryptionEnabled {
		content += `
  transit_encryption_enabled = true`
	}

	// Add KMS key if specified
	if r.KMSKeyID != "" {
		content += fmt.Sprintf(`
  kms_key_id          = "%s"`, r.KMSKeyID)
	}

	// Add snapshot settings if specified
	if r.SnapshotRetentionLimit > 0 {
		content += fmt.Sprintf(`
  snapshot_retention_limit = %d`, r.SnapshotRetentionLimit)
	}

	if r.SnapshotWindow != "" {
		content += fmt.Sprintf(`
  snapshot_window      = "%s"`, r.SnapshotWindow)
	}

	if r.MaintenanceWindow != "" {
		content += fmt.Sprintf(`
  maintenance_window   = "%s"`, r.MaintenanceWindow)
	}

	// Add notification topic if specified
	if r.NotificationTopicARN != "" {
		content += fmt.Sprintf(`
  notification_topic_arn = "%s"`, r.NotificationTopicARN)
	}

	// Add log delivery configuration if specified
	if len(r.LogDeliveryConfiguration) > 0 {
		content += `
  log_delivery_configuration {`
		for _, logConfig := range r.LogDeliveryConfiguration {
			content += fmt.Sprintf(`
    log_type = "%s"
    destination_type = "%s"
    log_format = "%s"
    enabled = %t`,
				logConfig.LogType,
				logConfig.DestinationType,
				logConfig.LogFormat,
				logConfig.Enabled)
			
			if logConfig.DestinationDetails != nil {
				if logConfig.DestinationDetails.CloudWatchLogsDetails != nil {
					content += fmt.Sprintf(`
    destination_details {
      cloud_watch_logs_details {
        log_group = "%s"
      }
    }`,
						logConfig.DestinationDetails.CloudWatchLogsDetails.LogGroup)
				} else if logConfig.DestinationDetails.KinesisFirehoseDetails != nil {
					content += fmt.Sprintf(`
    destination_details {
      kinesis_firehose_details {
        delivery_stream = "%s"
      }
    }`,
						logConfig.DestinationDetails.KinesisFirehoseDetails.DeliveryStream)
				}
			}
		}
		content += `
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

// DynamoDBResource represents an AWS DynamoDB table
type DynamoDBResource struct {
	BaseResource
	BillingMode          string `json:"billing_mode"`
	ReadCapacity         int64  `json:"read_capacity"`
	WriteCapacity        int64  `json:"write_capacity"`
	HashKey              string `json:"hash_key"`
	RangeKey             string `json:"range_key,omitempty"`
	StreamEnabled        bool   `json:"stream_enabled"`
	StreamViewType       string `json:"stream_view_type,omitempty"`
	ServerSideEncryption *ServerSideEncryption `json:"server_side_encryption,omitempty"`
	PointInTimeRecovery  *PointInTimeRecovery `json:"point_in_time_recovery,omitempty"`
	GlobalSecondaryIndexes []*GlobalSecondaryIndex `json:"global_secondary_indexes,omitempty"`
	LocalSecondaryIndexes []*LocalSecondaryIndex `json:"local_secondary_indexes,omitempty"`
	TTL                  *TTL `json:"ttl,omitempty"`
}

type ServerSideEncryption struct {
	Enabled     bool   `json:"enabled"`
	KMSKeyARN   string `json:"kms_key_arn,omitempty"`
}

type PointInTimeRecovery struct {
	Enabled bool `json:"enabled"`
}

type GlobalSecondaryIndex struct {
	Name               string `json:"name"`
	HashKey            string `json:"hash_key"`
	RangeKey           string `json:"range_key,omitempty"`
	WriteCapacity      int64  `json:"write_capacity"`
	ReadCapacity       int64  `json:"read_capacity"`
	ProjectionType     string `json:"projection_type"`
	NonKeyAttributes   []string `json:"non_key_attributes,omitempty"`
}

type LocalSecondaryIndex struct {
	Name               string `json:"name"`
	RangeKey           string `json:"range_key"`
	ProjectionType     string `json:"projection_type"`
	NonKeyAttributes   []string `json:"non_key_attributes,omitempty"`
}

type TTL struct {
	Enabled        bool   `json:"enabled"`
	AttributeName  string `json:"attribute_name"`
}

func (r *DynamoDBResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_dynamodb_table" "%s" {
  name           = "%s"
  billing_mode   = "%s"`,
		resourceName,
		r.Name,
		r.BillingMode)

	// Add capacity settings for PROVISIONED billing mode
	if r.BillingMode == "PROVISIONED" {
		content += fmt.Sprintf(`
  read_capacity  = %d
  write_capacity = %d`,
			r.ReadCapacity,
			r.WriteCapacity)
	}

	// Add hash key
	content += fmt.Sprintf(`
  hash_key       = "%s"`, r.HashKey)

	// Add range key if specified
	if r.RangeKey != "" {
		content += fmt.Sprintf(`
  range_key      = "%s"`, r.RangeKey)
	}

	// Add attribute definitions
	content += `
  attribute {
    name = "LockID"
    type = "S"
  }`

	// Add stream settings if enabled
	if r.StreamEnabled {
		content += fmt.Sprintf(`
  stream_enabled = true
  stream_view_type = "%s"`,
			r.StreamViewType)
	}

	// Add server-side encryption if specified
	if r.ServerSideEncryption != nil {
		content += `
  server_side_encryption {`
		if r.ServerSideEncryption.Enabled {
			content += `
    enabled = true`
		}
		if r.ServerSideEncryption.KMSKeyARN != "" {
			content += fmt.Sprintf(`
    kms_key_arn = "%s"`,
				r.ServerSideEncryption.KMSKeyARN)
		}
		content += `
  }`
	}

	// Add point-in-time recovery if specified
	if r.PointInTimeRecovery != nil {
		content += fmt.Sprintf(`
  point_in_time_recovery {
    enabled = %t
  }`,
			r.PointInTimeRecovery.Enabled)
	}

	// Add global secondary indexes if specified
	if len(r.GlobalSecondaryIndexes) > 0 {
		content += `
  global_secondary_index {`
		for _, gsi := range r.GlobalSecondaryIndexes {
			content += fmt.Sprintf(`
    name = "%s"
    hash_key = "%s"
    projection_type = "%s"`,
				gsi.Name,
				gsi.HashKey,
				gsi.ProjectionType)
			
			if gsi.RangeKey != "" {
				content += fmt.Sprintf(`
    range_key = "%s"`,
					gsi.RangeKey)
			}
			
			if gsi.WriteCapacity > 0 {
				content += fmt.Sprintf(`
    write_capacity = %d`,
					gsi.WriteCapacity)
			}
			
			if gsi.ReadCapacity > 0 {
				content += fmt.Sprintf(`
    read_capacity = %d`,
					gsi.ReadCapacity)
			}
			
			if len(gsi.NonKeyAttributes) > 0 {
				content += `
    non_key_attributes = [`
				for i, attr := range gsi.NonKeyAttributes {
					if i > 0 {
						content += ","
					}
					content += fmt.Sprintf(`"%s"`, attr)
				}
				content += `]`
			}
		}
		content += `
  }`
	}

	// Add local secondary indexes if specified
	if len(r.LocalSecondaryIndexes) > 0 {
		content += `
  local_secondary_index {`
		for _, lsi := range r.LocalSecondaryIndexes {
			content += fmt.Sprintf(`
    name = "%s"
    range_key = "%s"
    projection_type = "%s"`,
				lsi.Name,
				lsi.RangeKey,
				lsi.ProjectionType)
			
			if len(lsi.NonKeyAttributes) > 0 {
				content += `
    non_key_attributes = [`
				for i, attr := range lsi.NonKeyAttributes {
					if i > 0 {
						content += ","
					}
					content += fmt.Sprintf(`"%s"`, attr)
				}
				content += `]`
			}
		}
		content += `
  }`
	}

	// Add TTL if specified
	if r.TTL != nil {
		content += fmt.Sprintf(`
  ttl {
    enabled = %t
    attribute_name = "%s"
  }`,
			r.TTL.Enabled,
			r.TTL.AttributeName)
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
			content += fmt.Sprintf(`
    %s = "%s"`, k, v)
		}
	}
	content += `
  }
}`

	return content, nil
}

// CloudFrontResource represents an AWS CloudFront distribution
type CloudFrontResource struct {
	BaseResource
	Enabled              bool   `json:"enabled"`
	PriceClass           string `json:"price_class"`
	RetainOnDelete       bool   `json:"retain_on_delete"`
	WaitForDeployment    bool   `json:"wait_for_deployment"`
	IsIPV6Enabled        bool   `json:"is_ipv6_enabled"`
	DefaultRootObject    string `json:"default_root_object,omitempty"`
	WebACLID             string `json:"web_acl_id,omitempty"`
	HttpVersion          string `json:"http_version"`
	Aliases              []string `json:"aliases,omitempty"`
	Origins              []*Origin `json:"origins"`
	DefaultCacheBehavior *CacheBehavior `json:"default_cache_behavior"`
	OrderedCacheBehaviors []*CacheBehavior `json:"ordered_cache_behaviors,omitempty"`
	CustomErrorResponses []*CustomErrorResponse `json:"custom_error_responses,omitempty"`
	ViewerCertificate    *ViewerCertificate `json:"viewer_certificate"`
	Restrictions         *Restrictions `json:"restrictions,omitempty"`
}

type Origin struct {
	DomainName              string `json:"domain_name"`
	OriginID                string `json:"origin_id"`
	OriginPath              string `json:"origin_path,omitempty"`
	CustomOriginConfig      *CustomOriginConfig `json:"custom_origin_config,omitempty"`
	S3OriginConfig          *S3OriginConfig `json:"s3_origin_config,omitempty"`
	CustomHeaders           []*CustomHeader `json:"custom_headers,omitempty"`
	ConnectionAttempts      int32  `json:"connection_attempts"`
	ConnectionTimeout       int32  `json:"connection_timeout"`
	OriginShield           *OriginShield `json:"origin_shield,omitempty"`
}

type CustomOriginConfig struct {
	HTTPPort                int32  `json:"http_port"`
	HTTPSPort               int32  `json:"https_port"`
	OriginProtocolPolicy    string `json:"origin_protocol_policy"`
	OriginSSLProtocols      []string `json:"origin_ssl_protocols"`
	OriginReadTimeout       int32  `json:"origin_read_timeout"`
	OriginKeepaliveTimeout  int32  `json:"origin_keepalive_timeout"`
}

type S3OriginConfig struct {
	OriginAccessIdentity string `json:"origin_access_identity"`
}

type CustomHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type OriginShield struct {
	Enabled              bool   `json:"enabled"`
	OriginShieldRegion   string `json:"origin_shield_region"`
}

type CacheBehavior struct {
	AllowedMethods            []string `json:"allowed_methods"`
	CachedMethods             []string `json:"cached_methods"`
	TargetOriginID            string `json:"target_origin_id"`
	ViewerProtocolPolicy     string `json:"viewer_protocol_policy"`
	CachePolicyID             string `json:"cache_policy_id,omitempty"`
	OriginRequestPolicyID     string `json:"origin_request_policy_id,omitempty"`
	ResponseHeadersPolicyID   string `json:"response_headers_policy_id,omitempty"`
	Compress                  bool   `json:"compress"`
	DefaultTTL                int64  `json:"default_ttl"`
	MaxTTL                    int64  `json:"max_ttl"`
	MinTTL                    int64  `json:"min_ttl"`
	PathPattern               string `json:"path_pattern,omitempty"`
	TrustedSigners            []string `json:"trusted_signers,omitempty"`
	TrustedKeyGroups          []string `json:"trusted_key_groups,omitempty"`
	ForwardedValues           *ForwardedValues `json:"forwarded_values,omitempty"`
	LambdaFunctionAssociations []*LambdaFunctionAssociation `json:"lambda_function_associations,omitempty"`
	FunctionAssociations      []*FunctionAssociation `json:"function_associations,omitempty"`
}

type ForwardedValues struct {
	QueryString bool `json:"query_string"`
	Cookies     *Cookies `json:"cookies"`
	Headers     []string `json:"headers"`
	QueryStringCacheKeys []string `json:"query_string_cache_keys"`
}

type Cookies struct {
	Forward          string `json:"forward"`
	WhitelistedNames []string `json:"whitelisted_names,omitempty"`
}

type LambdaFunctionAssociation struct {
	EventType   string `json:"event_type"`
	LambdaARN   string `json:"lambda_arn"`
	IncludeBody bool   `json:"include_body"`
}

type FunctionAssociation struct {
	EventType   string `json:"event_type"`
	FunctionARN string `json:"function_arn"`
}

type CustomErrorResponse struct {
	ErrorCode            int32  `json:"error_code"`
	ResponseCode         string `json:"response_code,omitempty"`
	ResponsePagePath     string `json:"response_page_path,omitempty"`
	ErrorCachingMinTTL   int64  `json:"error_caching_min_ttl"`
}

type ViewerCertificate struct {
	ACMCertificateARN            string `json:"acm_certificate_arn,omitempty"`
	CloudFrontDefaultCertificate bool   `json:"cloudfront_default_certificate"`
	MinimumProtocolVersion       string `json:"minimum_protocol_version"`
	SSLSupportMethod             string `json:"ssl_support_method"`
}

type Restrictions struct {
	GeoRestriction *GeoRestriction `json:"geo_restriction"`
}

type GeoRestriction struct {
	RestrictionType string   `json:"restriction_type"`
	Locations       []string `json:"locations,omitempty"`
}

func (r *CloudFrontResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_cloudfront_distribution" "%s" {
  enabled              = %t
  price_class          = "%s"
  retain_on_delete     = %t
  wait_for_deployment  = %t
  is_ipv6_enabled      = %t
  http_version         = "%s"`,
		resourceName,
		r.Enabled,
		r.PriceClass,
		r.RetainOnDelete,
		r.WaitForDeployment,
		r.IsIPV6Enabled,
		r.HttpVersion)

	// Add default root object if specified
	if r.DefaultRootObject != "" {
		content += fmt.Sprintf(`
  default_root_object = "%s"`,
			r.DefaultRootObject)
	}

	// Add web ACL if specified
	if r.WebACLID != "" {
		content += fmt.Sprintf(`
  web_acl_id = "%s"`,
			r.WebACLID)
	}

	// Add aliases if specified
	if len(r.Aliases) > 0 {
		content += `
  aliases = [`
		for i, alias := range r.Aliases {
			if i > 0 {
				content += ","
			}
			content += fmt.Sprintf(`"%s"`, alias)
		}
		content += `]`
	}

	// Add origins
	content += `
  origin {`
	for _, origin := range r.Origins {
		content += fmt.Sprintf(`
    domain_name = "%s"
    origin_id   = "%s"`,
			origin.DomainName,
			origin.OriginID)
		
		if origin.OriginPath != "" {
			content += fmt.Sprintf(`
    origin_path = "%s"`,
				origin.OriginPath)
		}
		
		if origin.ConnectionAttempts > 0 {
			content += fmt.Sprintf(`
    connection_attempts = %d`,
				origin.ConnectionAttempts)
		}
		
		if origin.ConnectionTimeout > 0 {
			content += fmt.Sprintf(`
    connection_timeout = %d`,
				origin.ConnectionTimeout)
		}
		
		// Add custom origin config if specified
		if origin.CustomOriginConfig != nil {
			content += `
    custom_origin_config {`
			content += fmt.Sprintf(`
      http_port = %d
      https_port = %d
      origin_protocol_policy = "%s"`,
				origin.CustomOriginConfig.HTTPPort,
				origin.CustomOriginConfig.HTTPSPort,
				origin.CustomOriginConfig.OriginProtocolPolicy)
			
			if len(origin.CustomOriginConfig.OriginSSLProtocols) > 0 {
				content += `
      origin_ssl_protocols = [`
				for i, protocol := range origin.CustomOriginConfig.OriginSSLProtocols {
					if i > 0 {
						content += ","
					}
					content += fmt.Sprintf(`"%s"`, protocol)
				}
				content += `]`
			}
			
			if origin.CustomOriginConfig.OriginReadTimeout > 0 {
				content += fmt.Sprintf(`
      origin_read_timeout = %d`,
					origin.CustomOriginConfig.OriginReadTimeout)
			}
			
			if origin.CustomOriginConfig.OriginKeepaliveTimeout > 0 {
				content += fmt.Sprintf(`
      origin_keepalive_timeout = %d`,
					origin.CustomOriginConfig.OriginKeepaliveTimeout)
			}
			
			content += `
    }`
		}
		
		// Add S3 origin config if specified
		if origin.S3OriginConfig != nil {
			content += fmt.Sprintf(`
    s3_origin_config {
      origin_access_identity = "%s"
    }`,
				origin.S3OriginConfig.OriginAccessIdentity)
		}
		
		// Add custom headers if specified
		if len(origin.CustomHeaders) > 0 {
			content += `
    custom_header {`
			for _, header := range origin.CustomHeaders {
				content += fmt.Sprintf(`
      name  = "%s"
      value = "%s"`,
					header.Name,
					header.Value)
			}
			content += `
    }`
		}
		
		// Add origin shield if specified
		if origin.OriginShield != nil {
			content += fmt.Sprintf(`
    origin_shield {
      enabled = %t
      origin_shield_region = "%s"
    }`,
				origin.OriginShield.Enabled,
				origin.OriginShield.OriginShieldRegion)
		}
	}
	content += `
  }`

	// Add default cache behavior
	if r.DefaultCacheBehavior != nil {
		content += `
  default_cache_behavior {`
		content += `
    allowed_methods  = [`
		for i, method := range r.DefaultCacheBehavior.AllowedMethods {
			if i > 0 {
				content += ","
			}
			content += fmt.Sprintf(`"%s"`, method)
		}
		content += `]
    cached_methods   = [`
		for i, method := range r.DefaultCacheBehavior.CachedMethods {
			if i > 0 {
				content += ","
			}
			content += fmt.Sprintf(`"%s"`, method)
		}
		content += fmt.Sprintf(`]
    target_origin_id = "%s"
    viewer_protocol_policy = "%s"`,
			r.DefaultCacheBehavior.TargetOriginID,
			r.DefaultCacheBehavior.ViewerProtocolPolicy)
		
		if r.DefaultCacheBehavior.CachePolicyID != "" {
			content += fmt.Sprintf(`
    cache_policy_id = "%s"`,
				r.DefaultCacheBehavior.CachePolicyID)
		}
		
		if r.DefaultCacheBehavior.OriginRequestPolicyID != "" {
			content += fmt.Sprintf(`
    origin_request_policy_id = "%s"`,
				r.DefaultCacheBehavior.OriginRequestPolicyID)
		}
		
		if r.DefaultCacheBehavior.ResponseHeadersPolicyID != "" {
			content += fmt.Sprintf(`
    response_headers_policy_id = "%s"`,
				r.DefaultCacheBehavior.ResponseHeadersPolicyID)
		}
		
		if r.DefaultCacheBehavior.Compress {
			content += `
    compress = true`
		}
		
		if r.DefaultCacheBehavior.DefaultTTL > 0 {
			content += fmt.Sprintf(`
    default_ttl = %d`,
				r.DefaultCacheBehavior.DefaultTTL)
		}
		
		if r.DefaultCacheBehavior.MaxTTL > 0 {
			content += fmt.Sprintf(`
    max_ttl = %d`,
				r.DefaultCacheBehavior.MaxTTL)
		}
		
		if r.DefaultCacheBehavior.MinTTL > 0 {
			content += fmt.Sprintf(`
    min_ttl = %d`,
				r.DefaultCacheBehavior.MinTTL)
		}
		
		content += `
  }`
	}

	// Add viewer certificate
	if r.ViewerCertificate != nil {
		content += `
  viewer_certificate {`
		if r.ViewerCertificate.ACMCertificateARN != "" {
			content += fmt.Sprintf(`
    acm_certificate_arn = "%s"`,
				r.ViewerCertificate.ACMCertificateARN)
		}
		
		if r.ViewerCertificate.CloudFrontDefaultCertificate {
			content += `
    cloudfront_default_certificate = true`
		}
		
		if r.ViewerCertificate.MinimumProtocolVersion != "" {
			content += fmt.Sprintf(`
    minimum_protocol_version = "%s"`,
				r.ViewerCertificate.MinimumProtocolVersion)
		}
		
		if r.ViewerCertificate.SSLSupportMethod != "" {
			content += fmt.Sprintf(`
    ssl_support_method = "%s"`,
				r.ViewerCertificate.SSLSupportMethod)
		}
		
		content += `
  }`
	}

	// Add restrictions if specified
	if r.Restrictions != nil && r.Restrictions.GeoRestriction != nil {
		content += `
  restrictions {`
		content += fmt.Sprintf(`
    geo_restriction {
      restriction_type = "%s"`,
			r.Restrictions.GeoRestriction.RestrictionType)
		
		if len(r.Restrictions.GeoRestriction.Locations) > 0 {
			content += `
      locations = [`
			for i, location := range r.Restrictions.GeoRestriction.Locations {
				if i > 0 {
					content += ","
				}
				content += fmt.Sprintf(`"%s"`, location)
			}
			content += `]`
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

// RedshiftResource represents an AWS Redshift cluster
type RedshiftResource struct {
	BaseResource
	ClusterIdentifier     string `json:"cluster_identifier"`
	DatabaseName          string `json:"database_name"`
	MasterUsername        string `json:"master_username"`
	MasterPassword        string `json:"master_password"`
	NodeType              string `json:"node_type"`
	ClusterType           string `json:"cluster_type"`
	NumberOfNodes         int32  `json:"number_of_nodes"`
	AutomatedSnapshotRetentionPeriod int32 `json:"automated_snapshot_retention_period"`
	PreferredMaintenanceWindow string `json:"preferred_maintenance_window"`
	ClusterSubnetGroupName string `json:"cluster_subnet_group_name"`
	VpcSecurityGroupIds   []string `json:"vpc_security_group_ids"`
	Encrypted             bool   `json:"encrypted"`
	KMSKeyID              string `json:"kms_key_id,omitempty"`
	EnhancedVpcRouting    bool   `json:"enhanced_vpc_routing"`
	PubliclyAccessible    bool   `json:"publicly_accessible"`
	SkipFinalSnapshot     bool   `json:"skip_final_snapshot"`
	FinalSnapshotIdentifier string `json:"final_snapshot_identifier,omitempty"`
	SnapshotIdentifier    string `json:"snapshot_identifier,omitempty"`
	OwnerAccount          string `json:"owner_account,omitempty"`
	Port                  int32  `json:"port"`
	AvailabilityZone      string `json:"availability_zone,omitempty"`
	AllowVersionUpgrade   bool   `json:"allow_version_upgrade"`
	ClusterVersion        string `json:"cluster_version,omitempty"`
	ClusterParameterGroupName string `json:"cluster_parameter_group_name,omitempty"`
	Logging               *Logging `json:"logging,omitempty"`
}

type Logging struct {
	EnableLogging bool   `json:"enable_logging"`
	BucketName    string `json:"bucket_name,omitempty"`
	S3KeyPrefix   string `json:"s3_key_prefix,omitempty"`
}

func (r *RedshiftResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_redshift_cluster" "%s" {
  cluster_identifier = "%s"
  database_name      = "%s"
  master_username    = "%s"
  master_password    = var.redshift_password
  node_type          = "%s"
  cluster_type       = "%s"`,
		resourceName,
		r.ClusterIdentifier,
		r.DatabaseName,
		r.MasterUsername,
		r.NodeType,
		r.ClusterType)

	// Add number of nodes for multi-node clusters
	if r.ClusterType == "multi-node" && r.NumberOfNodes > 0 {
		content += fmt.Sprintf(`
  number_of_nodes    = %d`,
			r.NumberOfNodes)
	}

	// Add automated snapshot retention if specified
	if r.AutomatedSnapshotRetentionPeriod > 0 {
		content += fmt.Sprintf(`
  automated_snapshot_retention_period = %d`,
			r.AutomatedSnapshotRetentionPeriod)
	}

	// Add maintenance window if specified
	if r.PreferredMaintenanceWindow != "" {
		content += fmt.Sprintf(`
  preferred_maintenance_window = "%s"`,
			r.PreferredMaintenanceWindow)
	}

	// Add subnet group if specified
	if r.ClusterSubnetGroupName != "" {
		content += fmt.Sprintf(`
  cluster_subnet_group_name = "%s"`,
			r.ClusterSubnetGroupName)
	}

	// Add security groups if specified
	if len(r.VpcSecurityGroupIds) > 0 {
		content += `
  vpc_security_group_ids = [`
		for i, sg := range r.VpcSecurityGroupIds {
			if i > 0 {
				content += ","
			}
			content += fmt.Sprintf(`"%s"`, sg)
		}
		content += `]`
	}

	// Add encryption settings
	if r.Encrypted {
		content += `
  encrypted = true`
		if r.KMSKeyID != "" {
			content += fmt.Sprintf(`
  kms_key_id = "%s"`,
				r.KMSKeyID)
		}
	}

	// Add enhanced VPC routing if enabled
	if r.EnhancedVpcRouting {
		content += `
  enhanced_vpc_routing = true`
	}

	// Add publicly accessible if enabled
	if r.PubliclyAccessible {
		content += `
  publicly_accessible = true`
	}

	// Add skip final snapshot if enabled
	if r.SkipFinalSnapshot {
		content += `
  skip_final_snapshot = true`
	}

	// Add final snapshot identifier if specified
	if r.FinalSnapshotIdentifier != "" {
		content += fmt.Sprintf(`
  final_snapshot_identifier = "%s"`,
			r.FinalSnapshotIdentifier)
	}

	// Add snapshot identifier if specified
	if r.SnapshotIdentifier != "" {
		content += fmt.Sprintf(`
  snapshot_identifier = "%s"`,
			r.SnapshotIdentifier)
	}

	// Add owner account if specified
	if r.OwnerAccount != "" {
		content += fmt.Sprintf(`
  owner_account = "%s"`,
			r.OwnerAccount)
	}

	// Add port if specified
	if r.Port > 0 {
		content += fmt.Sprintf(`
  port = %d`,
			r.Port)
	}

	// Add availability zone if specified
	if r.AvailabilityZone != "" {
		content += fmt.Sprintf(`
  availability_zone = "%s"`,
			r.AvailabilityZone)
	}

	// Add allow version upgrade if specified
	if r.AllowVersionUpgrade {
		content += `
  allow_version_upgrade = true`
	}

	// Add cluster version if specified
	if r.ClusterVersion != "" {
		content += fmt.Sprintf(`
  cluster_version = "%s"`,
			r.ClusterVersion)
	}

	// Add parameter group if specified
	if r.ClusterParameterGroupName != "" {
		content += fmt.Sprintf(`
  cluster_parameter_group_name = "%s"`,
			r.ClusterParameterGroupName)
	}

	// Add logging if specified
	if r.Logging != nil {
		content += `
  logging {`
		if r.Logging.EnableLogging {
			content += `
    enable_logging = true`
		}
		if r.Logging.BucketName != "" {
			content += fmt.Sprintf(`
    bucket_name = "%s"`,
				r.Logging.BucketName)
		}
		if r.Logging.S3KeyPrefix != "" {
			content += fmt.Sprintf(`
    s3_key_prefix = "%s"`,
				r.Logging.S3KeyPrefix)
		}
		content += `
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

// Route53Resource represents an AWS Route53 hosted zone
type Route53Resource struct {
	BaseResource
	ZoneID                string `json:"zone_id"`
	Name                  string `json:"name"`
	Comment               string `json:"comment,omitempty"`
	PrivateZone           bool   `json:"private_zone"`
	VPCID                 string `json:"vpc_id,omitempty"`
	VPCRegion             string `json:"vpc_region,omitempty"`
	ForceDestroy          bool   `json:"force_destroy"`
	DelegationSetID       string `json:"delegation_set_id,omitempty"`
	Records               []*Route53Record `json:"records,omitempty"`
}

type Route53Record struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	TTL     int64  `json:"ttl"`
	Records []string `json:"records,omitempty"`
	Alias   *Alias `json:"alias,omitempty"`
	Failover *Failover `json:"failover,omitempty"`
	Geolocation *Geolocation `json:"geolocation,omitempty"`
	Latency *Latency `json:"latency,omitempty"`
	Weighted *Weighted `json:"weighted,omitempty"`
	Multivalue *Multivalue `json:"multivalue,omitempty"`
}

type Alias struct {
	Name                   string `json:"name"`
	ZoneID                 string `json:"zone_id"`
	EvaluateTargetHealth   bool   `json:"evaluate_target_health"`
}

type Failover struct {
	Type string `json:"type"`
}

type Geolocation struct {
	Continent   string `json:"continent,omitempty"`
	Country     string `json:"country,omitempty"`
	Subdivision string `json:"subdivision,omitempty"`
}

type Latency struct {
	Region string `json:"region"`
}

type Weighted struct {
	Weight int64 `json:"weight"`
}

type Multivalue struct {
	Weight int64 `json:"weight"`
}

func (r *Route53Resource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_route53_zone" "%s" {
  name = "%s"`,
		resourceName,
		r.Name)

	// Add comment if specified
	if r.Comment != "" {
		content += fmt.Sprintf(`
  comment = "%s"`,
			r.Comment)
	}

	// Add private zone settings if specified
	if r.PrivateZone {
		content += `
  private_zone = true`
		if r.VPCID != "" {
			content += fmt.Sprintf(`
  vpc {
    vpc_id = "%s"`,
				r.VPCID)
			if r.VPCRegion != "" {
				content += fmt.Sprintf(`
    vpc_region = "%s"`,
					r.VPCRegion)
			}
			content += `
  }`
		}
	}

	// Add force destroy if specified
	if r.ForceDestroy {
		content += `
  force_destroy = true`
	}

	// Add delegation set if specified
	if r.DelegationSetID != "" {
		content += fmt.Sprintf(`
  delegation_set_id = "%s"`,
			r.DelegationSetID)
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

	// Add records if specified
	if len(r.Records) > 0 {
		for i, record := range r.Records {
			recordName := SanitizeResourceName(fmt.Sprintf("%s_record_%d", resourceName, i))
			content += fmt.Sprintf(`

resource "aws_route53_record" "%s" {
  zone_id = aws_route53_zone.%s.zone_id
  name    = "%s"
  type    = "%s"
  ttl     = %d`,
				recordName,
				resourceName,
				record.Name,
				record.Type,
				record.TTL)

			// Add records if specified
			if len(record.Records) > 0 {
				content += `
  records = [`
				for j, rec := range record.Records {
					if j > 0 {
						content += ","
					}
					content += fmt.Sprintf(`"%s"`, rec)
				}
				content += `]`
			}

			// Add alias if specified
			if record.Alias != nil {
				content += fmt.Sprintf(`
  alias {
    name                   = "%s"
    zone_id                = "%s"
    evaluate_target_health = %t
  }`,
					record.Alias.Name,
					record.Alias.ZoneID,
					record.Alias.EvaluateTargetHealth)
			}

			// Add failover if specified
			if record.Failover != nil {
				content += fmt.Sprintf(`
  failover_routing_policy {
    type = "%s"
  }`,
					record.Failover.Type)
			}

			// Add geolocation if specified
			if record.Geolocation != nil {
				content += `
  geolocation_routing_policy {`
				if record.Geolocation.Continent != "" {
					content += fmt.Sprintf(`
    continent = "%s"`,
						record.Geolocation.Continent)
				}
				if record.Geolocation.Country != "" {
					content += fmt.Sprintf(`
    country = "%s"`,
						record.Geolocation.Country)
				}
				if record.Geolocation.Subdivision != "" {
					content += fmt.Sprintf(`
    subdivision = "%s"`,
						record.Geolocation.Subdivision)
				}
				content += `
  }`
			}

			// Add latency if specified
			if record.Latency != nil {
				content += fmt.Sprintf(`
  latency_routing_policy {
    region = "%s"
  }`,
					record.Latency.Region)
			}

			// Add weighted if specified
			if record.Weighted != nil {
				content += fmt.Sprintf(`
  weighted_routing_policy {
    weight = %d
  }`,
					record.Weighted.Weight)
			}

			// Add multivalue if specified
			if record.Multivalue != nil {
				content += fmt.Sprintf(`
  multivalue_answer_routing_policy {
    weight = %d
  }`,
					record.Multivalue.Weight)
			}

			content += `
}`
		}
	}

	return content, nil
}

// ECSResource represents an AWS ECS service
type ECSResource struct {
	BaseResource
	ClusterARN         string `json:"cluster_arn"`
	TaskDefinitionARN  string `json:"task_definition_arn"`
	DesiredCount       int32  `json:"desired_count"`
	LaunchType         string `json:"launch_type"`
	PlatformVersion    string `json:"platform_version"`
	NetworkConfiguration *NetworkConfiguration `json:"network_configuration,omitempty"`
	LoadBalancers      []*LoadBalancer `json:"load_balancers,omitempty"`
	ServiceRegistries  []*ServiceRegistry `json:"service_registries,omitempty"`
	DeploymentConfiguration *DeploymentConfiguration `json:"deployment_configuration,omitempty"`
	PlacementConstraints []*PlacementConstraint `json:"placement_constraints,omitempty"`
	PlacementStrategy  []*PlacementStrategy `json:"placement_strategy,omitempty"`
	HealthCheckGracePeriodSeconds int32 `json:"health_check_grace_period_seconds"`
	ForceNewDeployment bool `json:"force_new_deployment"`
	WaitForSteadyState bool `json:"wait_for_steady_state"`
}

type NetworkConfiguration struct {
	Subnets          []string `json:"subnets"`
	SecurityGroups   []string `json:"security_groups"`
	AssignPublicIP   bool     `json:"assign_public_ip"`
}

type LoadBalancer struct {
	TargetGroupARN string `json:"target_group_arn"`
	ContainerName  string `json:"container_name"`
	ContainerPort  int32  `json:"container_port"`
}

type ServiceRegistry struct {
	RegistryARN string `json:"registry_arn"`
	Port        int32  `json:"port"`
	ContainerName string `json:"container_name"`
	ContainerPort int32  `json:"container_port"`
}

type DeploymentConfiguration struct {
	MaximumPercent         int32 `json:"maximum_percent"`
	MinimumHealthyPercent  int32 `json:"minimum_healthy_percent"`
	DeploymentCircuitBreaker *DeploymentCircuitBreaker `json:"deployment_circuit_breaker,omitempty"`
}

type DeploymentCircuitBreaker struct {
	Enable   bool `json:"enable"`
	Rollback bool `json:"rollback"`
}

type PlacementConstraint struct {
	Type       string `json:"type"`
	Expression string `json:"expression"`
}

type PlacementStrategy struct {
	Type  string `json:"type"`
	Field string `json:"field"`
}

func (r *ECSResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_ecs_service" "%s" {
  name            = "%s"
  cluster         = "%s"
  task_definition = "%s"
  desired_count   = %d
  launch_type     = "%s"`,
		resourceName,
		r.Name,
		r.ClusterARN,
		r.TaskDefinitionARN,
		r.DesiredCount,
		r.LaunchType)

	// Add platform version if specified
	if r.PlatformVersion != "" {
		content += fmt.Sprintf(`
  platform_version = "%s"`,
			r.PlatformVersion)
	}

	// Add health check grace period if specified
	if r.HealthCheckGracePeriodSeconds > 0 {
		content += fmt.Sprintf(`
  health_check_grace_period_seconds = %d`,
			r.HealthCheckGracePeriodSeconds)
	}

	// Add force new deployment if enabled
	if r.ForceNewDeployment {
		content += `
  force_new_deployment = true`
	}

	// Add wait for steady state if enabled
	if r.WaitForSteadyState {
		content += `
  wait_for_steady_state = true`
	}

	// Add network configuration if specified
	if r.NetworkConfiguration != nil {
		content += `
  network_configuration {`
		if len(r.NetworkConfiguration.Subnets) > 0 {
			content += `
    subnets = [`
			for i, subnet := range r.NetworkConfiguration.Subnets {
				if i > 0 {
					content += ","
				}
				content += fmt.Sprintf(`"%s"`, subnet)
			}
			content += `]`
		}
		if len(r.NetworkConfiguration.SecurityGroups) > 0 {
			content += `
    security_groups = [`
			for i, sg := range r.NetworkConfiguration.SecurityGroups {
				if i > 0 {
					content += ","
				}
				content += fmt.Sprintf(`"%s"`, sg)
			}
			content += `]`
		}
		if r.NetworkConfiguration.AssignPublicIP {
			content += `
    assign_public_ip = true`
		}
		content += `
  }`
	}

	// Add load balancers if specified
	if len(r.LoadBalancers) > 0 {
		content += `
  load_balancer {`
		for _, lb := range r.LoadBalancers {
			content += fmt.Sprintf(`
    target_group_arn = "%s"
    container_name   = "%s"
    container_port   = %d`,
				lb.TargetGroupARN,
				lb.ContainerName,
				lb.ContainerPort)
		}
		content += `
  }`
	}

	// Add service registries if specified
	if len(r.ServiceRegistries) > 0 {
		content += `
  service_registries {`
		for _, sr := range r.ServiceRegistries {
			content += fmt.Sprintf(`
    registry_arn   = "%s"
    port           = %d
    container_name = "%s"
    container_port = %d`,
				sr.RegistryARN,
				sr.Port,
				sr.ContainerName,
				sr.ContainerPort)
		}
		content += `
  }`
	}

	// Add deployment configuration if specified
	if r.DeploymentConfiguration != nil {
		content += `
  deployment_configuration {`
		if r.DeploymentConfiguration.MaximumPercent > 0 {
			content += fmt.Sprintf(`
    maximum_percent = %d`,
				r.DeploymentConfiguration.MaximumPercent)
		}
		if r.DeploymentConfiguration.MinimumHealthyPercent > 0 {
			content += fmt.Sprintf(`
    minimum_healthy_percent = %d`,
				r.DeploymentConfiguration.MinimumHealthyPercent)
		}
		if r.DeploymentConfiguration.DeploymentCircuitBreaker != nil {
			content += `
    deployment_circuit_breaker {`
			if r.DeploymentConfiguration.DeploymentCircuitBreaker.Enable {
				content += `
      enable = true`
			}
			if r.DeploymentConfiguration.DeploymentCircuitBreaker.Rollback {
				content += `
      rollback = true`
			}
			content += `
    }`
		}
		content += `
  }`
	}

	// Add placement constraints if specified
	if len(r.PlacementConstraints) > 0 {
		content += `
  placement_constraints {`
		for _, pc := range r.PlacementConstraints {
			content += fmt.Sprintf(`
    type       = "%s"
    expression = "%s"`,
				pc.Type,
				pc.Expression)
		}
		content += `
  }`
	}

	// Add placement strategy if specified
	if len(r.PlacementStrategy) > 0 {
		content += `
  placement_strategy {`
		for _, ps := range r.PlacementStrategy {
			content += fmt.Sprintf(`
    type  = "%s"
    field = "%s"`,
				ps.Type,
				ps.Field)
		}
		content += `
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

// EKSResource represents an AWS EKS cluster
type EKSResource struct {
	BaseResource
	Version            string `json:"version"`
	RoleARN            string `json:"role_arn"`
	PlatformVersion    string `json:"platform_version"`
	Endpoint           string `json:"endpoint"`
	CertificateAuthority *CertificateAuthority `json:"certificate_authority,omitempty"`
	VPCConfig          *VPCConfig `json:"vpc_config"`
	EncryptionConfig   *EncryptionConfig `json:"encryption_config,omitempty"`
	OutpostConfig      *OutpostConfig `json:"outpost_config,omitempty"`
	AccessConfig       *AccessConfig `json:"access_config,omitempty"`
	EnabledClusterLogTypes []string `json:"enabled_cluster_log_types,omitempty"`
	DefaultAddonToVersion map[string]string `json:"default_addon_to_version,omitempty"`
}

type CertificateAuthority struct {
	Data string `json:"data"`
}

type VPCConfig struct {
	SubnetIDs              []string `json:"subnet_ids"`
	SecurityGroupIDs       []string `json:"security_group_ids"`
	EndpointPrivateAccess  bool     `json:"endpoint_private_access"`
	EndpointPublicAccess   bool     `json:"endpoint_public_access"`
	PublicAccessCIDRs      []string `json:"public_access_cidrs,omitempty"`
}

type EncryptionConfig struct {
	Provider *EncryptionProvider `json:"provider"`
	Resources []string `json:"resources"`
}

type EncryptionProvider struct {
	KeyARN string `json:"key_arn"`
}

type OutpostConfig struct {
	OutpostARNs []string `json:"outpost_arns"`
	ControlPlaneInstanceType string `json:"control_plane_instance_type"`
	ControlPlanePlacement *ControlPlanePlacement `json:"control_plane_placement,omitempty"`
}

type ControlPlanePlacement struct {
	GroupName string `json:"group_name"`
}

type AccessConfig struct {
	AuthenticationMode string `json:"authentication_mode"`
	BootstrapClusterCreatorAdminPermissions bool `json:"bootstrap_cluster_creator_admin_permissions"`
}

func (r *EKSResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_eks_cluster" "%s" {
  name     = "%s"
  role_arn = "%s"`,
		resourceName,
		r.Name,
		r.RoleARN)

	// Add version if specified
	if r.Version != "" {
		content += fmt.Sprintf(`
  version = "%s"`,
			r.Version)
	}

	// Add platform version if specified
	if r.PlatformVersion != "" {
		content += fmt.Sprintf(`
  platform_version = "%s"`,
			r.PlatformVersion)
	}

	// Add VPC config
	if r.VPCConfig != nil {
		content += `
  vpc_config {`
		if len(r.VPCConfig.SubnetIDs) > 0 {
			content += `
    subnet_ids = [`
			for i, subnet := range r.VPCConfig.SubnetIDs {
				if i > 0 {
					content += ","
				}
				content += fmt.Sprintf(`"%s"`, subnet)
			}
			content += `]`
		}
		if len(r.VPCConfig.SecurityGroupIDs) > 0 {
			content += `
    security_group_ids = [`
			for i, sg := range r.VPCConfig.SecurityGroupIDs {
				if i > 0 {
					content += ","
				}
				content += fmt.Sprintf(`"%s"`, sg)
			}
			content += `]`
		}
		if r.VPCConfig.EndpointPrivateAccess {
			content += `
    endpoint_private_access = true`
		}
		if r.VPCConfig.EndpointPublicAccess {
			content += `
    endpoint_public_access = true`
		}
		if len(r.VPCConfig.PublicAccessCIDRs) > 0 {
			content += `
    public_access_cidrs = [`
			for i, cidr := range r.VPCConfig.PublicAccessCIDRs {
				if i > 0 {
					content += ","
				}
				content += fmt.Sprintf(`"%s"`, cidr)
			}
			content += `]`
		}
		content += `
  }`
	}

	// Add encryption config if specified
	if r.EncryptionConfig != nil {
		content += `
  encryption_config {`
		if r.EncryptionConfig.Provider != nil {
			content += fmt.Sprintf(`
    provider {
      key_arn = "%s"
    }`,
				r.EncryptionConfig.Provider.KeyARN)
		}
		if len(r.EncryptionConfig.Resources) > 0 {
			content += `
    resources = [`
			for i, resource := range r.EncryptionConfig.Resources {
				if i > 0 {
					content += ","
				}
				content += fmt.Sprintf(`"%s"`, resource)
			}
			content += `]`
		}
		content += `
  }`
	}

	// Add outpost config if specified
	if r.OutpostConfig != nil {
		content += `
  outpost_config {`
		if len(r.OutpostConfig.OutpostARNs) > 0 {
			content += `
    outpost_arns = [`
			for i, arn := range r.OutpostConfig.OutpostARNs {
				if i > 0 {
					content += ","
				}
				content += fmt.Sprintf(`"%s"`, arn)
			}
			content += `]`
		}
		if r.OutpostConfig.ControlPlaneInstanceType != "" {
			content += fmt.Sprintf(`
    control_plane_instance_type = "%s"`,
				r.OutpostConfig.ControlPlaneInstanceType)
		}
		if r.OutpostConfig.ControlPlanePlacement != nil {
			content += fmt.Sprintf(`
    control_plane_placement {
      group_name = "%s"
    }`,
				r.OutpostConfig.ControlPlanePlacement.GroupName)
		}
		content += `
  }`
	}

	// Add access config if specified
	if r.AccessConfig != nil {
		content += `
  access_config {`
		if r.AccessConfig.AuthenticationMode != "" {
			content += fmt.Sprintf(`
    authentication_mode = "%s"`,
				r.AccessConfig.AuthenticationMode)
		}
		if r.AccessConfig.BootstrapClusterCreatorAdminPermissions {
			content += `
    bootstrap_cluster_creator_admin_permissions = true`
		}
		content += `
  }`
	}

	// Add enabled cluster log types if specified
	if len(r.EnabledClusterLogTypes) > 0 {
		content += `
  enabled_cluster_log_types = [`
		for i, logType := range r.EnabledClusterLogTypes {
			if i > 0 {
				content += ","
			}
			content += fmt.Sprintf(`"%s"`, logType)
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
  }
}`

	return content, nil
}

// ECRResource represents an AWS ECR repository
type ECRResource struct {
	BaseResource
	RepositoryURI      string `json:"repository_uri"`
	ImageTagMutability string `json:"image_tag_mutability"`
	ScanOnPush         bool   `json:"scan_on_push"`
	EncryptionConfiguration *EncryptionConfiguration `json:"encryption_configuration,omitempty"`
	ImageScanningConfiguration *ImageScanningConfiguration `json:"image_scanning_configuration,omitempty"`
	LifecyclePolicy    *LifecyclePolicy `json:"lifecycle_policy,omitempty"`
}

type EncryptionConfiguration struct {
	EncryptionType string `json:"encryption_type"`
	KMSKey         string `json:"kms_key,omitempty"`
}

type ImageScanningConfiguration struct {
	ScanOnPush bool `json:"scan_on_push"`
}

type LifecyclePolicy struct {
	Policy string `json:"policy"`
}

func (r *ECRResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_ecr_repository" "%s" {
  name                 = "%s"
  image_tag_mutability = "%s"`,
		resourceName,
		r.Name,
		r.ImageTagMutability)

	// Add scan on push if enabled
	if r.ScanOnPush {
		content += `
  scan_on_push = true`
	}

	// Add encryption configuration if specified
	if r.EncryptionConfiguration != nil {
		content += `
  encryption_configuration {`
		content += fmt.Sprintf(`
    encryption_type = "%s"`,
			r.EncryptionConfiguration.EncryptionType)
		if r.EncryptionConfiguration.KMSKey != "" {
			content += fmt.Sprintf(`
    kms_key = "%s"`,
				r.EncryptionConfiguration.KMSKey)
		}
		content += `
  }`
	}

	// Add image scanning configuration if specified
	if r.ImageScanningConfiguration != nil {
		content += `
  image_scanning_configuration {`
		if r.ImageScanningConfiguration.ScanOnPush {
			content += `
    scan_on_push = true`
		}
		content += `
  }`
	}

	// Add lifecycle policy if specified
	if r.LifecyclePolicy != nil {
		content += fmt.Sprintf(`
  lifecycle_policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 30 images"
        selection = {
          tagStatus     = "tagged"
          tagPrefixList = ["v"]
          countType     = "imageCountMoreThan"
          countNumber   = 30
        }
        action = {
          type = "expire"
        }
      }
    ]
  })`)
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

// SQSResource represents an AWS SQS queue
type SQSResource struct {
	BaseResource
	URL                string `json:"url"`
	DelaySeconds       int32  `json:"delay_seconds"`
	MaxMessageSize     int32  `json:"max_message_size"`
	MessageRetentionSeconds int32 `json:"message_retention_seconds"`
	ReceiveWaitTimeSeconds int32 `json:"receive_wait_time_seconds"`
	VisibilityTimeoutSeconds int32 `json:"visibility_timeout_seconds"`
	FIFOQueue         bool   `json:"fifo_queue"`
	ContentBasedDeduplication bool `json:"content_based_deduplication"`
	DeduplicationScope string `json:"deduplication_scope,omitempty"`
	FIFOThroughputLimit string `json:"fifo_throughput_limit,omitempty"`
	RedrivePolicy      *RedrivePolicy `json:"redrive_policy,omitempty"`
	DeadLetterTargetARN string `json:"dead_letter_target_arn,omitempty"`
	MaxReceiveCount    int32  `json:"max_receive_count,omitempty"`
}

type RedrivePolicy struct {
	DeadLetterTargetARN string `json:"dead_letter_target_arn"`
	MaxReceiveCount     int32  `json:"max_receive_count"`
}

func (r *SQSResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_sqs_queue" "%s" {
  name = "%s"`,
		resourceName,
		r.Name)

	// Add delay seconds if specified
	if r.DelaySeconds > 0 {
		content += fmt.Sprintf(`
  delay_seconds = %d`,
			r.DelaySeconds)
	}

	// Add max message size if specified
	if r.MaxMessageSize > 0 {
		content += fmt.Sprintf(`
  max_message_size = %d`,
			r.MaxMessageSize)
	}

	// Add message retention seconds if specified
	if r.MessageRetentionSeconds > 0 {
		content += fmt.Sprintf(`
  message_retention_seconds = %d`,
			r.MessageRetentionSeconds)
	}

	// Add receive wait time seconds if specified
	if r.ReceiveWaitTimeSeconds > 0 {
		content += fmt.Sprintf(`
  receive_wait_time_seconds = %d`,
			r.ReceiveWaitTimeSeconds)
	}

	// Add visibility timeout seconds if specified
	if r.VisibilityTimeoutSeconds > 0 {
		content += fmt.Sprintf(`
  visibility_timeout_seconds = %d`,
			r.VisibilityTimeoutSeconds)
	}

	// Add FIFO queue settings if enabled
	if r.FIFOQueue {
		content += `
  fifo_queue = true`
		if r.ContentBasedDeduplication {
			content += `
  content_based_deduplication = true`
		}
		if r.DeduplicationScope != "" {
			content += fmt.Sprintf(`
  deduplication_scope = "%s"`,
				r.DeduplicationScope)
		}
		if r.FIFOThroughputLimit != "" {
			content += fmt.Sprintf(`
  fifo_throughput_limit = "%s"`,
				r.FIFOThroughputLimit)
		}
	}

	// Add redrive policy if specified
	if r.RedrivePolicy != nil {
		content += fmt.Sprintf(`
  redrive_policy = jsonencode({
    deadLetterTargetArn = "%s"
    maxReceiveCount     = %d
  })`,
			r.RedrivePolicy.DeadLetterTargetARN,
			r.RedrivePolicy.MaxReceiveCount)
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

// SNSResource represents an AWS SNS topic
type SNSResource struct {
	BaseResource
	ARN                string `json:"arn"`
	DisplayName        string `json:"display_name,omitempty"`
	KMSMasterKeyID     string `json:"kms_master_key_id,omitempty"`
	DeliveryPolicy     string `json:"delivery_policy,omitempty"`
	Policy             string `json:"policy,omitempty"`
	ApplicationSuccessFeedbackRoleARN string `json:"application_success_feedback_role_arn,omitempty"`
	ApplicationFailureFeedbackRoleARN string `json:"application_failure_feedback_role_arn,omitempty"`
	ApplicationSuccessFeedbackSampleRate int32 `json:"application_success_feedback_sample_rate,omitempty"`
	HTTPSuccessFeedbackRoleARN string `json:"http_success_feedback_role_arn,omitempty"`
	HTTPFailureFeedbackRoleARN string `json:"http_failure_feedback_role_arn,omitempty"`
	HTTPSuccessFeedbackSampleRate int32 `json:"http_success_feedback_sample_rate,omitempty"`
	LambdaSuccessFeedbackRoleARN string `json:"lambda_success_feedback_role_arn,omitempty"`
	LambdaFailureFeedbackRoleARN string `json:"lambda_failure_feedback_role_arn,omitempty"`
	LambdaSuccessFeedbackSampleRate int32 `json:"lambda_success_feedback_sample_rate,omitempty"`
	SQSSuccessFeedbackRoleARN string `json:"sqs_success_feedback_role_arn,omitempty"`
	SQSFailureFeedbackRoleARN string `json:"sqs_failure_feedback_role_arn,omitempty"`
	SQSSuccessFeedbackSampleRate int32 `json:"sqs_success_feedback_sample_rate,omitempty"`
}

func (r *SNSResource) GetARN() string {
	return r.ARN
}

func (r *SNSResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_sns_topic" "%s" {
  name = "%s"`,
		resourceName,
		r.Name)

	// Add display name if specified
	if r.DisplayName != "" {
		content += fmt.Sprintf(`
  display_name = "%s"`,
			r.DisplayName)
	}

	// Add KMS master key ID if specified
	if r.KMSMasterKeyID != "" {
		content += fmt.Sprintf(`
  kms_master_key_id = "%s"`,
			r.KMSMasterKeyID)
	}

	// Add delivery policy if specified
	if r.DeliveryPolicy != "" {
		content += fmt.Sprintf(`
  delivery_policy = jsonencode(%s)`,
			r.DeliveryPolicy)
	}

	// Add policy if specified
	if r.Policy != "" {
		content += fmt.Sprintf(`
  policy = jsonencode(%s)`,
			r.Policy)
	}

	// Add application feedback settings if specified
	if r.ApplicationSuccessFeedbackRoleARN != "" {
		content += fmt.Sprintf(`
  application_success_feedback_role_arn = "%s"`,
			r.ApplicationSuccessFeedbackRoleARN)
	}
	if r.ApplicationFailureFeedbackRoleARN != "" {
		content += fmt.Sprintf(`
  application_failure_feedback_role_arn = "%s"`,
			r.ApplicationFailureFeedbackRoleARN)
	}
	if r.ApplicationSuccessFeedbackSampleRate > 0 {
		content += fmt.Sprintf(`
  application_success_feedback_sample_rate = %d`,
			r.ApplicationSuccessFeedbackSampleRate)
	}

	// Add HTTP feedback settings if specified
	if r.HTTPSuccessFeedbackRoleARN != "" {
		content += fmt.Sprintf(`
  http_success_feedback_role_arn = "%s"`,
			r.HTTPSuccessFeedbackRoleARN)
	}
	if r.HTTPFailureFeedbackRoleARN != "" {
		content += fmt.Sprintf(`
  http_failure_feedback_role_arn = "%s"`,
			r.HTTPFailureFeedbackRoleARN)
	}
	if r.HTTPSuccessFeedbackSampleRate > 0 {
		content += fmt.Sprintf(`
  http_success_feedback_sample_rate = %d`,
			r.HTTPSuccessFeedbackSampleRate)
	}

	// Add Lambda feedback settings if specified
	if r.LambdaSuccessFeedbackRoleARN != "" {
		content += fmt.Sprintf(`
  lambda_success_feedback_role_arn = "%s"`,
			r.LambdaSuccessFeedbackRoleARN)
	}
	if r.LambdaFailureFeedbackRoleARN != "" {
		content += fmt.Sprintf(`
  lambda_failure_feedback_role_arn = "%s"`,
			r.LambdaFailureFeedbackRoleARN)
	}
	if r.LambdaSuccessFeedbackSampleRate > 0 {
		content += fmt.Sprintf(`
  lambda_success_feedback_sample_rate = %d`,
			r.LambdaSuccessFeedbackSampleRate)
	}

	// Add SQS feedback settings if specified
	if r.SQSSuccessFeedbackRoleARN != "" {
		content += fmt.Sprintf(`
  sqs_success_feedback_role_arn = "%s"`,
			r.SQSSuccessFeedbackRoleARN)
	}
	if r.SQSFailureFeedbackRoleARN != "" {
		content += fmt.Sprintf(`
  sqs_failure_feedback_role_arn = "%s"`,
			r.SQSFailureFeedbackRoleARN)
	}
	if r.SQSSuccessFeedbackSampleRate > 0 {
		content += fmt.Sprintf(`
  sqs_success_feedback_sample_rate = %d`,
			r.SQSSuccessFeedbackSampleRate)
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

// CloudWatchResource represents an AWS CloudWatch log group
type CloudWatchResource struct {
	BaseResource
	LogGroupName       string `json:"log_group_name"`
	RetentionInDays    int32  `json:"retention_in_days"`
	KMSKeyID           string `json:"kms_key_id,omitempty"`
	SkipDestroy        bool   `json:"skip_destroy"`
}

func (r *CloudWatchResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_cloudwatch_log_group" "%s" {
  name = "%s"`,
		resourceName,
		r.LogGroupName)

	// Add retention in days if specified
	if r.RetentionInDays > 0 {
		content += fmt.Sprintf(`
  retention_in_days = %d`,
			r.RetentionInDays)
	}

	// Add KMS key ID if specified
	if r.KMSKeyID != "" {
		content += fmt.Sprintf(`
  kms_key_id = "%s"`,
			r.KMSKeyID)
	}

	// Add skip destroy if enabled
	if r.SkipDestroy {
		content += `
  skip_destroy = true`
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

// ElasticsearchResource represents an AWS Elasticsearch domain
type ElasticsearchResource struct {
	BaseResource
	DomainName           string `json:"domain_name"`
	ElasticsearchVersion string `json:"elasticsearch_version"`
	ClusterConfig        *ElasticsearchClusterConfig `json:"cluster_config,omitempty"`
	EBSOptions           *ElasticsearchEBSOptions `json:"ebs_options,omitempty"`
	SnapshotOptions      *ElasticsearchSnapshotOptions `json:"snapshot_options,omitempty"`
	VPCOptions           *ElasticsearchVPCOptions `json:"vpc_options,omitempty"`
	EncryptAtRest        *ElasticsearchEncryptAtRest `json:"encrypt_at_rest,omitempty"`
	NodeToNodeEncryption *ElasticsearchNodeToNodeEncryption `json:"node_to_node_encryption,omitempty"`
	DomainEndpointOptions *ElasticsearchDomainEndpointOptions `json:"domain_endpoint_options,omitempty"`
	AdvancedOptions      map[string]string `json:"advanced_options,omitempty"`
}

type ElasticsearchClusterConfig struct {
	InstanceType            string `json:"instance_type"`
	InstanceCount           int32  `json:"instance_count"`
	DedicatedMasterEnabled  bool   `json:"dedicated_master_enabled"`
	DedicatedMasterType     string `json:"dedicated_master_type,omitempty"`
	DedicatedMasterCount    int32  `json:"dedicated_master_count,omitempty"`
	ZoneAwarenessEnabled    bool   `json:"zone_awareness_enabled"`
	WarmEnabled             bool   `json:"warm_enabled"`
	WarmType                string `json:"warm_type,omitempty"`
	WarmCount               int32  `json:"warm_count,omitempty"`
	ColdStorageOptions      *ElasticsearchColdStorageOptions `json:"cold_storage_options,omitempty"`
}

type ElasticsearchEBSOptions struct {
	EBSEnabled bool   `json:"ebs_enabled"`
	VolumeType string `json:"volume_type"`
	VolumeSize int32  `json:"volume_size"`
	Iops       int32  `json:"iops,omitempty"`
}

type ElasticsearchSnapshotOptions struct {
	AutomatedSnapshotStartHour int32 `json:"automated_snapshot_start_hour"`
}

type ElasticsearchVPCOptions struct {
	SubnetIDs        []string `json:"subnet_ids"`
	SecurityGroupIDs []string `json:"security_group_ids"`
}

type ElasticsearchEncryptAtRest struct {
	Enabled  bool   `json:"enabled"`
	KMSKeyID string `json:"kms_key_id,omitempty"`
}

type ElasticsearchNodeToNodeEncryption struct {
	Enabled bool `json:"enabled"`
}

type ElasticsearchDomainEndpointOptions struct {
	EnforceHTTPS       bool   `json:"enforce_https"`
	TLSSecurityPolicy  string `json:"tls_security_policy"`
	CustomEndpointEnabled bool `json:"custom_endpoint_enabled"`
	CustomEndpoint      string `json:"custom_endpoint,omitempty"`
	CustomEndpointCertificateARN string `json:"custom_endpoint_certificate_arn,omitempty"`
}

type ElasticsearchColdStorageOptions struct {
	Enabled bool `json:"enabled"`
}

func (r *ElasticsearchResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_elasticsearch_domain" "%s" {
  domain_name           = "%s"
  elasticsearch_version = "%s"`,
		resourceName, r.DomainName, r.ElasticsearchVersion)

	if r.ClusterConfig != nil {
		content += fmt.Sprintf(`
  cluster_config {
    instance_type = "%s"
    instance_count = %d
    dedicated_master_enabled = %t
    zone_awareness_enabled = %t`,
			r.ClusterConfig.InstanceType, r.ClusterConfig.InstanceCount,
			r.ClusterConfig.DedicatedMasterEnabled, r.ClusterConfig.ZoneAwarenessEnabled)
		
		if r.ClusterConfig.DedicatedMasterType != "" {
			content += fmt.Sprintf(`
    dedicated_master_type = "%s"`, r.ClusterConfig.DedicatedMasterType)
		}
		if r.ClusterConfig.DedicatedMasterCount > 0 {
			content += fmt.Sprintf(`
    dedicated_master_count = %d`, r.ClusterConfig.DedicatedMasterCount)
		}
		content += `
  }`
	}

	if r.EBSOptions != nil {
		content += fmt.Sprintf(`
  ebs_options {
    ebs_enabled = %t
    volume_type = "%s"
    volume_size = %d`,
			r.EBSOptions.EBSEnabled, r.EBSOptions.VolumeType, r.EBSOptions.VolumeSize)
		if r.EBSOptions.Iops > 0 {
			content += fmt.Sprintf(`
    iops = %d`, r.EBSOptions.Iops)
		}
		content += `
  }`
	}

	if r.EncryptAtRest != nil {
		content += fmt.Sprintf(`
  encrypt_at_rest {
    enabled = %t`,
			r.EncryptAtRest.Enabled)
		if r.EncryptAtRest.KMSKeyID != "" {
			content += fmt.Sprintf(`
    kms_key_id = "%s"`, r.EncryptAtRest.KMSKeyID)
		}
		content += `
  }`
	}

	if r.NodeToNodeEncryption != nil {
		content += fmt.Sprintf(`
  node_to_node_encryption {
    enabled = %t
  }`,
			r.NodeToNodeEncryption.Enabled)
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// NeptuneResource represents an AWS Neptune cluster
type NeptuneResource struct {
	BaseResource
	ClusterIdentifier     string `json:"cluster_identifier"`
	Engine               string `json:"engine"`
	EngineVersion        string `json:"engine_version"`
	AvailabilityZones    []string `json:"availability_zones"`
	BackupRetentionPeriod int32  `json:"backup_retention_period"`
	PreferredBackupWindow string `json:"preferred_backup_window"`
	PreferredMaintenanceWindow string `json:"preferred_maintenance_window"`
	Port                 int32  `json:"port"`
	DBSubnetGroupName    string `json:"db_subnet_group_name"`
	VpcSecurityGroupIds  []string `json:"vpc_security_group_ids"`
	StorageEncrypted     bool   `json:"storage_encrypted"`
	KMSKeyARN            string `json:"kms_key_arn,omitempty"`
	SkipFinalSnapshot    bool   `json:"skip_final_snapshot"`
	FinalSnapshotIdentifier string `json:"final_snapshot_identifier,omitempty"`
	DeletionProtection   bool   `json:"deletion_protection"`
	EnableCloudwatchLogsExports []string `json:"enable_cloudwatch_logs_exports,omitempty"`
}

func (r *NeptuneResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_neptune_cluster" "%s" {
  cluster_identifier = "%s"
  engine = "%s"
  engine_version = "%s"
  backup_retention_period = %d
  preferred_backup_window = "%s"
  preferred_maintenance_window = "%s"
  port = %d
  db_subnet_group_name = "%s"
  storage_encrypted = %t
  skip_final_snapshot = %t
  deletion_protection = %t`,
		resourceName, r.ClusterIdentifier, r.Engine, r.EngineVersion,
		r.BackupRetentionPeriod, r.PreferredBackupWindow, r.PreferredMaintenanceWindow,
		r.Port, r.DBSubnetGroupName, r.StorageEncrypted, r.SkipFinalSnapshot, r.DeletionProtection)

	if len(r.AvailabilityZones) > 0 {
		content += fmt.Sprintf(`
  availability_zones = %s`, formatStringSlice(r.AvailabilityZones))
	}

	if len(r.VpcSecurityGroupIds) > 0 {
		content += fmt.Sprintf(`
  vpc_security_group_ids = %s`, formatStringSlice(r.VpcSecurityGroupIds))
	}

	if r.KMSKeyARN != "" {
		content += fmt.Sprintf(`
  kms_key_arn = "%s"`, r.KMSKeyARN)
	}

	if r.FinalSnapshotIdentifier != "" {
		content += fmt.Sprintf(`
  final_snapshot_identifier = "%s"`, r.FinalSnapshotIdentifier)
	}

	if len(r.EnableCloudwatchLogsExports) > 0 {
		content += fmt.Sprintf(`
  enable_cloudwatch_logs_exports = %s`, formatStringSlice(r.EnableCloudwatchLogsExports))
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// DocDBResource represents an AWS DocumentDB cluster
type DocDBResource struct {
	BaseResource
	ClusterIdentifier     string `json:"cluster_identifier"`
	Engine               string `json:"engine"`
	EngineVersion        string `json:"engine_version"`
	AvailabilityZones    []string `json:"availability_zones"`
	BackupRetentionPeriod int32  `json:"backup_retention_period"`
	PreferredBackupWindow string `json:"preferred_backup_window"`
	PreferredMaintenanceWindow string `json:"preferred_maintenance_window"`
	Port                 int32  `json:"port"`
	DBSubnetGroupName    string `json:"db_subnet_group_name"`
	VpcSecurityGroupIds  []string `json:"vpc_security_group_ids"`
	StorageEncrypted     bool   `json:"storage_encrypted"`
	KMSKeyARN            string `json:"kms_key_arn,omitempty"`
	SkipFinalSnapshot    bool   `json:"skip_final_snapshot"`
	FinalSnapshotIdentifier string `json:"final_snapshot_identifier,omitempty"`
	DeletionProtection   bool   `json:"deletion_protection"`
	EnableCloudwatchLogsExports []string `json:"enable_cloudwatch_logs_exports,omitempty"`
}

func (r *DocDBResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_docdb_cluster" "%s" {
  cluster_identifier = "%s"
  engine = "%s"
  engine_version = "%s"
  backup_retention_period = %d
  preferred_backup_window = "%s"
  preferred_maintenance_window = "%s"
  port = %d
  db_subnet_group_name = "%s"
  storage_encrypted = %t
  skip_final_snapshot = %t
  deletion_protection = %t`,
		resourceName, r.ClusterIdentifier, r.Engine, r.EngineVersion,
		r.BackupRetentionPeriod, r.PreferredBackupWindow, r.PreferredMaintenanceWindow,
		r.Port, r.DBSubnetGroupName, r.StorageEncrypted, r.SkipFinalSnapshot, r.DeletionProtection)

	if len(r.AvailabilityZones) > 0 {
		content += fmt.Sprintf(`
  availability_zones = %s`, formatStringSlice(r.AvailabilityZones))
	}

	if len(r.VpcSecurityGroupIds) > 0 {
		content += fmt.Sprintf(`
  vpc_security_group_ids = %s`, formatStringSlice(r.VpcSecurityGroupIds))
	}

	if r.KMSKeyARN != "" {
		content += fmt.Sprintf(`
  kms_key_arn = "%s"`, r.KMSKeyARN)
	}

	if r.FinalSnapshotIdentifier != "" {
		content += fmt.Sprintf(`
  final_snapshot_identifier = "%s"`, r.FinalSnapshotIdentifier)
	}

	if len(r.EnableCloudwatchLogsExports) > 0 {
		content += fmt.Sprintf(`
  enable_cloudwatch_logs_exports = %s`, formatStringSlice(r.EnableCloudwatchLogsExports))
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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
	
	content := fmt.Sprintf(`resource "%s" "%s" {
  # Generic resource - review and customize as needed`,
		r.ResourceType, resourceName)

	// Add attributes
	for key, value := range r.Attributes {
		switch v := value.(type) {
		case string:
			content += fmt.Sprintf(`
  %s = "%s"`, key, v)
		case bool:
			content += fmt.Sprintf(`
  %s = %t`, key, v)
		case int, int32, int64:
			content += fmt.Sprintf(`
  %s = %v`, key, v)
		default:
			content += fmt.Sprintf(`
  # %s = %v # Unhandled type`, key, v)
		}
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
			content += fmt.Sprintf(`
    %s = "%s"`, k, v)
		}
	}
	content += `
  }
}`

	return content, nil
}

// Helper function to format string slices for OpenTofu
func formatStringSlice(slice []string) string {
	if len(slice) == 0 {
		return "[]"
	}
	
	quoted := make([]string, len(slice))
	for i, s := range slice {
		quoted[i] = fmt.Sprintf(`"%s"`, s)
	}
	return fmt.Sprintf(`[%s]`, strings.Join(quoted, ", "))
}

// ElasticBeanstalkResource represents an AWS Elastic Beanstalk environment
type ElasticBeanstalkResource struct {
	BaseResource
	ApplicationName      string `json:"application_name"`
	EnvironmentName      string `json:"environment_name"`
	SolutionStackName    string `json:"solution_stack_name"`
	PlatformARN          string `json:"platform_arn,omitempty"`
	TemplateName         string `json:"template_name,omitempty"`
	VersionLabel         string `json:"version_label,omitempty"`
	Description          string `json:"description,omitempty"`
	Setting             []*ElasticBeanstalkSetting `json:"setting,omitempty"`
	AllSettings         []*ElasticBeanstalkAllSetting `json:"all_settings,omitempty"`
	WaitForReadyTimeout  string `json:"wait_for_ready_timeout,omitempty"`
	PollInterval         string `json:"poll_interval,omitempty"`
}

type ElasticBeanstalkSetting struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	Resource  string `json:"resource,omitempty"`
}

type ElasticBeanstalkAllSetting struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	Resource  string `json:"resource,omitempty"`
}

func (r *ElasticBeanstalkResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_elastic_beanstalk_environment" "%s" {
  name                = "%s"
  application         = "%s"
  solution_stack_name = "%s"`,
		resourceName, r.EnvironmentName, r.ApplicationName, r.SolutionStackName)

	if r.PlatformARN != "" {
		content += fmt.Sprintf(`
  platform_arn = "%s"`, r.PlatformARN)
	}

	if r.TemplateName != "" {
		content += fmt.Sprintf(`
  template_name = "%s"`, r.TemplateName)
	}

	if r.VersionLabel != "" {
		content += fmt.Sprintf(`
  version_label = "%s"`, r.VersionLabel)
	}

	if r.Description != "" {
		content += fmt.Sprintf(`
  description = "%s"`, r.Description)
	}

	if len(r.Setting) > 0 {
		content += `
  setting {`
		for _, setting := range r.Setting {
			content += fmt.Sprintf(`
    namespace = "%s"
    name      = "%s"
    value     = "%s"`,
				setting.Namespace, setting.Name, setting.Value)
			if setting.Resource != "" {
				content += fmt.Sprintf(`
    resource  = "%s"`, setting.Resource)
			}
		}
		content += `
  }`
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// CodeBuildResource represents an AWS CodeBuild project
type CodeBuildResource struct {
	BaseResource
	ProjectName          string `json:"project_name"`
	Description          string `json:"description,omitempty"`
	BuildTimeout         int32  `json:"build_timeout"`
	QueuedTimeout        int32  `json:"queued_timeout"`
	ServiceRole          string `json:"service_role"`
	Source               *CodeBuildSource `json:"source"`
	Artifacts            *CodeBuildArtifacts `json:"artifacts"`
	Environment          *CodeBuildEnvironment `json:"environment"`
	Cache                *CodeBuildCache `json:"cache,omitempty"`
	LogsConfig           *CodeBuildLogsConfig `json:"logs_config,omitempty"`
	VPCConfig            *CodeBuildVPCConfig `json:"vpc_config,omitempty"`
	SecondaryArtifacts   []*CodeBuildArtifacts `json:"secondary_artifacts,omitempty"`
	SecondarySources     []*CodeBuildSource `json:"secondary_sources,omitempty"`
}

type CodeBuildSource struct {
	Type            string `json:"type"`
	Location        string `json:"location,omitempty"`
	GitCloneDepth   int32  `json:"git_clone_depth,omitempty"`
	Buildspec       string `json:"buildspec,omitempty"`
	ReportBuildStatus bool `json:"report_build_status"`
}

type CodeBuildArtifacts struct {
	Type                string `json:"type"`
	Location            string `json:"location,omitempty"`
	Path                string `json:"path,omitempty"`
	NamespaceType       string `json:"namespace_type,omitempty"`
	Name                string `json:"name,omitempty"`
	Packaging           string `json:"packaging,omitempty"`
	OverrideArtifactName bool `json:"override_artifact_name"`
	EncryptionDisabled  bool `json:"encryption_disabled"`
	ArtifactIdentifier  string `json:"artifact_identifier,omitempty"`
}

type CodeBuildEnvironment struct {
	Type                        string `json:"type"`
	Image                       string `json:"image"`
	ComputeType                 string `json:"compute_type"`
	ImagePullCredentialsType    string `json:"image_pull_credentials_type,omitempty"`
	Certificate                 string `json:"certificate,omitempty"`
	PrivilegedMode              bool `json:"privileged_mode"`
	EnvironmentVariables        []*CodeBuildEnvironmentVariable `json:"environment_variables,omitempty"`
}

type CodeBuildEnvironmentVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

type CodeBuildCache struct {
	Type     string `json:"type"`
	Location string `json:"location,omitempty"`
	Modes    []string `json:"modes,omitempty"`
}

type CodeBuildLogsConfig struct {
	CloudWatchLogs *CodeBuildCloudWatchLogs `json:"cloud_watch_logs,omitempty"`
	S3Logs         *CodeBuildS3Logs `json:"s3_logs,omitempty"`
}

type CodeBuildCloudWatchLogs struct {
	GroupName  string `json:"group_name"`
	StreamName string `json:"stream_name"`
	Status     string `json:"status"`
}

type CodeBuildS3Logs struct {
	Status              string `json:"status"`
	Location            string `json:"location"`
	EncryptionDisabled  bool `json:"encryption_disabled"`
}

type CodeBuildVPCConfig struct {
	VPCID             string   `json:"vpc_id"`
	Subnets           []string `json:"subnets"`
	SecurityGroupIDs  []string `json:"security_group_ids"`
}

func (r *CodeBuildResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_codebuild_project" "%s" {
  name          = "%s"
  description   = "%s"
  build_timeout = %d
  service_role  = "%s"`,
		resourceName, r.ProjectName, r.Description, r.BuildTimeout, r.ServiceRole)

	if r.QueuedTimeout > 0 {
		content += fmt.Sprintf(`
  queued_timeout = %d`, r.QueuedTimeout)
	}

	// Add source configuration
	if r.Source != nil {
		content += fmt.Sprintf(`
  source {
    type = "%s"`,
			r.Source.Type)
		if r.Source.Location != "" {
			content += fmt.Sprintf(`
    location = "%s"`, r.Source.Location)
		}
		if r.Source.GitCloneDepth > 0 {
			content += fmt.Sprintf(`
    git_clone_depth = %d`, r.Source.GitCloneDepth)
		}
		if r.Source.Buildspec != "" {
			content += fmt.Sprintf(`
    buildspec = "%s"`, r.Source.Buildspec)
		}
		content += fmt.Sprintf(`
    report_build_status = %t
  }`, r.Source.ReportBuildStatus)
	}

	// Add artifacts configuration
	if r.Artifacts != nil {
		content += fmt.Sprintf(`
  artifacts {
    type = "%s"`,
			r.Artifacts.Type)
		if r.Artifacts.Location != "" {
			content += fmt.Sprintf(`
    location = "%s"`, r.Artifacts.Location)
		}
		if r.Artifacts.Path != "" {
			content += fmt.Sprintf(`
    path = "%s"`, r.Artifacts.Path)
		}
		if r.Artifacts.NamespaceType != "" {
			content += fmt.Sprintf(`
    namespace_type = "%s"`, r.Artifacts.NamespaceType)
		}
		if r.Artifacts.Name != "" {
			content += fmt.Sprintf(`
    name = "%s"`, r.Artifacts.Name)
		}
		if r.Artifacts.Packaging != "" {
			content += fmt.Sprintf(`
    packaging = "%s"`, r.Artifacts.Packaging)
		}
		content += fmt.Sprintf(`
    override_artifact_name = %t
    encryption_disabled = %t
  }`, r.Artifacts.OverrideArtifactName, r.Artifacts.EncryptionDisabled)
	}

	// Add environment configuration
	if r.Environment != nil {
		content += fmt.Sprintf(`
  environment {
    type = "%s"
    image = "%s"
    compute_type = "%s"
    privileged_mode = %t`,
			r.Environment.Type, r.Environment.Image, r.Environment.ComputeType, r.Environment.PrivilegedMode)
		
		if r.Environment.ImagePullCredentialsType != "" {
			content += fmt.Sprintf(`
    image_pull_credentials_type = "%s"`, r.Environment.ImagePullCredentialsType)
		}
		if r.Environment.Certificate != "" {
			content += fmt.Sprintf(`
    certificate = "%s"`, r.Environment.Certificate)
		}
		
		if len(r.Environment.EnvironmentVariables) > 0 {
			content += `
    environment_variable {`
			for _, envVar := range r.Environment.EnvironmentVariables {
				content += fmt.Sprintf(`
      name  = "%s"
      value = "%s"`,
					envVar.Name, envVar.Value)
				if envVar.Type != "" {
					content += fmt.Sprintf(`
      type  = "%s"`, envVar.Type)
				}
			}
			content += `
    }`
		}
		content += `
  }`
	}

	// Add VPC configuration
	if r.VPCConfig != nil {
		content += fmt.Sprintf(`
  vpc_config {
    vpc_id = "%s"
    subnets = %s
    security_group_ids = %s
  }`,
			r.VPCConfig.VPCID, formatStringSlice(r.VPCConfig.Subnets), formatStringSlice(r.VPCConfig.SecurityGroupIDs))
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// CodeDeployResource represents an AWS CodeDeploy application
type CodeDeployResource struct {
	BaseResource
	ApplicationName      string `json:"application_name"`
	ComputePlatform      string `json:"compute_platform"`
	LinkedToGitHub       bool   `json:"linked_to_git_hub"`
	GitHubAccountName    string `json:"git_hub_account_name,omitempty"`
	DeploymentGroupName  string `json:"deployment_group_name,omitempty"`
	ServiceRoleARN       string `json:"service_role_arn,omitempty"`
	DeploymentStyle      *CodeDeployDeploymentStyle `json:"deployment_style,omitempty"`
	AutoScalingGroups    []string `json:"auto_scaling_groups,omitempty"`
	LoadBalancerInfo     *CodeDeployLoadBalancerInfo `json:"load_balancer_info,omitempty"`
	BlueGreenDeploymentConfig *CodeDeployBlueGreenDeploymentConfig `json:"blue_green_deployment_config,omitempty"`
	AlarmConfiguration   *CodeDeployAlarmConfiguration `json:"alarm_configuration,omitempty"`
	AutoRollbackConfiguration *CodeDeployAutoRollbackConfiguration `json:"auto_rollback_configuration,omitempty"`
	TriggerConfiguration *CodeDeployTriggerConfiguration `json:"trigger_configuration,omitempty"`
}

type CodeDeployDeploymentStyle struct {
	DeploymentType   string `json:"deployment_type"`
	DeploymentOption string `json:"deployment_option"`
}

type CodeDeployLoadBalancerInfo struct {
	ELBInfoList []*CodeDeployELBInfo `json:"elb_info_list,omitempty"`
	TargetGroupInfoList []*CodeDeployTargetGroupInfo `json:"target_group_info_list,omitempty"`
	TargetGroupPairInfo *CodeDeployTargetGroupPairInfo `json:"target_group_pair_info,omitempty"`
}

type CodeDeployELBInfo struct {
	Name string `json:"name"`
}

type CodeDeployTargetGroupInfo struct {
	Name string `json:"name"`
}

type CodeDeployTargetGroupPairInfo struct {
	TargetGroups []*CodeDeployTargetGroupInfo `json:"target_groups"`
	ProdTrafficRoute *CodeDeployTrafficRoute `json:"prod_traffic_route"`
	TestTrafficRoute *CodeDeployTrafficRoute `json:"test_traffic_route"`
}

type CodeDeployTrafficRoute struct {
	ListenerARNs []string `json:"listener_arns"`
}

type CodeDeployBlueGreenDeploymentConfig struct {
	TerminateBlueInstancesOnDeploymentSuccess *CodeDeployBlueInstanceTerminationOption `json:"terminate_blue_instances_on_deployment_success,omitempty"`
	DeploymentReadyOption *CodeDeployDeploymentReadyOption `json:"deployment_ready_option,omitempty"`
	GreenFleetProvisioningOption *CodeDeployGreenFleetProvisioningOption `json:"green_fleet_provisioning_option,omitempty"`
}

type CodeDeployBlueInstanceTerminationOption struct {
	Action string `json:"action"`
	TerminationWaitTimeInMinutes int32 `json:"termination_wait_time_in_minutes"`
}

type CodeDeployDeploymentReadyOption struct {
	ActionOnTimeout string `json:"action_on_timeout"`
	WaitTimeInMinutes int32 `json:"wait_time_in_minutes"`
}

type CodeDeployGreenFleetProvisioningOption struct {
	Action string `json:"action"`
}

type CodeDeployAlarmConfiguration struct {
	Enabled bool `json:"enabled"`
	Alarms  []string `json:"alarms,omitempty"`
	IgnorePollAlarmFailure bool `json:"ignore_poll_alarm_failure"`
}

type CodeDeployAutoRollbackConfiguration struct {
	Enabled bool `json:"enabled"`
	Events  []string `json:"events,omitempty"`
}

type CodeDeployTriggerConfiguration struct {
	TriggerEvents []string `json:"trigger_events"`
	TriggerName   string `json:"trigger_name"`
	TriggerTargetARN string `json:"trigger_target_arn"`
}

func (r *CodeDeployResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_codedeploy_app" "%s" {
  name             = "%s"
  compute_platform = "%s"`,
		resourceName, r.ApplicationName, r.ComputePlatform)

	if r.LinkedToGitHub {
		content += `
  linked_to_git_hub = true`
		if r.GitHubAccountName != "" {
			content += fmt.Sprintf(`
  git_hub_account_name = "%s"`, r.GitHubAccountName)
		}
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// ConfigResource represents an AWS Config configuration recorder
type ConfigResource struct {
	BaseResource
	RecorderName string `json:"recorder_name"`
	RoleARN      string `json:"role_arn"`
	RecordingGroup *ConfigRecordingGroup `json:"recording_group,omitempty"`
	RecordingMode *ConfigRecordingMode `json:"recording_mode,omitempty"`
}

type ConfigRecordingGroup struct {
	AllSupported bool `json:"all_supported"`
	IncludeGlobalResources bool `json:"include_global_resources"`
	ResourceTypes []string `json:"resource_types,omitempty"`
}

type ConfigRecordingMode struct {
	RecordingFrequency string `json:"recording_frequency"`
	MaximumExecutionFrequency string `json:"maximum_execution_frequency,omitempty"`
}

func (r *ConfigResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_config_configuration_recorder" "%s" {
  name = "%s"
  role_arn = "%s"`,
		resourceName, r.RecorderName, r.RoleARN)

	if r.RecordingGroup != nil {
		content += fmt.Sprintf(`
  recording_group {
    all_supported = %t
    include_global_resources = %t`,
			r.RecordingGroup.AllSupported, r.RecordingGroup.IncludeGlobalResources)
		
		if len(r.RecordingGroup.ResourceTypes) > 0 {
			content += fmt.Sprintf(`
    resource_types = %s`, formatStringSlice(r.RecordingGroup.ResourceTypes))
		}
		content += `
  }`
	}

	if r.RecordingMode != nil {
		content += fmt.Sprintf(`
  recording_mode {
    recording_frequency = "%s"`,
			r.RecordingMode.RecordingFrequency)
		
		if r.RecordingMode.MaximumExecutionFrequency != "" {
			content += fmt.Sprintf(`
    maximum_execution_frequency = "%s"`, r.RecordingMode.MaximumExecutionFrequency)
		}
		content += `
  }`
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// KinesisResource represents an AWS Kinesis stream
type KinesisResource struct {
	BaseResource
	StreamName           string `json:"stream_name"`
	ShardCount           int32  `json:"shard_count"`
	RetentionPeriodHours int32  `json:"retention_period_hours"`
	StreamARN            string `json:"stream_arn"`
	EncryptionType       string `json:"encryption_type,omitempty"`
	KMSKeyID             string `json:"kms_key_id,omitempty"`
	StreamModeDetails     *KinesisStreamModeDetails `json:"stream_mode_details,omitempty"`
}

type KinesisStreamModeDetails struct {
	StreamMode string `json:"stream_mode"`
}

func (r *KinesisResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_kinesis_stream" "%s" {
  name = "%s"
  shard_count = %d
  retention_period = %d`,
		resourceName, r.StreamName, r.ShardCount, r.RetentionPeriodHours)

	if r.EncryptionType != "" {
		content += fmt.Sprintf(`
  encryption_type = "%s"`, r.EncryptionType)
	}

	if r.KMSKeyID != "" {
		content += fmt.Sprintf(`
  kms_key_id = "%s"`, r.KMSKeyID)
	}

	if r.StreamModeDetails != nil {
		content += fmt.Sprintf(`
  stream_mode_details {
    stream_mode = "%s"
  }`,
			r.StreamModeDetails.StreamMode)
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// FSxResource represents an AWS FSx file system
type FSxResource struct {
	BaseResource
	FileSystemID         string `json:"file_system_id"`
	FileSystemType       string `json:"file_system_type"`
	StorageCapacity      int32  `json:"storage_capacity"`
	SubnetIDs            []string `json:"subnet_ids"`
	SecurityGroupIDs     []string `json:"security_group_ids"`
	KMSKeyID             string `json:"kms_key_id,omitempty"`
	StorageType          string `json:"storage_type"`
	ThroughputCapacity   int32  `json:"throughput_capacity,omitempty"`
	WeeklyMaintenanceStartTime string `json:"weekly_maintenance_start_time,omitempty"`
	DailyAutomaticBackupStartTime string `json:"daily_automatic_backup_start_time,omitempty"`
	AutomaticBackupRetentionDays int32 `json:"automatic_backup_retention_days"`
	CopyTagsToBackups    bool `json:"copy_tags_to_backups"`
	DeploymentType       string `json:"deployment_type"`
	PreferredSubnetID    string `json:"preferred_subnet_id,omitempty"`
	RouteTableIDs        []string `json:"route_table_ids,omitempty"`
}

func (r *FSxResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_fsx_windows_file_system" "%s" {
  file_system_type = "%s"
  storage_capacity = %d
  subnet_ids = %s
  security_group_ids = %s
  storage_type = "%s"
  automatic_backup_retention_days = %d
  copy_tags_to_backups = %t
  deployment_type = "%s"`,
		resourceName, r.FileSystemType, r.StorageCapacity, 
		formatStringSlice(r.SubnetIDs), formatStringSlice(r.SecurityGroupIDs),
		r.StorageType, r.AutomaticBackupRetentionDays, r.CopyTagsToBackups, r.DeploymentType)

	if r.KMSKeyID != "" {
		content += fmt.Sprintf(`
  kms_key_id = "%s"`, r.KMSKeyID)
	}

	if r.ThroughputCapacity > 0 {
		content += fmt.Sprintf(`
  throughput_capacity = %d`, r.ThroughputCapacity)
	}

	if r.WeeklyMaintenanceStartTime != "" {
		content += fmt.Sprintf(`
  weekly_maintenance_start_time = "%s"`, r.WeeklyMaintenanceStartTime)
	}

	if r.DailyAutomaticBackupStartTime != "" {
		content += fmt.Sprintf(`
  daily_automatic_backup_start_time = "%s"`, r.DailyAutomaticBackupStartTime)
	}

	if r.PreferredSubnetID != "" {
		content += fmt.Sprintf(`
  preferred_subnet_id = "%s"`, r.PreferredSubnetID)
	}

	if len(r.RouteTableIDs) > 0 {
		content += fmt.Sprintf(`
  route_table_ids = %s`, formatStringSlice(r.RouteTableIDs))
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// BackupResource represents an AWS Backup vault
type BackupResource struct {
	BaseResource
	BackupVaultName string `json:"backup_vault_name"`
	BackupVaultARN  string `json:"backup_vault_arn"`
	KMSKeyARN       string `json:"kms_key_arn,omitempty"`
	AccessPolicy    string `json:"access_policy,omitempty"`
	EncryptionKeyARN string `json:"encryption_key_arn,omitempty"`
	ForceDestroy    bool   `json:"force_destroy"`
}

func (r *BackupResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_backup_vault" "%s" {
  name = "%s"
  force_destroy = %t`,
		resourceName, r.BackupVaultName, r.ForceDestroy)

	if r.KMSKeyARN != "" {
		content += fmt.Sprintf(`
  kms_key_arn = "%s"`, r.KMSKeyARN)
	}

	if r.AccessPolicy != "" {
		content += fmt.Sprintf(`
  access_policy = %s`, r.AccessPolicy)
	}

	if r.EncryptionKeyARN != "" {
		content += fmt.Sprintf(`
  encryption_key_arn = "%s"`, r.EncryptionKeyARN)
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// GlacierResource represents an AWS Glacier vault
type GlacierResource struct {
	BaseResource
	VaultName string `json:"vault_name"`
	VaultARN  string `json:"vault_arn"`
	AccessPolicy string `json:"access_policy,omitempty"`
	Notification []*GlacierNotification `json:"notification,omitempty"`
}

type GlacierNotification struct {
	Events []string `json:"events"`
	SNSTopic string `json:"sns_topic"`
}

func (r *GlacierResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_glacier_vault" "%s" {
  name = "%s"`,
		resourceName, r.VaultName)

	if r.AccessPolicy != "" {
		content += fmt.Sprintf(`
  access_policy = %s`, r.AccessPolicy)
	}

	if len(r.Notification) > 0 {
		for _, notification := range r.Notification {
			content += fmt.Sprintf(`
  notification {
    events = %s
    sns_topic = "%s"
  }`,
				formatStringSlice(notification.Events), notification.SNSTopic)
		}
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// MQResource represents an AWS MQ broker
type MQResource struct {
	BaseResource
	BrokerName string `json:"broker_name"`
	BrokerID   string `json:"broker_id"`
	EngineType string `json:"engine_type"`
	EngineVersion string `json:"engine_version"`
	HostInstanceType string `json:"host_instance_type"`
	DeploymentMode string `json:"deployment_mode"`
	SecurityGroups []string `json:"security_groups"`
	SubnetIDs []string `json:"subnet_ids"`
	MaintenanceWindowStartTime *MQMaintenanceWindow `json:"maintenance_window_start_time,omitempty"`
	Logs *MQLogs `json:"logs,omitempty"`
	Configuration *MQConfiguration `json:"configuration,omitempty"`
}

type MQMaintenanceWindow struct {
	DayOfWeek string `json:"day_of_week"`
	TimeOfDay string `json:"time_of_day"`
	TimeZone  string `json:"time_zone"`
}

type MQLogs struct {
	Audit   bool `json:"audit"`
	General bool `json:"general"`
}

type MQConfiguration struct {
	ID       string `json:"id"`
	Revision int32  `json:"revision"`
}

func (r *MQResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_mq_broker" "%s" {
  broker_name = "%s"
  engine_type = "%s"
  engine_version = "%s"
  host_instance_type = "%s"
  deployment_mode = "%s"
  security_groups = %s
  subnet_ids = %s`,
		resourceName, r.BrokerName, r.EngineType, r.EngineVersion,
		r.HostInstanceType, r.DeploymentMode, formatStringSlice(r.SecurityGroups),
		formatStringSlice(r.SubnetIDs))

	if r.MaintenanceWindowStartTime != nil {
		content += fmt.Sprintf(`
  maintenance_window_start_time {
    day_of_week = "%s"
    time_of_day = "%s"
    time_zone = "%s"
  }`,
			r.MaintenanceWindowStartTime.DayOfWeek,
			r.MaintenanceWindowStartTime.TimeOfDay,
			r.MaintenanceWindowStartTime.TimeZone)
	}

	if r.Logs != nil {
		content += fmt.Sprintf(`
  logs {
    audit = %t
    general = %t
  }`,
			r.Logs.Audit, r.Logs.General)
	}

	if r.Configuration != nil {
		content += fmt.Sprintf(`
  configuration {
    id = "%s"
    revision = %d
  }`,
			r.Configuration.ID, r.Configuration.Revision)
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// GlueResource represents an AWS Glue catalog database
type GlueResource struct {
	BaseResource
	DatabaseName string `json:"database_name"`
	Description  string `json:"description,omitempty"`
	LocationURI  string `json:"location_uri,omitempty"`
	Parameters   map[string]string `json:"parameters,omitempty"`
}

func (r *GlueResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_glue_catalog_database" "%s" {
  name = "%s"`,
		resourceName, r.DatabaseName)

	if r.Description != "" {
		content += fmt.Sprintf(`
  description = "%s"`, r.Description)
	}

	if r.LocationURI != "" {
		content += fmt.Sprintf(`
  location_uri = "%s"`, r.LocationURI)
	}

	if len(r.Parameters) > 0 {
		content += `
  parameters = {`
		for k, v := range r.Parameters {
			quotedKey := k
			if strings.ContainsAny(k, ":-") {
				quotedKey = fmt.Sprintf(`"%s"`, k)
			}
			content += fmt.Sprintf(`
    %s = "%s"`, quotedKey, v)
		}
		content += `
  }`
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// AthenaResource represents an AWS Athena workgroup
type AthenaResource struct {
	BaseResource
	WorkgroupName string `json:"workgroup_name"`
	Description   string `json:"description,omitempty"`
	State         string `json:"state"`
	Configuration *AthenaWorkgroupConfiguration `json:"configuration,omitempty"`
}

type AthenaWorkgroupConfiguration struct {
	EnforceWorkgroupConfiguration    bool `json:"enforce_workgroup_configuration"`
	PublishCloudwatchMetricsEnabled  bool `json:"publish_cloudwatch_metrics_enabled"`
	BytesScannedCutoffPerQuery      int64 `json:"bytes_scanned_cutoff_per_query,omitempty"`
	RequesterPaysEnabled            bool `json:"requester_pays_enabled"`
	EngineVersion                   *AthenaEngineVersion `json:"engine_version,omitempty"`
	ResultConfiguration             *AthenaResultConfiguration `json:"result_configuration,omitempty"`
}

type AthenaEngineVersion struct {
	SelectedEngineVersion string `json:"selected_engine_version"`
	EffectiveEngineVersion string `json:"effective_engine_version"`
}

type AthenaResultConfiguration struct {
	OutputLocation string `json:"output_location"`
	EncryptionConfiguration *AthenaEncryptionConfiguration `json:"encryption_configuration,omitempty"`
}

type AthenaEncryptionConfiguration struct {
	EncryptionOption string `json:"encryption_option"`
	KMSKey           string `json:"kms_key,omitempty"`
}

func (r *AthenaResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_athena_workgroup" "%s" {
  name = "%s"
  state = "%s"`,
		resourceName, r.WorkgroupName, r.State)

	if r.Description != "" {
		content += fmt.Sprintf(`
  description = "%s"`, r.Description)
	}

	if r.Configuration != nil {
		content += fmt.Sprintf(`
  configuration {
    enforce_workgroup_configuration = %t
    publish_cloudwatch_metrics_enabled = %t
    requester_pays_enabled = %t`,
			r.Configuration.EnforceWorkgroupConfiguration,
			r.Configuration.PublishCloudwatchMetricsEnabled,
			r.Configuration.RequesterPaysEnabled)

		if r.Configuration.BytesScannedCutoffPerQuery > 0 {
			content += fmt.Sprintf(`
    bytes_scanned_cutoff_per_query = %d`, r.Configuration.BytesScannedCutoffPerQuery)
		}

		if r.Configuration.EngineVersion != nil {
			content += fmt.Sprintf(`
    engine_version {
      selected_engine_version = "%s"
    }`,
				r.Configuration.EngineVersion.SelectedEngineVersion)
		}

		if r.Configuration.ResultConfiguration != nil {
			content += fmt.Sprintf(`
    result_configuration {
      output_location = "%s"`,
				r.Configuration.ResultConfiguration.OutputLocation)

			if r.Configuration.ResultConfiguration.EncryptionConfiguration != nil {
				content += fmt.Sprintf(`
      encryption_configuration {
        encryption_option = "%s"`,
					r.Configuration.ResultConfiguration.EncryptionConfiguration.EncryptionOption)

				if r.Configuration.ResultConfiguration.EncryptionConfiguration.KMSKey != "" {
					content += fmt.Sprintf(`
        kms_key = "%s"`,
						r.Configuration.ResultConfiguration.EncryptionConfiguration.KMSKey)
				}
				content += `
      }`
			}
			content += `
    }`
		}
		content += `
  }`
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// QuickSightResource represents an AWS QuickSight user
type QuickSightResource struct {
	BaseResource
	UserName    string `json:"user_name"`
	Email       string `json:"email"`
	IdentityType string `json:"identity_type"`
	UserRole    string `json:"user_role"`
	Namespace   string `json:"namespace,omitempty"`
}

func (r *QuickSightResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_quicksight_user" "%s" {
  user_name = "%s"
  email = "%s"
  identity_type = "%s"
  user_role = "%s"`,
		resourceName, r.UserName, r.Email, r.IdentityType, r.UserRole)

	if r.Namespace != "" {
		content += fmt.Sprintf(`
  namespace = "%s"`, r.Namespace)
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// WorkSpacesResource represents an AWS WorkSpaces workspace
type WorkSpacesResource struct {
	BaseResource
	WorkspaceID    string `json:"workspace_id"`
	DirectoryID    string `json:"directory_id"`
	BundleID       string `json:"bundle_id"`
	UserName       string `json:"user_name"`
	RootVolumeSizeGib int32 `json:"root_volume_size_gib"`
	UserVolumeSizeGib int32 `json:"user_volume_size_gib"`
	ComputeTypeName string `json:"compute_type_name"`
	UserVolumeEncryptionEnabled bool `json:"user_volume_encryption_enabled"`
	RootVolumeEncryptionEnabled bool `json:"root_volume_encryption_enabled"`
	RunningMode string `json:"running_mode"`
	AutoStopTimeoutInMinutes int32 `json:"auto_stop_timeout_in_minutes,omitempty"`
}

func (r *WorkSpacesResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_workspaces_workspace" "%s" {
  directory_id = "%s"
  bundle_id = "%s"
  user_name = "%s"
  root_volume_size_gib = %d
  user_volume_size_gib = %d
  compute_type_name = "%s"
  user_volume_encryption_enabled = %t
  root_volume_encryption_enabled = %t
  running_mode = "%s"`,
		resourceName, r.DirectoryID, r.BundleID, r.UserName,
		r.RootVolumeSizeGib, r.UserVolumeSizeGib, r.ComputeTypeName,
		r.UserVolumeEncryptionEnabled, r.RootVolumeEncryptionEnabled, r.RunningMode)

	if r.AutoStopTimeoutInMinutes > 0 {
		content += fmt.Sprintf(`
  auto_stop_timeout_in_minutes = %d`, r.AutoStopTimeoutInMinutes)
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// StorageGatewayResource represents an AWS Storage Gateway gateway
type StorageGatewayResource struct {
	BaseResource
	GatewayName string `json:"gateway_name"`
	GatewayARN  string `json:"gateway_arn"`
	GatewayType string `json:"gateway_type"`
	GatewayTimezone string `json:"gateway_timezone"`
	GatewayRegion string `json:"gateway_region"`
	GatewayVPCEndpoint string `json:"gateway_vpc_endpoint,omitempty"`
	CloudWatchLogGroupARN string `json:"cloud_watch_log_group_arn,omitempty"`
	AverageDownloadRateLimitInBitsPerSec int64 `json:"average_download_rate_limit_in_bits_per_sec,omitempty"`
	AverageUploadRateLimitInBitsPerSec int64 `json:"average_upload_rate_limit_in_bits_per_sec,omitempty"`
}

func (r *StorageGatewayResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_storagegateway_gateway" "%s" {
  gateway_name = "%s"
  gateway_timezone = "%s"
  gateway_region = "%s"`,
		resourceName, r.GatewayName, r.GatewayTimezone, r.GatewayRegion)

	if r.GatewayVPCEndpoint != "" {
		content += fmt.Sprintf(`
  gateway_vpc_endpoint = "%s"`, r.GatewayVPCEndpoint)
	}

	if r.CloudWatchLogGroupARN != "" {
		content += fmt.Sprintf(`
  cloud_watch_log_group_arn = "%s"`, r.CloudWatchLogGroupARN)
	}

	if r.AverageDownloadRateLimitInBitsPerSec > 0 {
		content += fmt.Sprintf(`
  average_download_rate_limit_in_bits_per_sec = %d`, r.AverageDownloadRateLimitInBitsPerSec)
	}

	if r.AverageUploadRateLimitInBitsPerSec > 0 {
		content += fmt.Sprintf(`
  average_upload_rate_limit_in_bits_per_sec = %d`, r.AverageUploadRateLimitInBitsPerSec)
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// TransferResource represents an AWS Transfer server
type TransferResource struct {
	BaseResource
	ServerID string `json:"server_id"`
	IdentityProviderType string `json:"identity_provider_type"`
	LoggingRole string `json:"logging_role,omitempty"`
	Protocols []string `json:"protocols"`
	EndpointType string `json:"endpoint_type"`
	SecurityPolicyName string `json:"security_policy_name,omitempty"`
	WorkflowDetails *TransferWorkflowDetails `json:"workflow_details,omitempty"`
	StructuredLogDestinations []string `json:"structured_log_destinations,omitempty"`
}

type TransferWorkflowDetails struct {
	OnUpload []*TransferWorkflowDetail `json:"on_upload,omitempty"`
	OnPartialUpload []*TransferWorkflowDetail `json:"on_partial_upload,omitempty"`
}

type TransferWorkflowDetail struct {
	WorkflowID string `json:"workflow_id"`
	ExecutionRole string `json:"execution_role"`
}

func (r *TransferResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_transfer_server" "%s" {
  identity_provider_type = "%s"
  protocols = %s
  endpoint_type = "%s"`,
		resourceName, r.IdentityProviderType, formatStringSlice(r.Protocols), r.EndpointType)

	if r.LoggingRole != "" {
		content += fmt.Sprintf(`
  logging_role = "%s"`, r.LoggingRole)
	}

	if r.SecurityPolicyName != "" {
		content += fmt.Sprintf(`
  security_policy_name = "%s"`, r.SecurityPolicyName)
	}

	if r.WorkflowDetails != nil {
		content += `
  workflow_details {`
		
		if len(r.WorkflowDetails.OnUpload) > 0 {
			content += `
    on_upload {`
			for _, workflow := range r.WorkflowDetails.OnUpload {
				content += fmt.Sprintf(`
      workflow_id = "%s"
      execution_role = "%s"`,
					workflow.WorkflowID, workflow.ExecutionRole)
			}
			content += `
    }`
		}

		if len(r.WorkflowDetails.OnPartialUpload) > 0 {
			content += `
    on_partial_upload {`
			for _, workflow := range r.WorkflowDetails.OnPartialUpload {
				content += fmt.Sprintf(`
      workflow_id = "%s"
      execution_role = "%s"`,
					workflow.WorkflowID, workflow.ExecutionRole)
			}
			content += `
    }`
		}
		content += `
  }`
	}

	if len(r.StructuredLogDestinations) > 0 {
		content += fmt.Sprintf(`
  structured_log_destinations = %s`, formatStringSlice(r.StructuredLogDestinations))
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// MediaStoreResource represents an AWS MediaStore container
type MediaStoreResource struct {
	BaseResource
	ContainerName string `json:"container_name"`
	ContainerARN  string `json:"container_arn"`
	Status        string `json:"status"`
	AccessLoggingEnabled bool `json:"access_logging_enabled"`
	CorsPolicy    string `json:"cors_policy,omitempty"`
	LifecyclePolicy string `json:"lifecycle_policy,omitempty"`
	Policy        string `json:"policy,omitempty"`
}

func (r *MediaStoreResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_mediastore_container" "%s" {
  name = "%s"`,
		resourceName, r.ContainerName)

	if r.AccessLoggingEnabled {
		content += `
  access_logging_enabled = true`
	}

	if r.CorsPolicy != "" {
		content += fmt.Sprintf(`
  cors_policy = %s`, r.CorsPolicy)
	}

	if r.LifecyclePolicy != "" {
		content += fmt.Sprintf(`
  lifecycle_policy = %s`, r.LifecyclePolicy)
	}

	if r.Policy != "" {
		content += fmt.Sprintf(`
  policy = %s`, r.Policy)
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// MediaConvertResource represents an AWS MediaConvert queue
type MediaConvertResource struct {
	BaseResource
	QueueName string `json:"queue_name"`
	QueueARN  string `json:"queue_arn"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	PricingPlan string `json:"pricing_plan"`
	ReservationPlan *MediaConvertReservationPlan `json:"reservation_plan,omitempty"`
}

type MediaConvertReservationPlan struct {
	Commitment     string `json:"commitment"`
	ReservedSlots  int32  `json:"reserved_slots"`
	RenewalType    string `json:"renewal_type"`
}

func (r *MediaConvertResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_mediaconvert_queue" "%s" {
  name = "%s"
  type = "%s"
  pricing_plan = "%s"`,
		resourceName, r.QueueName, r.Type, r.PricingPlan)

	if r.ReservationPlan != nil {
		content += fmt.Sprintf(`
  reservation_plan {
    commitment = "%s"
    reserved_slots = %d
    renewal_type = "%s"
  }`,
			r.ReservationPlan.Commitment,
			r.ReservationPlan.ReservedSlots,
			r.ReservationPlan.RenewalType)
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// MediaLiveResource represents an AWS MediaLive channel
type MediaLiveResource struct {
	BaseResource
	ChannelID string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	ChannelARN string `json:"channel_arn"`
	State     string `json:"state"`
	ChannelClass string `json:"channel_class"`
	InputSpecification *MediaLiveInputSpecification `json:"input_specification,omitempty"`
	EncoderSettings *MediaLiveEncoderSettings `json:"encoder_settings,omitempty"`
}

type MediaLiveInputSpecification struct {
	Codec string `json:"codec"`
	Resolution string `json:"resolution"`
	MaximumBitrate string `json:"maximum_bitrate"`
}

type MediaLiveEncoderSettings struct {
	AudioDescriptions []*MediaLiveAudioDescription `json:"audio_descriptions,omitempty"`
	VideoDescriptions []*MediaLiveVideoDescription `json:"video_descriptions,omitempty"`
}

type MediaLiveAudioDescription struct {
	AudioSelectorName string `json:"audio_selector_name"`
	CodecSettings *MediaLiveAudioCodecSettings `json:"codec_settings,omitempty"`
}

type MediaLiveVideoDescription struct {
	Name string `json:"name"`
	CodecSettings *MediaLiveVideoCodecSettings `json:"codec_settings,omitempty"`
}

type MediaLiveAudioCodecSettings struct {
	AacSettings *MediaLiveAacSettings `json:"aac_settings,omitempty"`
}

type MediaLiveVideoCodecSettings struct {
	H264Settings *MediaLiveH264Settings `json:"h264_settings,omitempty"`
}

type MediaLiveAacSettings struct {
	Bitrate int32 `json:"bitrate"`
	CodingMode string `json:"coding_mode"`
	SampleRate int32 `json:"sample_rate"`
}

type MediaLiveH264Settings struct {
	Bitrate int32 `json:"bitrate"`
	FramerateDenominator int32 `json:"framerate_denominator"`
	FramerateNumerator int32 `json:"framerate_numerator"`
	Profile string `json:"profile"`
}

func (r *MediaLiveResource) ToOpenTofu() (string, error) {
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_medialive_channel" "%s" {
  name = "%s"
  channel_class = "%s"`,
		resourceName, r.ChannelName, r.ChannelClass)

	if r.InputSpecification != nil {
		content += fmt.Sprintf(`
  input_specification {
    codec = "%s"
    resolution = "%s"
    maximum_bitrate = "%s"
  }`,
			r.InputSpecification.Codec,
			r.InputSpecification.Resolution,
			r.InputSpecification.MaximumBitrate)
	}

	if r.EncoderSettings != nil {
		content += `
  encoder_settings {`
		
		if len(r.EncoderSettings.AudioDescriptions) > 0 {
			content += `
    audio_descriptions {`
			for _, audio := range r.EncoderSettings.AudioDescriptions {
				content += fmt.Sprintf(`
      audio_selector_name = "%s"`,
					audio.AudioSelectorName)
				
				if audio.CodecSettings != nil && audio.CodecSettings.AacSettings != nil {
					content += fmt.Sprintf(`
      codec_settings {
        aac_settings {
          bitrate = %d
          coding_mode = "%s"
          sample_rate = %d
        }
      }`,
						audio.CodecSettings.AacSettings.Bitrate,
						audio.CodecSettings.AacSettings.CodingMode,
						audio.CodecSettings.AacSettings.SampleRate)
				}
			}
			content += `
    }`
		}

		if len(r.EncoderSettings.VideoDescriptions) > 0 {
			content += `
    video_descriptions {`
			for _, video := range r.EncoderSettings.VideoDescriptions {
				content += fmt.Sprintf(`
      name = "%s"`,
					video.Name)
				
				if video.CodecSettings != nil && video.CodecSettings.H264Settings != nil {
					content += fmt.Sprintf(`
      codec_settings {
        h264_settings {
          bitrate = %d
          framerate_denominator = %d
          framerate_numerator = %d
          profile = "%s"
        }
      }`,
						video.CodecSettings.H264Settings.Bitrate,
						video.CodecSettings.H264Settings.FramerateDenominator,
						video.CodecSettings.H264Settings.FramerateNumerator,
						video.CodecSettings.H264Settings.Profile)
				}
			}
			content += `
    }`
		}
		content += `
  }`
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// APIGatewayResource represents an AWS API Gateway HTTP API
type APIGatewayResource struct {
	BaseResource
	APIID        string `json:"api_id"`
	Name         string `json:"name"`
	ProtocolType string `json:"protocol_type"`
	Version      string `json:"version,omitempty"`
}

func (r *APIGatewayResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_apigatewayv2_api" "%s" {
  name          = "%s"
  protocol_type = "%s"`,
		resourceName, r.Name, r.ProtocolType)

	if r.Version != "" {
		content += fmt.Sprintf(`
  version       = "%s"`,
			r.Version)
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// SSMParameterResource represents an AWS SSM Parameter
type SSMParameterResource struct {
	BaseResource
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Tier        string `json:"tier,omitempty"`
	Value       string `json:"value,omitempty"`
}

func (r *SSMParameterResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_ssm_parameter" "%s" {
  name  = "%s"
  type  = "%s"`,
		resourceName, r.Name, r.Type)

	if r.Description != "" {
		content += fmt.Sprintf(`
  description = "%s"`,
			r.Description)
	}

	if r.Tier != "" {
		content += fmt.Sprintf(`
  tier = "%s"`,
			r.Tier)
	}

	if r.Value != "" {
		content += fmt.Sprintf(`
  value = "%s"`,
			r.Value)
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// SecretsManagerResource represents an AWS Secrets Manager secret
type SecretsManagerResource struct {
	BaseResource
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	KMSKeyID    string `json:"kms_key_id,omitempty"`
}

func (r *SecretsManagerResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu using SecretsManager-specific sanitization
	resourceName := SanitizeSecretsManagerName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_secretsmanager_secret" "%s" {
  name = "%s"`,
		resourceName, SanitizeSecretsManagerName(r.Name))

	if r.Description != "" {
		content += fmt.Sprintf(`
  description = "%s"`,
			r.Description)
	}

	if r.KMSKeyID != "" {
		content += fmt.Sprintf(`
  kms_key_id = "%s"`,
			r.KMSKeyID)
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// KMSResource represents an AWS KMS key
type KMSResource struct {
	BaseResource
	KeyID           string `json:"key_id"`
	Description     string `json:"description,omitempty"`
	KeyUsage        string `json:"key_usage"`
	CustomerMasterKeySpec string `json:"customer_master_key_spec"`
	DeletionWindowInDays int32 `json:"deletion_window_in_days"`
	EnableKeyRotation bool `json:"enable_key_rotation"`
}

func (r *KMSResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_kms_key" "%s" {
  description             = "%s"
  deletion_window_in_days = %d
  enable_key_rotation     = %t`,
		resourceName, r.Description, r.DeletionWindowInDays, r.EnableKeyRotation)

	if r.KeyUsage != "" {
		content += fmt.Sprintf(`
  key_usage = "%s"`,
			r.KeyUsage)
	}

	if r.CustomerMasterKeySpec != "" {
		content += fmt.Sprintf(`
  customer_master_key_spec = "%s"`,
			r.CustomerMasterKeySpec)
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// CodeCommitResource represents an AWS CodeCommit repository
type CodeCommitResource struct {
	BaseResource
	RepositoryName string `json:"repository_name"`
	Description    string `json:"description,omitempty"`
	DefaultBranch  string `json:"default_branch,omitempty"`
}

func (r *CodeCommitResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_codecommit_repository" "%s" {
  repository_name = "%s"`,
		resourceName, r.RepositoryName)

	if r.Description != "" {
		content += fmt.Sprintf(`
  description = "%s"`,
			r.Description)
	}

	if r.DefaultBranch != "" {
		content += fmt.Sprintf(`
  default_branch = "%s"`,
			r.DefaultBranch)
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// CodePipelineResource represents an AWS CodePipeline
type CodePipelineResource struct {
	BaseResource
	PipelineName string `json:"pipeline_name"`
	RoleARN      string `json:"role_arn"`
}

func (r *CodePipelineResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_codepipeline" "%s" {
  name     = "%s"
  role_arn = "%s"`,
		resourceName, r.PipelineName, r.RoleARN)

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// CloudFormationResource represents an AWS CloudFormation stack
type CloudFormationResource struct {
	BaseResource
	StackName string `json:"stack_name"`
	TemplateBody string `json:"template_body,omitempty"`
	TemplateURL string `json:"template_url,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

func (r *CloudFormationResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_cloudformation_stack" "%s" {
  name = "%s"`,
		resourceName, r.StackName)

	if r.TemplateBody != "" {
		content += fmt.Sprintf(`
  template_body = <<EOF
%s
EOF`,
			r.TemplateBody)
	}

	if r.TemplateURL != "" {
		content += fmt.Sprintf(`
  template_url = "%s"`,
			r.TemplateURL)
	}

	if len(r.Capabilities) > 0 {
		content += `
  capabilities = [`
		for i, capability := range r.Capabilities {
			if i > 0 {
				content += ","
			}
			content += fmt.Sprintf(`"%s"`, capability)
		}
		content += `]`
	}

	if len(r.Parameters) > 0 {
		content += `
  parameters = {`
		for k, v := range r.Parameters {
			content += fmt.Sprintf(`
    %s = "%s"`, k, v)
		}
		content += `
  }`
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// FirehoseResource represents an AWS Kinesis Firehose delivery stream
type FirehoseResource struct {
	BaseResource
	Name                 string `json:"name"`
	DeliveryStreamType   string `json:"delivery_stream_type"`
	DeliveryStreamStatus string `json:"delivery_stream_status"`
}

func (r *FirehoseResource) GetARN() string {
	// Firehose streams don't have a specific ARN field, so we'll construct one
	return fmt.Sprintf("arn:aws:firehose:%s:123456789012:deliverystream/%s", r.Region, r.Name)
}

func (r *FirehoseResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_kinesis_firehose_delivery_stream" "%s" {
  name = "%s"
  destination = "extended_s3"
  extended_s3_configuration {
    bucket_arn = "arn:aws:s3:::your-bucket-name"
    role_arn   = "arn:aws:iam::123456789012:role/firehose-role"
  }`,
		resourceName, r.Name)

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

// CloudTrailResource represents an AWS CloudTrail trail
type CloudTrailResource struct {
	BaseResource
	Name                    string `json:"name"`
	S3BucketName            string `json:"s3_bucket_name"`
	S3KeyPrefix             string `json:"s3_key_prefix,omitempty"`
	CloudWatchLogGroupARN   string `json:"cloud_watch_log_group_arn,omitempty"`
	CloudWatchLogsRoleARN   string `json:"cloud_watch_logs_role_arn,omitempty"`
	IncludeGlobalServiceEvents bool `json:"include_global_service_events"`
	IsMultiRegionTrail      bool `json:"is_multi_region_trail"`
	EnableLogFileValidation bool `json:"enable_log_file_validation"`
	KMSKeyID                string `json:"kms_key_id,omitempty"`
	EventSelectors          []*CloudTrailEventSelector `json:"event_selectors,omitempty"`
	InsightSelectors        []*CloudTrailInsightSelector `json:"insight_selectors,omitempty"`
}

type CloudTrailEventSelector struct {
	ReadWriteType string `json:"read_write_type,omitempty"`
	IncludeManagementEvents bool `json:"include_management_events"`
	DataResources []*CloudTrailDataResource `json:"data_resources,omitempty"`
	ExcludeManagementEventSources []string `json:"exclude_management_event_sources,omitempty"`
}

type CloudTrailDataResource struct {
	Type   string   `json:"type"`
	Values []string `json:"values"`
}

type CloudTrailInsightSelector struct {
	InsightType string `json:"insight_type"`
}

func (r *CloudTrailResource) ToOpenTofu() (string, error) {
	// Sanitize the resource name for OpenTofu
	resourceName := SanitizeResourceName(r.Name)
	
	content := fmt.Sprintf(`resource "aws_cloudtrail" "%s" {
  name = "%s"
  s3_bucket_name = "%s"`,
		resourceName, r.Name, r.S3BucketName)

	if r.S3KeyPrefix != "" {
		content += fmt.Sprintf(`
  s3_key_prefix = "%s"`,
			r.S3KeyPrefix)
	}

	if r.CloudWatchLogGroupARN != "" {
		content += fmt.Sprintf(`
  cloud_watch_logs_group_arn = "%s"`,
			r.CloudWatchLogGroupARN)
	}

	if r.CloudWatchLogsRoleARN != "" {
		content += fmt.Sprintf(`
  cloud_watch_logs_role_arn = "%s"`,
			r.CloudWatchLogsRoleARN)
	}

	content += fmt.Sprintf(`
  include_global_service_events = %t
  is_multi_region_trail = %t
  enable_log_file_validation = %t`,
		r.IncludeGlobalServiceEvents, r.IsMultiRegionTrail, r.EnableLogFileValidation)

	if r.KMSKeyID != "" {
		content += fmt.Sprintf(`
  kms_key_id = "%s"`,
			r.KMSKeyID)
	}

	// Add event selectors if present
	if len(r.EventSelectors) > 0 {
		content += `
  event_selector {`
		for _, selector := range r.EventSelectors {
			if selector.ReadWriteType != "" {
				content += fmt.Sprintf(`
    read_write_type = "%s"`,
					selector.ReadWriteType)
			}
			content += fmt.Sprintf(`
    include_management_events = %t`,
				selector.IncludeManagementEvents)
			
			if len(selector.DataResources) > 0 {
				content += `
    data_resource {`
				for _, dataResource := range selector.DataResources {
					content += fmt.Sprintf(`
      type = "%s"`,
						dataResource.Type)
					if len(dataResource.Values) > 0 {
						content += `
      values = [`
						for i, value := range dataResource.Values {
							if i > 0 {
								content += ","
							}
							content += fmt.Sprintf(`"%s"`, value)
						}
						content += `]`
					}
				}
				content += `
    }`
			}
		}
		content += `
  }`
	}

	// Add insight selectors if present
	if len(r.InsightSelectors) > 0 {
		content += `
  insight_selector {`
		for _, selector := range r.InsightSelectors {
			content += fmt.Sprintf(`
    insight_type = "%s"`,
				selector.InsightType)
		}
		content += `
  }`
	}

	// Add tags
	content += `
  tags = {`
	if len(r.Tags) > 0 {
		for k, v := range r.Tags {
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

