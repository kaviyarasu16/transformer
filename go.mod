module transformer

go 1.22

toolchain go1.24.5

require (
	github.com/aws/aws-sdk-go-v2 v1.37.1
	github.com/aws/aws-sdk-go-v2/config v1.26.6
	github.com/aws/aws-sdk-go-v2/service/apigatewayv2 v1.29.1
	github.com/aws/aws-sdk-go-v2/service/athena v1.52.1
	github.com/aws/aws-sdk-go-v2/service/autoscaling v1.35.0
	github.com/aws/aws-sdk-go-v2/service/backup v1.44.1
	github.com/aws/aws-sdk-go-v2/service/cloudformation v1.62.1
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.49.0
	github.com/aws/aws-sdk-go-v2/service/cloudtrail v1.35.0
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.32.2
	github.com/aws/aws-sdk-go-v2/service/codebuild v1.62.1
	github.com/aws/aws-sdk-go-v2/service/codecommit v1.29.1
	github.com/aws/aws-sdk-go-v2/service/codedeploy v1.31.1
	github.com/aws/aws-sdk-go-v2/service/codepipeline v1.43.1
	github.com/aws/aws-sdk-go-v2/service/configservice v1.54.1
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.45.1
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.149.1
	github.com/aws/aws-sdk-go-v2/service/ecr v1.47.1
	github.com/aws/aws-sdk-go-v2/service/ecs v1.40.0
	github.com/aws/aws-sdk-go-v2/service/eks v1.40.0
	github.com/aws/aws-sdk-go-v2/service/elasticache v1.47.1
	github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk v1.30.1
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.26.0
	github.com/aws/aws-sdk-go-v2/service/elasticsearchservice v1.34.1
	github.com/aws/aws-sdk-go-v2/service/firehose v1.38.1
	github.com/aws/aws-sdk-go-v2/service/fsx v1.56.1
	github.com/aws/aws-sdk-go-v2/service/glacier v1.28.1
	github.com/aws/aws-sdk-go-v2/service/glue v1.120.1
	github.com/aws/aws-sdk-go-v2/service/guardduty v1.58.1
	github.com/aws/aws-sdk-go-v2/service/iam v1.28.7
	github.com/aws/aws-sdk-go-v2/service/iot v1.65.1
	github.com/aws/aws-sdk-go-v2/service/iotanalytics v1.28.1
	github.com/aws/aws-sdk-go-v2/service/iotevents v1.29.1
	github.com/aws/aws-sdk-go-v2/service/iotsitewise v1.48.1
	github.com/aws/aws-sdk-go-v2/service/iotthingsgraph v1.27.1
	github.com/aws/aws-sdk-go-v2/service/iotwireless v1.50.0
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.36.1
	github.com/aws/aws-sdk-go-v2/service/kms v1.42.1
	github.com/aws/aws-sdk-go-v2/service/lambda v1.49.7
	github.com/aws/aws-sdk-go-v2/service/mediaconvert v1.78.1
	github.com/aws/aws-sdk-go-v2/service/medialive v1.77.1
	github.com/aws/aws-sdk-go-v2/service/mediastore v1.26.1
	github.com/aws/aws-sdk-go-v2/service/mediatailor v1.49.1
	github.com/aws/aws-sdk-go-v2/service/mq v1.30.1
	github.com/aws/aws-sdk-go-v2/service/quicksight v1.89.1
	github.com/aws/aws-sdk-go-v2/service/rds v1.74.0
	github.com/aws/aws-sdk-go-v2/service/redshift v1.55.1
	github.com/aws/aws-sdk-go-v2/service/route53 v1.54.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.48.1
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.36.1
	github.com/aws/aws-sdk-go-v2/service/sns v1.26.0
	github.com/aws/aws-sdk-go-v2/service/sqs v1.29.7
	github.com/aws/aws-sdk-go-v2/service/ssm v1.61.1
	github.com/aws/aws-sdk-go-v2/service/storagegateway v1.39.1
	github.com/aws/aws-sdk-go-v2/service/transfer v1.62.1
	github.com/aws/aws-sdk-go-v2/service/workspaces v1.59.1
	github.com/charmbracelet/bubbletea v0.24.2
	github.com/charmbracelet/lipgloss v0.7.1
	github.com/spf13/cobra v1.8.0
	github.com/spf13/viper v1.18.2
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.0 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.16.16 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.14.11 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.7.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.2.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.2.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.11.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.16.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.18.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.21.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.26.7 // indirect
	github.com/aws/smithy-go v1.22.5 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/containerd/console v1.0.4-0.20230313162750-1ae8d489ac81 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-isatty v0.0.18 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/muesli/ansi v0.0.0-20211018074035-2e021307bc4b // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/muesli/termenv v0.15.1 // indirect
	github.com/pelletier/go-toml/v2 v2.1.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/term v0.6.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
