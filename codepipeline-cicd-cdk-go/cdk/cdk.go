package main

import (
	"github.com/aws/aws-cdk-go/awscdk/awscodebuild"
	"github.com/aws/aws-cdk-go/awscdk/awscodepipeline"
	"github.com/aws/aws-cdk-go/awscdk/awscodepipelineactions"
	"github.com/aws/aws-cdk-go/awscdk/awss3"
	"github.com/aws/aws-cdk-go/awscdk"
	"github.com/aws/constructs-go/constructs/v3"
	"github.com/aws/jsii-runtime-go"
)

var FirstStageArtifactName string = "GithubSource"
var SecondStageArtifactName string = "CIArtifact"
var FirstStageNameSpace string = "SourceVariables"
var SecondStageNameSpace string = "ValidateAndBuildSources"
var Project awscodebuild.Project

type CdkStackProps struct {
	awscdk.StackProps
	GithubOwner *string
	GithubRepoName *string
	BranchName *string
}

func NewCdkStack(scope constructs.Construct, id string, props *CdkStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	
	if props != nil {
		sprops = props.StackProps
	}

	stack := awscdk.NewStack(scope, &id, &sprops)
	createCodeBuildStack(stack, "CdkGoDemoCodeBuildStack")
	codepipelineBucket := awss3.NewBucket(stack, jsii.String("DemoCodePipelineBucket"), &awss3.BucketProps{
		AutoDeleteObjects: jsii.Bool(true),
		BucketName: jsii.String("cdkgo-demo-codepipeline-bucket"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	sourceAction := createSourceAction(props)
	buildAction := createBuildAction(sourceAction)

	sourceActions := []awscodepipeline.IAction{}
	sourceActions = append(sourceActions, sourceAction)
	buildActions := []awscodepipeline.IAction{}
	buildActions = append(buildActions, buildAction)

	demoCodepipeline := awscodepipeline.NewPipeline(stack, jsii.String("DemoCodePipeline"), &awscodepipeline.PipelineProps{
		ArtifactBucket: codepipelineBucket,
		PipelineName: jsii.String("cdkgo-demo-codepipeline"),
		RestartExecutionOnUpdate: jsii.Bool(false),
	})
	demoCodepipeline.AddStage(&awscodepipeline.StageOptions{
		StageName: jsii.String("RetrieveStage"),
		Actions: &sourceActions,
	})
	demoCodepipeline.AddStage(&awscodepipeline.StageOptions{
		StageName: jsii.String("ValidateAndBuildSources"),
		Actions: &buildActions,
	})

	awscdk.NewCfnOutput(stack, jsii.String("OutputS3BucketName"), &awscdk.CfnOutputProps{Value: codepipelineBucket.BucketName()})
	awscdk.NewCfnOutput(stack, jsii.String("OutputS3BucketArn"), &awscdk.CfnOutputProps{Value: codepipelineBucket.BucketArn()})
	awscdk.NewCfnOutput(stack, jsii.String("OutputPipelineName"), &awscdk.CfnOutputProps{Value: demoCodepipeline.PipelineName()})
	awscdk.NewCfnOutput(stack, jsii.String("OutputCodeBuildProjectName"), &awscdk.CfnOutputProps{Value: Project.ProjectName()})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewCdkStack(app, "CdkGoDemoCodePipelineStack", &CdkStackProps{
		awscdk.StackProps {
			Env: env(),
			Description: jsii.String("A CodePipeline with CodeBuild and CodeDeploy from Github as source deployed by cdk-go."),
			StackName: jsii.String("cdk-go-demo-codepipeline"),
		},
		jsii.String("HsiehShuJeng"),
		jsii.String("cdk-journey"),
		jsii.String("main"),
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}

func createSourceAction(props *CdkStackProps) awscodepipelineactions.GitHubSourceAction {
	sourceOutput := awscodepipeline.NewArtifact(jsii.String(FirstStageArtifactName))
	sourceAction := awscodepipelineactions.NewGitHubSourceAction(&awscodepipelineactions.GitHubSourceActionProps{
		ActionName: jsii.String("DownloadSource"),
		RunOrder: jsii.Number(1),
		VariablesNamespace: jsii.String(FirstStageNameSpace),
		OauthToken: awscdk.SecretValue_SecretsManager(jsii.String("github/access"), &awscdk.SecretsManagerSecretOptions{
			JsonField: jsii.String("DemoToken"),
		}),
		Output: sourceOutput,
		Owner: props.GithubOwner,
		Repo: props.GithubRepoName,
		Branch: props.BranchName,
		Trigger: awscodepipelineactions.GitHubTrigger_WEBHOOK,
	})
	return sourceAction
}

func createBuildAction(sourceAction awscodepipelineactions.GitHubSourceAction) awscodepipelineactions.CodeBuildAction {
	buildInput := awscodepipeline.NewArtifact(jsii.String(FirstStageArtifactName))
	buildOutput := awscodepipeline.NewArtifact(jsii.String(SecondStageArtifactName))

	environmentVariables := make(map[string]*awscodebuild.BuildEnvironmentVariable)
	environmentVariables["COMMITTER_DATE"] = &awscodebuild.BuildEnvironmentVariable{Value: sourceAction.Variables().CommitterDate}
	environmentVariables["COMMIT_ID"] = &awscodebuild.BuildEnvironmentVariable{Value: sourceAction.Variables().CommitId}
	buildAction := awscodepipelineactions.NewCodeBuildAction(&awscodepipelineactions.CodeBuildActionProps{
		ActionName: jsii.String("BuildArtifacts"),
		RunOrder: jsii.Number(1),
		VariablesNamespace: jsii.String(SecondStageNameSpace),
		Input: buildInput,
		Project: Project,
		EnvironmentVariables: &environmentVariables,
		Outputs: &[]awscodepipeline.Artifact{buildOutput},
	})
	return buildAction
}

func createCodeBuildStack(scope constructs.Construct, id string) awscdk.NestedStack {
	nestedStack := awscdk.NewNestedStack(scope, &id, &awscdk.NestedStackProps{})
	Project = awscodebuild.NewPipelineProject(nestedStack, jsii.String("CdkGoDemoProject"), &awscodebuild.PipelineProjectProps{
		Cache: awscodebuild.Cache_Local(awscodebuild.LocalCacheMode_SOURCE, awscodebuild.LocalCacheMode_CUSTOM),
		CheckSecretsInPlainTextEnvVariables: jsii.Bool(true),
		Description: jsii.String("CI for the CodePipeline demo via cdk-go."),
		Environment: &awscodebuild.BuildEnvironment{
			BuildImage: awscodebuild.LinuxBuildImage_STANDARD_5_0(),
			ComputeType: awscodebuild.ComputeType_SMALL,
		},
		ProjectName: jsii.String("cdkgo-demo-codebuild-project"),
		QueuedTimeout: awscdk.Duration_Minutes(jsii.Number(15)),
		Timeout: awscdk.Duration_Minutes(jsii.Number(5)),
	})

	return nestedStack
}