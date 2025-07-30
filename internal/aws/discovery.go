package aws

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/eks"
)

// Helper function to decode URL-encoded JSON
func decodeURLEncodedJSON(encoded string) string {
	decoded, err := url.QueryUnescape(encoded)
	if err != nil {
		// If decoding fails, return the original string
		return encoded
	}
	return decoded
}

// Helper function to check if a resource has already been seen
func isResourceSeen(seenResources map[string]bool, resourceID string) bool {
	if seenResources[resourceID] {
		return true
	}
	seenResources[resourceID] = true
	return false
}

// Helper function to convert EC2 tags to map
func convertTags(tags []types.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		if tag.Key != nil && tag.Value != nil {
			result[*tag.Key] = *tag.Value
		}
	}
	return result
}

// discoverVPCResources discovers VPC resources
func (c *Client) discoverVPCResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.ec2Client.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe VPCs: %w", err)
	}

	for _, vpc := range result.Vpcs {
		// Get VPC name from tags
		name := *vpc.VpcId
		if vpc.Tags != nil {
			for _, tag := range vpc.Tags {
				if tag.Key != nil && *tag.Key == "Name" && tag.Value != nil {
					name = *tag.Value
					break
				}
			}
		}

		resource := &VPCResource{
			BaseResource: BaseResource{
				Type:   "vpc",
				ID:     *vpc.VpcId,
				Name:   name,
				Region: c.region,
				Tags:   convertTags(vpc.Tags),
			},
			CIDRBlock:          *vpc.CidrBlock,
			EnableDNSHostnames: true, // Default value, would need separate API call
			EnableDNSSupport:   true, // Default value, would need separate API call
			InstanceTenancy:    string(vpc.InstanceTenancy),
		}

		// Add IPv6 CIDR block if present
		if vpc.Ipv6CidrBlockAssociationSet != nil && len(vpc.Ipv6CidrBlockAssociationSet) > 0 {
			for _, ipv6Assoc := range vpc.Ipv6CidrBlockAssociationSet {
				if ipv6Assoc.Ipv6CidrBlock != nil {
					resource.IPv6CIDRBlock = *ipv6Assoc.Ipv6CidrBlock
					break
				}
			}
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverEC2Resources discovers EC2 instance resources
func (c *Client) discoverEC2Resources() ([]Resource, error) {
	var resources []Resource

	result, err := c.ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe EC2 instances: %w", err)
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			// Get instance name from tags
			name := *instance.InstanceId
			if instance.Tags != nil {
				for _, tag := range instance.Tags {
					if tag.Key != nil && *tag.Key == "Name" && tag.Value != nil {
						name = *tag.Value
						break
					}
				}
			}

			// Get security groups
			var securityGroups []string
			if instance.SecurityGroups != nil {
				for _, sg := range instance.SecurityGroups {
					if sg.GroupId != nil {
						securityGroups = append(securityGroups, *sg.GroupId)
					}
				}
			}

			// Get root block device
			var rootBlockDevice *BlockDeviceMapping
			if instance.BlockDeviceMappings != nil {
				for _, bdm := range instance.BlockDeviceMappings {
					if bdm.DeviceName != nil && *bdm.DeviceName == "/dev/xvda" {
						if bdm.Ebs != nil {
							rootBlockDevice = &BlockDeviceMapping{
								DeviceName: *bdm.DeviceName,
								VolumeSize: 20, // Default size, would need separate API call for actual size
								VolumeType: "gp2", // Default type
								Encrypted:  false, // Default value
							}
							// Note: EBS block device details would need separate API calls
						}
						break
					}
				}
			}

			resource := &EC2InstanceResource{
				BaseResource: BaseResource{
					Type:   "ec2",
					ID:     *instance.InstanceId,
					Name:   name,
					Region: c.region,
					Tags:   convertTags(instance.Tags),
				},
				InstanceType: string(instance.InstanceType),
				AMI:          *instance.ImageId,
				SubnetID:     *instance.SubnetId,
				SecurityGroups: securityGroups,
				RootBlockDevice: rootBlockDevice,
			}

			// Add optional fields
			if instance.KeyName != nil {
				resource.KeyName = *instance.KeyName
			}
			if instance.IamInstanceProfile != nil && instance.IamInstanceProfile.Arn != nil {
				resource.IAMInstanceProfile = *instance.IamInstanceProfile.Arn
			}
			// Note: UserData would need separate API call to retrieve

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// discoverIAMResources discovers IAM role resources
func (c *Client) discoverIAMResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.iamClient.ListRoles(context.TODO(), &iam.ListRolesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list IAM roles: %w", err)
	}

	for _, role := range result.Roles {
		// Get role details
		roleDetails, err := c.iamClient.GetRole(context.TODO(), &iam.GetRoleInput{
			RoleName: role.RoleName,
		})
		if err != nil {
			continue
		}

		// Get managed policies
		attachedPolicies, err := c.iamClient.ListAttachedRolePolicies(context.TODO(), &iam.ListAttachedRolePoliciesInput{
			RoleName: role.RoleName,
		})
		if err != nil {
			continue
		}

		var managedPolicyARNs []string
		for _, policy := range attachedPolicies.AttachedPolicies {
			managedPolicyARNs = append(managedPolicyARNs, *policy.PolicyArn)
		}

		// Get inline policies
		inlinePolicies, err := c.iamClient.ListRolePolicies(context.TODO(), &iam.ListRolePoliciesInput{
			RoleName: role.RoleName,
		})
		if err != nil {
			continue
		}

		inlinePolicyMap := make(map[string]string)
		for _, policyName := range inlinePolicies.PolicyNames {
			policyDoc, err := c.iamClient.GetRolePolicy(context.TODO(), &iam.GetRolePolicyInput{
				RoleName:   role.RoleName,
				PolicyName: &policyName,
			})
			if err == nil && policyDoc.PolicyDocument != nil {
				// Decode URL-encoded policy document
				decodedPolicy := decodeURLEncodedJSON(*policyDoc.PolicyDocument)
				inlinePolicyMap[policyName] = decodedPolicy
			}
		}

		// Convert tags
		tags := make(map[string]string)
		if roleDetails.Role.Tags != nil {
			for _, tag := range roleDetails.Role.Tags {
				if tag.Key != nil && tag.Value != nil {
					tags[*tag.Key] = *tag.Value
				}
			}
		}

		// Decode assume role policy
		var assumeRolePolicy string
		if roleDetails.Role.AssumeRolePolicyDocument != nil {
			assumeRolePolicy = decodeURLEncodedJSON(*roleDetails.Role.AssumeRolePolicyDocument)
		}

		resource := &IAMRoleResource{
			BaseResource: BaseResource{
				Type:   "iam",
				ID:     *role.RoleId,
				Name:   *role.RoleName,
				Region: c.region,
				Tags:   tags,
			},
			AssumeRolePolicy: assumeRolePolicy,
			ManagedPolicyARNs: managedPolicyARNs,
			InlinePolicies:    inlinePolicyMap,
			Path:              *role.Path,
			Description:       "",
		}

		if role.Description != nil {
			resource.Description = *role.Description
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverRDSResources discovers RDS instance resources
func (c *Client) discoverRDSResources() ([]Resource, error) {
	var resources []Resource
	seenResources := make(map[string]bool)

	result, err := c.rdsClient.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe RDS instances: %w", err)
	}

	for _, instance := range result.DBInstances {
		// Use DBInstanceIdentifier as the unique identifier to prevent duplicates
		resourceID := *instance.DBInstanceIdentifier
		
		// Skip if we've already seen this resource
		if isResourceSeen(seenResources, resourceID) {
			continue
		}

		// Convert tags
		tags := make(map[string]string)
		if instance.TagList != nil {
			for _, tag := range instance.TagList {
				if tag.Key != nil && tag.Value != nil {
					tags[*tag.Key] = *tag.Value
				}
			}
		}

		// Get VPC security groups
		var vpcSecurityGroupIDs []string
		if instance.VpcSecurityGroups != nil {
			for _, sg := range instance.VpcSecurityGroups {
				if sg.VpcSecurityGroupId != nil {
					vpcSecurityGroupIDs = append(vpcSecurityGroupIDs, *sg.VpcSecurityGroupId)
				}
			}
		}

		// Safely get values with nil checks
		engine := ""
		if instance.Engine != nil {
			engine = *instance.Engine
		}
		
		engineVersion := ""
		if instance.EngineVersion != nil {
			engineVersion = *instance.EngineVersion
		}
		
		instanceClass := ""
		if instance.DBInstanceClass != nil {
			instanceClass = *instance.DBInstanceClass
		}
		
		allocatedStorage := 0
		if instance.AllocatedStorage != nil {
			allocatedStorage = int(*instance.AllocatedStorage)
		}
		
		storageType := "gp2"
		if instance.StorageType != nil {
			storageType = *instance.StorageType
		}
		
		storageEncrypted := false
		if instance.StorageEncrypted != nil {
			storageEncrypted = *instance.StorageEncrypted
		}
		
		dbName := ""
		if instance.DBName != nil {
			dbName = *instance.DBName
		}
		
		username := ""
		if instance.MasterUsername != nil {
			username = *instance.MasterUsername
		}
		
		dbSubnetGroupName := ""
		if instance.DBSubnetGroup != nil && instance.DBSubnetGroup.DBSubnetGroupName != nil {
			dbSubnetGroupName = *instance.DBSubnetGroup.DBSubnetGroupName
		}
		
		backupRetentionPeriod := 0
		if instance.BackupRetentionPeriod != nil {
			backupRetentionPeriod = int(*instance.BackupRetentionPeriod)
		}
		
		multiAZ := false
		if instance.MultiAZ != nil {
			multiAZ = *instance.MultiAZ
		}
		
		publiclyAccessible := false
		if instance.PubliclyAccessible != nil {
			publiclyAccessible = *instance.PubliclyAccessible
		}

		resource := &RDSInstanceResource{
			BaseResource: BaseResource{
				Type:   "rds",
				ID:     *instance.DBInstanceIdentifier,
				Name:   *instance.DBInstanceIdentifier,
				Region: c.region,
				Tags:   tags,
			},
			Engine:               engine,
			EngineVersion:        engineVersion,
			InstanceClass:        instanceClass,
			AllocatedStorage:     allocatedStorage,
			StorageType:          storageType,
			StorageEncrypted:     storageEncrypted,
			DBName:               dbName,
			Username:             username,
			Password:             "CHANGE_ME", // Will be replaced with variable
			VpcSecurityGroupIDs:  vpcSecurityGroupIDs,
			DBSubnetGroupName:    dbSubnetGroupName,
			BackupRetentionPeriod: backupRetentionPeriod,
			MultiAZ:              multiAZ,
			PubliclyAccessible:   publiclyAccessible,
			SkipFinalSnapshot:    false,
		}

		// Add KMS key if present
		if instance.KmsKeyId != nil {
			resource.KMSKeyID = *instance.KmsKeyId
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverS3Resources discovers S3 bucket resources
func (c *Client) discoverS3Resources() ([]Resource, error) {
	var resources []Resource

	result, err := c.s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 buckets: %w", err)
	}

	for _, bucket := range result.Buckets {
		// Get bucket location
		location, err := c.s3Client.GetBucketLocation(context.TODO(), &s3.GetBucketLocationInput{
			Bucket: bucket.Name,
		})
		if err != nil {
			continue
		}

		// Only include buckets in the current region
		if location.LocationConstraint != "" && string(location.LocationConstraint) != c.region {
			continue
		}

		// Get bucket versioning
		versioning, err := c.s3Client.GetBucketVersioning(context.TODO(), &s3.GetBucketVersioningInput{
			Bucket: bucket.Name,
		})
		if err != nil {
			continue
		}

		// Get bucket encryption
		encryption, err := c.s3Client.GetBucketEncryption(context.TODO(), &s3.GetBucketEncryptionInput{
			Bucket: bucket.Name,
		})
		if err != nil {
			continue
		}

		// Get bucket tags
		tags, err := c.s3Client.GetBucketTagging(context.TODO(), &s3.GetBucketTaggingInput{
			Bucket: bucket.Name,
		})
		if err != nil {
			continue
		}

		tagMap := make(map[string]string)
		if tags.TagSet != nil {
			for _, tag := range tags.TagSet {
				if tag.Key != nil && tag.Value != nil {
					tagMap[*tag.Key] = *tag.Value
				}
			}
		}

		resource := &S3BucketResource{
			BaseResource: BaseResource{
				Type:   "s3",
				ID:     *bucket.Name,
				Name:   *bucket.Name,
				Region: c.region,
				Tags:   tagMap,
			},
			VersioningEnabled: versioning.Status == "Enabled",
			EncryptionEnabled: encryption.ServerSideEncryptionConfiguration != nil,
		}

		// Add KMS key if present
		if encryption.ServerSideEncryptionConfiguration != nil {
			for _, rule := range encryption.ServerSideEncryptionConfiguration.Rules {
				if rule.ApplyServerSideEncryptionByDefault != nil && rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID != nil {
					resource.KMSKeyID = *rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID
					break
				}
			}
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverLoadBalancerResources discovers Application Load Balancer resources
func (c *Client) discoverLoadBalancerResources() ([]Resource, error) {
	var resources []Resource
	seenResources := make(map[string]bool)

	result, err := c.elbv2Client.DescribeLoadBalancers(context.TODO(), &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe load balancers: %w", err)
	}

	for _, lb := range result.LoadBalancers {
		// Use LoadBalancerName as the unique identifier to prevent duplicates
		resourceID := *lb.LoadBalancerName
		
		// Skip if we've already seen this resource
		if isResourceSeen(seenResources, resourceID) {
			continue
		}

		// Convert tags - LoadBalancer doesn't have Tags field in this version
		tags := make(map[string]string)
		// Tags would need to be fetched separately if needed

		// Get security groups
		var securityGroups []string
		if lb.SecurityGroups != nil {
			for _, sg := range lb.SecurityGroups {
				securityGroups = append(securityGroups, sg)
			}
		}

		// Get subnets
		var subnets []string
		if lb.AvailabilityZones != nil {
			for _, az := range lb.AvailabilityZones {
				if az.SubnetId != nil {
					subnets = append(subnets, *az.SubnetId)
				}
			}
		}

		resource := &LoadBalancerResource{
			BaseResource: BaseResource{
				Type:   "alb",
				ID:     *lb.LoadBalancerArn,
				Name:   *lb.LoadBalancerName,
				Region: c.region,
				Tags:   tags,
			},
			Internal:           lb.Scheme == "internal",
			LoadBalancerType:   string(lb.Type),
			SecurityGroups:     securityGroups,
			Subnets:            subnets,
			EnableDeletionProtection: false, // Default value, would need separate API call
			IdleTimeout:        60, // Default value
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverAutoScalingResources discovers Auto Scaling Group resources
func (c *Client) discoverAutoScalingResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.asgClient.DescribeAutoScalingGroups(context.TODO(), &autoscaling.DescribeAutoScalingGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe Auto Scaling Groups: %w", err)
	}

	for _, asg := range result.AutoScalingGroups {
		// Convert tags
		tags := make(map[string]string)
		if asg.Tags != nil {
			for _, tag := range asg.Tags {
				if tag.Key != nil && tag.Value != nil {
					tags[*tag.Key] = *tag.Value
				}
			}
		}

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "asg",
				ID:     *asg.AutoScalingGroupName,
				Name:   *asg.AutoScalingGroupName,
				Region: c.region,
				Tags:   tags,
			},
			ResourceType: "aws_autoscaling_group",
			Attributes: map[string]interface{}{
				"name":                *asg.AutoScalingGroupName,
				"max_size":            asg.MaxSize,
				"min_size":            asg.MinSize,
				"desired_capacity":    asg.DesiredCapacity,
				"health_check_type":   asg.HealthCheckType,
				"health_check_grace_period": asg.HealthCheckGracePeriod,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverLambdaResources discovers Lambda function resources
func (c *Client) discoverLambdaResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.lambdaClient.ListFunctions(context.TODO(), &lambda.ListFunctionsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Lambda functions: %w", err)
	}

	for _, function := range result.Functions {
		// Convert tags - Lambda tags would need separate API call
		tags := make(map[string]string)

		resource := &LambdaFunctionResource{
			BaseResource: BaseResource{
				Type:   "lambda",
				ID:     *function.FunctionName,
				Name:   *function.FunctionName,
				Region: c.region,
				Tags:   tags,
			},
			Runtime:    string(function.Runtime),
			Handler:    *function.Handler,
			Role:       *function.Role,
			Code:       "lambda_function.zip", // Placeholder
			Timeout:    int(*function.Timeout),
			MemorySize: int(*function.MemorySize),
		}

		// Add environment variables if present
		if function.Environment != nil && function.Environment.Variables != nil {
			resource.Environment = function.Environment.Variables
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverSQSResources discovers SQS queue resources
func (c *Client) discoverSQSResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.sqsClient.ListQueues(context.TODO(), &sqs.ListQueuesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list SQS queues: %w", err)
	}

	for _, queueURL := range result.QueueUrls {
		// Get queue attributes
		attributes, err := c.sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
			QueueUrl: &queueURL,
			AttributeNames: []sqsTypes.QueueAttributeName{sqsTypes.QueueAttributeNameAll},
		})
		if err != nil {
			continue
		}

		// Extract queue name from URL
		parts := strings.Split(queueURL, "/")
		queueName := parts[len(parts)-1]

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "sqs",
				ID:     queueName,
				Name:   queueName,
				Region: c.region,
				Tags:   make(map[string]string), // SQS tags would need separate call
			},
			ResourceType: "aws_sqs_queue",
			Attributes: map[string]interface{}{
				"name": queueName,
				"url":  queueURL,
			},
		}

		// Add attributes if available
		if attributes.Attributes != nil {
			for k, v := range attributes.Attributes {
				resource.Attributes[k] = v
			}
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverSNSResources discovers SNS topic resources
func (c *Client) discoverSNSResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.snsClient.ListTopics(context.TODO(), &sns.ListTopicsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list SNS topics: %w", err)
	}

	for _, topic := range result.Topics {
		// Extract topic name from ARN
		parts := strings.Split(*topic.TopicArn, ":")
		topicName := parts[len(parts)-1]

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "sns",
				ID:     topicName,
				Name:   topicName,
				Region: c.region,
				Tags:   make(map[string]string), // SNS tags would need separate call
			},
			ResourceType: "aws_sns_topic",
			Attributes: map[string]interface{}{
				"name": topicName,
				"arn":  *topic.TopicArn,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverCloudWatchResources discovers CloudWatch resources
func (c *Client) discoverCloudWatchResources() ([]Resource, error) {
	var resources []Resource

	// Discover CloudWatch alarms
	alarms, err := c.cloudwatchClient.DescribeAlarms(context.TODO(), &cloudwatch.DescribeAlarmsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe CloudWatch alarms: %w", err)
	}

	for _, alarm := range alarms.MetricAlarms {
		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "cloudwatch",
				ID:     *alarm.AlarmName,
				Name:   *alarm.AlarmName,
				Region: c.region,
				Tags:   make(map[string]string), // CloudWatch tags would need separate call
			},
			ResourceType: "aws_cloudwatch_metric_alarm",
			Attributes: map[string]interface{}{
				"alarm_name":          *alarm.AlarmName,
				"comparison_operator": alarm.ComparisonOperator,
				"evaluation_periods":  alarm.EvaluationPeriods,
				"metric_name":         alarm.MetricName,
				"namespace":           alarm.Namespace,
				"period":              alarm.Period,
				"statistic":           alarm.Statistic,
				"threshold":           alarm.Threshold,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverCloudTrailResources discovers CloudTrail resources
func (c *Client) discoverCloudTrailResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.cloudtrailClient.ListTrails(context.TODO(), &cloudtrail.ListTrailsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list CloudTrail trails: %w", err)
	}

	for _, trail := range result.Trails {
		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "cloudtrail",
				ID:     *trail.Name,
				Name:   *trail.Name,
				Region: c.region,
				Tags:   make(map[string]string), // CloudTrail tags would need separate call
			},
			ResourceType: "aws_cloudtrail",
			Attributes: map[string]interface{}{
				"name": *trail.Name,
				"s3_bucket_name": "unknown", // Would need separate API call to get S3 bucket name
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverECSResources discovers ECS resources
func (c *Client) discoverECSResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.ecsClient.ListServices(context.TODO(), &ecs.ListServicesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ECS services: %w", err)
	}

	for _, serviceARN := range result.ServiceArns {
		// Get service details
		serviceDetails, err := c.ecsClient.DescribeServices(context.TODO(), &ecs.DescribeServicesInput{
			Services: []string{serviceARN},
		})
		if err != nil || len(serviceDetails.Services) == 0 {
			continue
		}

		service := serviceDetails.Services[0]

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "ecs",
				ID:     *service.ServiceName,
				Name:   *service.ServiceName,
				Region: c.region,
				Tags:   make(map[string]string), // ECS tags would need separate call
			},
			ResourceType: "aws_ecs_service",
			Attributes: map[string]interface{}{
				"name":            *service.ServiceName,
				"cluster":         service.ClusterArn,
				"task_definition": service.TaskDefinition,
				"desired_count":   service.DesiredCount,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// discoverEKSResources discovers EKS cluster resources
func (c *Client) discoverEKSResources() ([]Resource, error) {
	var resources []Resource

	result, err := c.eksClient.ListClusters(context.TODO(), &eks.ListClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list EKS clusters: %w", err)
	}

	for _, clusterName := range result.Clusters {
		// Get cluster details
		clusterDetails, err := c.eksClient.DescribeCluster(context.TODO(), &eks.DescribeClusterInput{
			Name: &clusterName,
		})
		if err != nil {
			continue
		}

		cluster := clusterDetails.Cluster

		resource := &GenericResource{
			BaseResource: BaseResource{
				Type:   "eks",
				ID:     *cluster.Name,
				Name:   *cluster.Name,
				Region: c.region,
				Tags:   cluster.Tags,
			},
			ResourceType: "aws_eks_cluster",
			Attributes: map[string]interface{}{
				"name":     *cluster.Name,
				"version":  cluster.Version,
				"platform_version": cluster.PlatformVersion,
				"endpoint": cluster.Endpoint,
			},
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// Stub functions for other services (to be implemented as needed)
func (c *Client) discoverElastiCacheResources() ([]Resource, error) {
	// TODO: Implement ElastiCache discovery
	return []Resource{}, nil
}

func (c *Client) discoverDynamoDBResources() ([]Resource, error) {
	// TODO: Implement DynamoDB discovery
	return []Resource{}, nil
}

func (c *Client) discoverRedshiftResources() ([]Resource, error) {
	// TODO: Implement Redshift discovery
	return []Resource{}, nil
}

func (c *Client) discoverRoute53Resources() ([]Resource, error) {
	// TODO: Implement Route53 discovery
	return []Resource{}, nil
}

// Additional stub functions for other AWS services
func (c *Client) discoverCloudFrontResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverAPIGatewayResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverElasticsearchResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverNeptuneResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverDocDBResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverElasticBeanstalkResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverECRResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverCodeCommitResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverCodeBuildResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverCodeDeployResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverCodePipelineResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverCloudFormationResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverSSMResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverSecretsManagerResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverKMSResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverGuardDutyResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverConfigResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverBackupResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverGlacierResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverGlueResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverAthenaResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverQuickSightResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverWorkSpacesResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverDirectoryServiceResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverFSxResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverStorageGatewayResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverTransferResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverMQResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverKinesisResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverFirehoseResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverMediaStoreResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverMediaConvertResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverMediaLiveResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverMediaTailorResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverIoTResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverGreengrassResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverGreengrassV2Resources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverIoTAnalyticsResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverIoTEventsResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverIoTSiteWiseResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverIoTThingsGraphResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverIoTWirelessResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverIoTDeviceAdvisorResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverIoTFleetHubResources() ([]Resource, error) { return []Resource{}, nil }
func (c *Client) discoverIoTSecureTunnelingResources() ([]Resource, error) { return []Resource{}, nil } 