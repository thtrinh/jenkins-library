// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/spf13/cobra"
)

type transportRequestUploadRFCOptions struct {
	Endpoint                   string `json:"endpoint,omitempty"`
	Client                     string `json:"client,omitempty"`
	Instance                   string `json:"instance,omitempty"`
	Username                   string `json:"username,omitempty"`
	Password                   string `json:"password,omitempty"`
	ApplicationName            string `json:"applicationName,omitempty"`
	ApplicationDescription     string `json:"applicationDescription,omitempty"`
	AbapPackage                string `json:"abapPackage,omitempty"`
	ApplicationURL             string `json:"applicationUrl,omitempty"`
	CodePage                   string `json:"codePage,omitempty"`
	AcceptUnixStyleLineEndings bool   `json:"acceptUnixStyleLineEndings,omitempty"`
	FailUploadOnWarning        bool   `json:"failUploadOnWarning,omitempty"`
	TransportRequestID         string `json:"transportRequestId,omitempty"`
}

// TransportRequestUploadRFCCommand Uploads content to a transport request
func TransportRequestUploadRFCCommand() *cobra.Command {
	const STEP_NAME = "transportRequestUploadRFC"

	metadata := transportRequestUploadRFCMetadata()
	var stepConfig transportRequestUploadRFCOptions
	var startTime time.Time
	var logCollector *log.CollectorHook

	var createTransportRequestUploadRFCCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Uploads content to a transport request",
		Long:  `Uploads content to a transport request.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

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

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetryData.ErrorCategory = log.GetErrorCategory().String()
				telemetry.Send(&telemetryData)
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunk.Send(&telemetryData, logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunk.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			transportRequestUploadRFC(stepConfig, &telemetryData)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addTransportRequestUploadRFCFlags(createTransportRequestUploadRFCCmd, &stepConfig)
	return createTransportRequestUploadRFCCmd
}

func addTransportRequestUploadRFCFlags(cmd *cobra.Command, stepConfig *transportRequestUploadRFCOptions) {
	cmd.Flags().StringVar(&stepConfig.Endpoint, "endpoint", os.Getenv("PIPER_endpoint"), "Service endpoint")
	cmd.Flags().StringVar(&stepConfig.Client, "client", os.Getenv("PIPER_client"), "AS ABAP client number")
	cmd.Flags().StringVar(&stepConfig.Instance, "instance", os.Getenv("PIPER_instance"), "AS ABAP instance number")
	cmd.Flags().StringVar(&stepConfig.Username, "username", os.Getenv("PIPER_username"), "Deploy user")
	cmd.Flags().StringVar(&stepConfig.Password, "password", os.Getenv("PIPER_password"), "Password for the deploy user")
	cmd.Flags().StringVar(&stepConfig.ApplicationName, "applicationName", os.Getenv("PIPER_applicationName"), "Name of the application.")
	cmd.Flags().StringVar(&stepConfig.ApplicationDescription, "applicationDescription", os.Getenv("PIPER_applicationDescription"), "Description of the application.")
	cmd.Flags().StringVar(&stepConfig.AbapPackage, "abapPackage", os.Getenv("PIPER_abapPackage"), "ABAP package name of your application")
	cmd.Flags().StringVar(&stepConfig.ApplicationURL, "applicationUrl", os.Getenv("PIPER_applicationUrl"), "URL where to find the UI5 package to upload to the transport request")
	cmd.Flags().StringVar(&stepConfig.CodePage, "codePage", `UTF-8`, "Code page")
	cmd.Flags().BoolVar(&stepConfig.AcceptUnixStyleLineEndings, "acceptUnixStyleLineEndings", true, "If unix style line endings should be accepted")
	cmd.Flags().BoolVar(&stepConfig.FailUploadOnWarning, "failUploadOnWarning", true, "If the upload should fail in case the log contains warnings")
	cmd.Flags().StringVar(&stepConfig.TransportRequestID, "transportRequestId", os.Getenv("PIPER_transportRequestId"), "Transport request id")

	cmd.MarkFlagRequired("endpoint")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("applicationName")
	cmd.MarkFlagRequired("abapPackage")
	cmd.MarkFlagRequired("applicationUrl")
	cmd.MarkFlagRequired("abapPackage")
	cmd.MarkFlagRequired("transportRequestId")
}

// retrieve step metadata
func transportRequestUploadRFCMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "transportRequestUploadRFC",
			Aliases:     []config.Alias{{Name: "transportRequestUploadFile", Deprecated: false}},
			Description: "Uploads content to a transport request",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "uploadCredentialsId", Description: "Jenkins 'Username with password' credentials ID containing user and password to authenticate against the ABAP backend.", Type: "jenkins", Aliases: []config.Alias{{Name: "changeManagement/credentialsId", Deprecated: false}}},
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
						Name:        "client",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/rfc/developmentClient"}},
						Default:     os.Getenv("PIPER_client"),
					},
					{
						Name:        "instance",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "changeManagement/rfc/developmentInstance"}},
						Default:     os.Getenv("PIPER_instance"),
					},
					{
						Name:        "username",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_username"),
					},
					{
						Name:        "password",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_password"),
					},
					{
						Name:        "applicationName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_applicationName"),
					},
					{
						Name:        "applicationDescription",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_applicationDescription"),
					},
					{
						Name:        "abapPackage",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_abapPackage"),
					},
					{
						Name:        "applicationUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_applicationUrl"),
					},
					{
						Name:        "abapPackage",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_abapPackage"),
					},
					{
						Name:        "codePage",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `UTF-8`,
					},
					{
						Name:        "acceptUnixStyleLineEndings",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     true,
					},
					{
						Name:        "failUploadOnWarning",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS", "GENERAL"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "failOnWarning"}},
						Default:     true,
					},
					{
						Name:        "transportRequestId",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_transportRequestId"),
					},
				},
			},
		},
	}
	return theMetaData
}
