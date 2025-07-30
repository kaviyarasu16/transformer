package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/codecommit"
	"github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	"github.com/aws/aws-sdk-go-v2/service/glacier"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	"github.com/aws/aws-sdk-go-v2/service/iotanalytics"
	"github.com/aws/aws-sdk-go-v2/service/iotevents"
	"github.com/aws/aws-sdk-go-v2/service/iotsitewise"
	"github.com/aws/aws-sdk-go-v2/service/iotthingsgraph"
	"github.com/aws/aws-sdk-go-v2/service/iotwireless"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert"
	"github.com/aws/aws-sdk-go-v2/service/medialive"
	"github.com/aws/aws-sdk-go-v2/service/mediastore"
	"github.com/aws/aws-sdk-go-v2/service/mediatailor"
	"github.com/aws/aws-sdk-go-v2/service/mq"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/storagegateway"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"

)

// Client represents an AWS client for resource discovery
type Client struct {
	region              string
	cfg                 aws.Config
	ec2Client           *ec2.Client
	iamClient           *iam.Client
	rdsClient           *rds.Client
	s3Client            *s3.Client
	elbv2Client         *elasticloadbalancingv2.Client
	asgClient           *autoscaling.Client
	lambdaClient        *lambda.Client
	sqsClient           *sqs.Client
	snsClient           *sns.Client
	cloudwatchClient    *cloudwatch.Client
	cloudtrailClient    *cloudtrail.Client
	ecsClient           *ecs.Client
	eksClient           *eks.Client
	dynamodbClient      *dynamodb.Client
	elasticacheClient   *elasticache.Client
	redshiftClient      *redshift.Client
	route53Client       *route53.Client
	cloudfrontClient    *cloudfront.Client
	apigatewayClient    *apigatewayv2.Client
	elasticsearchClient *elasticsearchservice.Client
	ecrClient           *ecr.Client
	codebuildClient     *codebuild.Client
	codedeployClient    *codedeploy.Client
	codepipelineClient  *codepipeline.Client
	cloudformationClient *cloudformation.Client
	ssmClient           *ssm.Client
	secretsmanagerClient *secretsmanager.Client
	kmsClient           *kms.Client
	kinesisClient       *kinesis.Client
	fsxClient           *fsx.Client
	configClient        *configservice.Client
	// Additional clients for remaining services
	codecommitClient    *codecommit.Client
	elasticbeanstalkClient *elasticbeanstalk.Client
	guarddutyClient     *guardduty.Client
	backupClient        *backup.Client
	glacierClient       *glacier.Client
	glueClient          *glue.Client
	athenaClient        *athena.Client
	quicksightClient    *quicksight.Client
	workspacesClient    *workspaces.Client
	storagegatewayClient *storagegateway.Client
	transferClient      *transfer.Client
	mqClient            *mq.Client
	firehoseClient      *firehose.Client
	mediastoreClient    *mediastore.Client
	mediaconvertClient  *mediaconvert.Client
	medialiveClient     *medialive.Client
	mediatailorClient   *mediatailor.Client
	iotClient           *iot.Client
	iotanalyticsClient  *iotanalytics.Client
	ioteventsClient     *iotevents.Client
	iotsitewiseClient   *iotsitewise.Client
	iotthingsgraphClient *iotthingsgraph.Client
	iotwirelessClient   *iotwireless.Client
}

// NewClient creates a new AWS client
func NewClient(region string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Client{
		region:              region,
		cfg:                 cfg,
		ec2Client:           ec2.NewFromConfig(cfg),
		iamClient:           iam.NewFromConfig(cfg),
		rdsClient:           rds.NewFromConfig(cfg),
		s3Client:            s3.NewFromConfig(cfg),
		elbv2Client:         elasticloadbalancingv2.NewFromConfig(cfg),
		asgClient:           autoscaling.NewFromConfig(cfg),
		lambdaClient:        lambda.NewFromConfig(cfg),
		sqsClient:           sqs.NewFromConfig(cfg),
		snsClient:           sns.NewFromConfig(cfg),
		cloudwatchClient:    cloudwatch.NewFromConfig(cfg),
		cloudtrailClient:    cloudtrail.NewFromConfig(cfg),
		ecsClient:           ecs.NewFromConfig(cfg),
		eksClient:           eks.NewFromConfig(cfg),
		dynamodbClient:      dynamodb.NewFromConfig(cfg),
		elasticacheClient:   elasticache.NewFromConfig(cfg),
		redshiftClient:      redshift.NewFromConfig(cfg),
		route53Client:       route53.NewFromConfig(cfg),
		cloudfrontClient:    cloudfront.NewFromConfig(cfg),
		apigatewayClient:    apigatewayv2.NewFromConfig(cfg),
		elasticsearchClient: elasticsearchservice.NewFromConfig(cfg),
		ecrClient:           ecr.NewFromConfig(cfg),
		codebuildClient:     codebuild.NewFromConfig(cfg),
		codedeployClient:    codedeploy.NewFromConfig(cfg),
		codepipelineClient:  codepipeline.NewFromConfig(cfg),
		cloudformationClient: cloudformation.NewFromConfig(cfg),
		ssmClient:           ssm.NewFromConfig(cfg),
		secretsmanagerClient: secretsmanager.NewFromConfig(cfg),
		kmsClient:           kms.NewFromConfig(cfg),
		kinesisClient:       kinesis.NewFromConfig(cfg),
		fsxClient:           fsx.NewFromConfig(cfg),
		configClient:        configservice.NewFromConfig(cfg),
		// Initialize additional clients
		codecommitClient:    codecommit.NewFromConfig(cfg),
		elasticbeanstalkClient: elasticbeanstalk.NewFromConfig(cfg),
		guarddutyClient:     guardduty.NewFromConfig(cfg),
		backupClient:        backup.NewFromConfig(cfg),
		glacierClient:       glacier.NewFromConfig(cfg),
		glueClient:          glue.NewFromConfig(cfg),
		athenaClient:        athena.NewFromConfig(cfg),
		quicksightClient:    quicksight.NewFromConfig(cfg),
		workspacesClient:    workspaces.NewFromConfig(cfg),
		storagegatewayClient: storagegateway.NewFromConfig(cfg),
		transferClient:      transfer.NewFromConfig(cfg),
		mqClient:            mq.NewFromConfig(cfg),
		firehoseClient:      firehose.NewFromConfig(cfg),
		mediastoreClient:    mediastore.NewFromConfig(cfg),
		mediaconvertClient:  mediaconvert.NewFromConfig(cfg),
		medialiveClient:     medialive.NewFromConfig(cfg),
		mediatailorClient:   mediatailor.NewFromConfig(cfg),
		iotClient:           iot.NewFromConfig(cfg),
		iotanalyticsClient:  iotanalytics.NewFromConfig(cfg),
		ioteventsClient:     iotevents.NewFromConfig(cfg),
		iotsitewiseClient:   iotsitewise.NewFromConfig(cfg),
		iotthingsgraphClient: iotthingsgraph.NewFromConfig(cfg),
		iotwirelessClient:   iotwireless.NewFromConfig(cfg),
	}, nil
}

// DiscoverResources discovers AWS resources of the specified types
func (c *Client) DiscoverResources(resourceTypes []string) ([]Resource, error) {
	var allResources []Resource

	for _, resourceType := range resourceTypes {
		resources, err := c.discoverResourceType(resourceType)
		if err != nil {
			return nil, fmt.Errorf("failed to discover %s resources: %w", resourceType, err)
		}
		allResources = append(allResources, resources...)
	}

	return allResources, nil
}

// discoverResourceType discovers resources of a specific type
func (c *Client) discoverResourceType(resourceType string) ([]Resource, error) {
	switch resourceType {
	case "vpc":
		return c.discoverVPCResources()
	case "ec2":
		return c.discoverEC2Resources()
	case "iam":
		return c.discoverIAMResources()
	case "rds":
		return c.discoverRDSResources()
	case "s3":
		return c.discoverS3Resources()
	case "alb", "elb":
		return c.discoverLoadBalancerResources()
	case "asg":
		return c.discoverAutoScalingResources()
	case "lambda":
		return c.discoverLambdaResources()
	case "sqs":
		return c.discoverSQSResources()
	case "sns":
		return c.discoverSNSResources()
	case "cloudwatch":
		return c.discoverCloudWatchResources()
	case "cloudtrail":
		return c.discoverCloudTrailResources()
	case "ecs":
		return c.discoverECSResources()
	case "eks":
		return c.discoverEKSResources()
	case "elasticache":
		return c.discoverElastiCacheResources()
	case "dynamodb":
		return c.discoverDynamoDBResources()
	case "redshift":
		return c.discoverRedshiftResources()
	case "route53":
		return c.discoverRoute53Resources()
	case "cloudfront":
		return c.discoverCloudFrontResources()
	case "apigateway":
		return c.discoverAPIGatewayResources()
	case "elasticsearch":
		return c.discoverElasticsearchResources()
	case "neptune":
		return c.discoverNeptuneResources()
	case "docdb":
		return c.discoverDocDBResources()
	case "elasticbeanstalk":
		return c.discoverElasticBeanstalkResources()
	case "ecr":
		return c.discoverECRResources()
	case "codecommit":
		return c.discoverCodeCommitResources()
	case "codebuild":
		return c.discoverCodeBuildResources()
	case "codedeploy":
		return c.discoverCodeDeployResources()
	case "codepipeline":
		return c.discoverCodePipelineResources()
	case "cloudformation":
		return c.discoverCloudFormationResources()
	case "ssm":
		return c.discoverSSMResources()
	case "secretsmanager":
		return c.discoverSecretsManagerResources()
	case "kms":
		return c.discoverKMSResources()
	case "guardduty":
		return c.discoverGuardDutyResources()
	case "config":
		return c.discoverConfigResources()
	case "backup":
		return c.discoverBackupResources()
	case "glacier":
		return c.discoverGlacierResources()
	case "glue":
		return c.discoverGlueResources()
	case "athena":
		return c.discoverAthenaResources()
	case "quicksight":
		return c.discoverQuickSightResources()
	case "workspaces":
		return c.discoverWorkSpacesResources()
	// Removed stub service: directoryservice
	case "fsx":
		return c.discoverFSxResources()
	case "storagegateway":
		return c.discoverStorageGatewayResources()
	case "transfer":
		return c.discoverTransferResources()
	case "mq":
		return c.discoverMQResources()
	case "kinesis":
		return c.discoverKinesisResources()
	case "firehose":
		return c.discoverFirehoseResources()
	case "mediastore":
		return c.discoverMediaStoreResources()
	case "mediaconvert":
		return c.discoverMediaConvertResources()
	case "medialive":
		return c.discoverMediaLiveResources()
	case "mediatailor":
		return c.discoverMediaTailorResources()
	case "iot":
		return c.discoverIoTResources()
	// Removed stub services: greengrass, greengrassv2
	case "iotanalytics":
		return c.discoverIoTAnalyticsResources()
	case "iotevents":
		return c.discoverIoTEventsResources()
	case "iotsitewise":
		return c.discoverIoTSiteWiseResources()
	case "iotthingsgraph":
		return c.discoverIoTThingsGraphResources()
	case "iotwireless":
		return c.discoverIoTWirelessResources()
	// Removed stub services: iotdeviceadvisor, iotfleethub, iotsecuretunneling
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// GetAllSupportedResources returns all supported resource types
func GetAllSupportedResources() []string {
	return []string{
		// Core Infrastructure
		"vpc", "ec2", "iam", "rds", "s3", "alb", "elb", "asg", "lambda", "sqs", "sns",
		
		// Monitoring & Logging
		"cloudwatch", "cloudtrail", "config",
		
		// Containers & Orchestration
		"ecs", "eks", "ecr",
		
		// Databases
		"elasticache", "redshift", "dynamodb", "elasticsearch", "neptune", "docdb",
		
		// Networking & CDN
		"route53", "cloudfront", "apigateway",
		
		// Development & CI/CD
		"codebuild", "codedeploy", "codecommit", "codepipeline", "elasticbeanstalk",
		
		// Security & Management
		"ssm", "secretsmanager", "kms", "guardduty",
		
		// Storage & File Systems
		"fsx", "backup", "glacier", "storagegateway", "transfer",
		
		// Analytics & Streaming
		"kinesis", "firehose", "glue", "athena", "quicksight",
		
		// Media Services
		"mediastore", "mediaconvert", "medialive", "mediatailor",
		
		// IoT & Edge Computing
		"iot", "iotanalytics", "iotevents", "iotsitewise",
		"iotthingsgraph", "iotwireless",
		
		// Enterprise Services
		"workspaces", "mq",
	}
} 