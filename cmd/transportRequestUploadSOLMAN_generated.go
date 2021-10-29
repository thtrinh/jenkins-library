// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type transportRequestUploadSOLMANOptions struct {
	Endpoint           string   `json:"endpoint,omitempty"`
	Username           string   `json:"username,omitempty"`
	Password           string   `json:"password,omitempty"`
	ApplicationID      string   `json:"applicationId,omitempty"`
	ChangeDocumentID   string   `json:"changeDocumentId,omitempty"`
	TransportRequestID string   `json:"transportRequestId,omitempty"`
	FilePath           string   `json:"filePath,omitempty"`
	CmClientOpts       []string `json:"cmClientOpts,omitempty"`
}

type transportRequestUploadSOLMANCommonPipelineEnvironment struct {
	custom struct {
		changeDocumentID   string
		transportRequestID string
	}
}

func (p *transportRequestUploadSOLMANCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    interface{}
	}{
		{category: "custom", name: "changeDocumentId", value: p.custom.changeDocumentID},
		{category: "custom", name: "transportRequestId", value: p.custom.transportRequestID},
	}

	errCount := 0
	for _, param := range content {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(param.category, param.name), param.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting piper environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Fatal("failed to persist Piper environment")
	}
}

// TransportRequestUploadSOLMANCommand Uploads a specified file into a given transport via Solution Manager
func TransportRequestUploadSOLMANCommand() *cobra.Command {
	const STEP_NAME = "transportRequestUploadSOLMAN"

	metadata := transportRequestUploadSOLMANMetadata()
	var stepConfig transportRequestUploadSOLMANOptions
	var startTime time.Time
	var commonPipelineEnvironment transportRequestUploadSOLMANCommonPipelineEnvironment
	var logCollector *log.CollectorHook
	var provider orchestrator.OrchestratorSpecificConfigProviding
	var err error
	splunkClient := &splunk.Splunk{}
	telemetryClient := &telemetry.Telemetry{}
	provider, err = orchestrator.NewOrchestratorSpecificConfigProvider()
	if err != nil {
		log.Entry().Error(err)
		provider = &orchestrator.UnknownOrchestratorConfigProvider{}
	}

	var createTransportRequestUploadSOLMANCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Uploads a specified file into a given transport via Solution Manager",
		Long: `Uploads the specified file into the given transport request via Solution Manager.
The mandatory change document ID points to the associate change document item.
The application ID specifies how the file needs to be handled on server side.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}
			log.RegisterSecret(stepConfig.Username)
			log.RegisterSecret(stepConfig.Password)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			validation, err := validation.New(validation.WithJSONNamesForStructFields(), validation.WithPredefinedErrorMessages())
			if err != nil {
				return err
			}
			if err = validation.ValidateStruct(stepConfig); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			customTelemetryData := telemetry.CustomData{}
			customTelemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				commonPipelineEnvironment.persist(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
				customTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				customTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				customTelemetryData.Custom1Label = "PiperCommitHash"
				customTelemetryData.Custom1 = GitCommit
				customTelemetryData.Custom2Label = "PiperTag"
				customTelemetryData.Custom2 = GitTag
				customTelemetryData.Custom3Label = "Stage"
				customTelemetryData.Custom3 = provider.GetStageName()
				customTelemetryData.Custom4Label = "Orchestrator"
				customTelemetryData.Custom4 = provider.OrchestratorType()
				telemetryClient.SetData(&customTelemetryData)
				telemetryClient.Send()
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunkClient.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			transportRequestUploadSOLMAN(stepConfig, &customTelemetryData, &commonPipelineEnvironment)
			customTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addTransportRequestUploadSOLMANFlags(createTransportRequestUploadSOLMANCmd, &stepConfig)
	return createTransportRequestUploadSOLMANCmd
}

func addTransportRequestUploadSOLMANFlags(cmd *cobra.Command, stepConfig *transportRequestUploadSOLMANOptions) {
	cmd.Flags().StringVar(&stepConfig.Endpoint, "endpoint", os.Getenv("PIPER_endpoint"), "Service endpoint")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "Service user for uploading to the Solution Manager")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Service user password for uploading to the Solution Manager")
	cmd.Flags().StringVar(&stepConfig.ApplicationID, "applicationId", os.Getenv("PIPER_applicationId"), "Id of the application. Specifies how the file needs to be handled on server side")
	cmd.Flags().StringVar(&stepConfig.ChangeDocumentID, "changeDocumentId", os.Getenv("PIPER_changeDocumentId"), "ID of the change document to which the file is uploaded")
	cmd.Flags().StringVar(&stepConfig.TransportRequestID, "transportRequestId", os.Getenv("PIPER_transportRequestId"), "ID of the transport request to which the file is uploaded")
	cmd.Flags().StringVar(&stepConfig.FilePath, "filePath", os.Getenv("PIPER_filePath"), "Name/Path of the file which should be uploaded")
	cmd.Flags().StringSliceVar(&stepConfig.CmClientOpts, "cmClientOpts", []string{}, "Additional options handed over to the cm client")

	cmd.MarkFlagRequired("endpoint")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("applicationId")
	cmd.MarkFlagRequired("changeDocumentId")
	cmd.MarkFlagRequired("transportRequestId")
	cmd.MarkFlagRequired("filePath")
	cmd.MarkFlagRequired("cmClientOpts")
}

// retrieve step metadata
func transportRequestUploadSOLMANMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "transportRequestUploadSOLMAN",
			Aliases:     []config.Alias{{Name: "transportRequestUploadFile", Deprecated: false}},
			Description: "Uploads a specified file into a given transport via Solution Manager",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "uploadCredentialsId", Description: "Jenkins 'Username with password' credentials ID containing user and password to authenticate against the ABAP backend", Type: "jenkins", Aliases: []config.Alias{{Name: "changeManagement/credentialsId", Deprecated: false}}},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "endpoint",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "changeManagement/endpoint"}},
						Default:     os.Getenv("PIPER_endpoint"),
					},
					{
						Name: "username",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "uploadCredentialsId",
								Param: "username",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_username"),
					},
					{
						Name: "password",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "uploadCredentialsId",
								Param: "password",
								Type:  "secret",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_password"),
					},
					{
						Name:        "applicationId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_applicationId"),
					},
					{
						Name: "changeDocumentId",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/changeDocumentId",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_changeDocumentId"),
					},
					{
						Name: "transportRequestId",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/transportRequestId",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_transportRequestId"),
					},
					{
						Name: "filePath",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "mtarFilePath",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_filePath"),
					},
					{
						Name:        "cmClientOpts",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "[]string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "clientOpts"}, {Name: "changeManagement/clientOpts"}},
						Default:     []string{},
					},
				},
			},
			Containers: []config.Container{
				{Name: "cmclient", Image: "ppiper/cm-client"},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "commonPipelineEnvironment",
						Type: "piperEnvironment",
						Parameters: []map[string]interface{}{
							{"Name": "custom/changeDocumentId"},
							{"Name": "custom/transportRequestId"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
