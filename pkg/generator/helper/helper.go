package helper

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type stepInfo struct {
	CobraCmdFuncName string
	CreateCmdVar     string
	ExportPrefix     string
	FlagsFunc        string
	Long             string
	StepParameters   []config.StepParameters
	StepAliases      []config.Alias
	OSImport         bool
	OutputResources  []map[string]string
	Short            string
	StepFunc         string
	StepName         string
	StepSecrets      []string
	Containers       []config.Container
	Sidecars         []config.Container
	Outputs          config.StepOutputs
	Resources        []config.StepResources
	Secrets          []config.StepSecrets
}

//StepGoTemplate ...
const stepGoTemplate = `// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	{{ if .OutputResources -}}
	"path/filepath"
	{{ end -}}
	"time"

	{{ if .ExportPrefix -}}
	{{ .ExportPrefix }} "github.com/SAP/jenkins-library/cmd"
	{{ end -}}
	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	{{ if .OutputResources -}}
	"github.com/SAP/jenkins-library/pkg/piperenv"
	{{ end -}}
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/SAP/jenkins-library/pkg/orchestrator"
	"github.com/spf13/cobra"
)

type {{ .StepName }}Options struct {
	{{- $names := list ""}}
	{{- range $key, $value := uniqueName .StepParameters }}
	{{ if ne (has $value.Name $names) true -}}
	{{ $names | last }}{{ $value.Name | golangName }} {{ $value.Type }} ` + "`json:\"{{$value.Name}},omitempty\"" +
	"{{ if or $value.PossibleValues $value.MandatoryIf}} validate:\"" +
	"{{ if $value.PossibleValues }}oneof={{ range $i,$a := $value.PossibleValues }}{{if gt $i 0 }} {{ end }}{{.}}{{ end }}{{ end }}" +
	"{{ if and $value.PossibleValues $value.MandatoryIf }},{{ end }}" +
	"{{ if $value.MandatoryIf }}required_if={{ range $i,$a := $value.MandatoryIf }}{{ if gt $i 0 }} {{ end }}{{ $a.Name | title }} {{ $a.Value }}{{ end }}{{ end }}" +
	"\"{{ end }}`" + `
	{{- else -}}
	{{- $names = append $names $value.Name }} {{ end -}}
	{{ end }}
}

{{ range $notused, $oRes := .OutputResources }}
{{ index $oRes "def"}}
{{ end }}

// {{.CobraCmdFuncName}} {{.Short}}
func {{.CobraCmdFuncName}}() *cobra.Command {
	const STEP_NAME = {{ .StepName | quote }}

	metadata := {{ .StepName }}Metadata()
	var stepConfig {{.StepName}}Options
	var startTime time.Time
	{{- range $notused, $oRes := .OutputResources }}
	var {{ index $oRes "name" }} {{ index $oRes "objectname" }}{{ end }}
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

	var {{.CreateCmdVar}} = &cobra.Command{
		Use:   STEP_NAME,
		Short: {{.Short | quote }},
		Long: {{ $tick := "` + "`" + `" }}{{ $tick }}{{.Long | longName }}{{ $tick }},
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose({{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.Verbose)

			{{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.GitHubAccessTokens = {{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}ResolveAccessTokens({{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.GitHubTokens)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: {{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := {{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			{{- range $key, $value := .StepSecrets }}
			log.RegisterSecret(stepConfig.{{ $value | golangName  }}){{end}}

			if len({{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook({{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.HookConfig.SentryConfig.Dsn, {{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len({{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				logCollector = &log.CollectorHook{CorrelationID: {{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.CorrelationID}
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
				{{- range $notused, $oRes := .OutputResources }}
				{{ index $oRes "name" }}.persist({{if $.ExportPrefix}}{{ $.ExportPrefix }}.{{end}}GeneralConfig.EnvRootPath, {{ index $oRes "name" | quote }}){{ end }}
				customTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				customTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				customTelemetryData.Custom1Label = "PiperCommitHash"
				customTelemetryData.Custom1 = {{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GitCommit
				customTelemetryData.Custom2Label = "PiperTag"
				customTelemetryData.Custom2 = {{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GitTag
				customTelemetryData.Custom3Label = "Stage"
				customTelemetryData.Custom3 = provider.GetStageName()
				customTelemetryData.Custom4Label = "Orchestrator"
                                customTelemetryData.Custom4 = provider.OrchestratorType()
				telemetryClient.SetData(&customTelemetryData)
				telemetryClient.Send()
				if len({{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize({{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.NoTelemetry, STEP_NAME)
			if len({{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunkClient.Initialize({{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.CorrelationID,
				{{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.HookConfig.SplunkConfig.Dsn,
				{{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.HookConfig.SplunkConfig.Token,
				{{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.HookConfig.SplunkConfig.Index,
				{{if .ExportPrefix}}{{ .ExportPrefix }}.{{end}}GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			{{.StepName}}(stepConfig, &customTelemetryData{{ range $notused, $oRes := .OutputResources}}, &{{ index $oRes "name" }}{{ end }})
			customTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	{{.FlagsFunc}}({{.CreateCmdVar}}, &stepConfig)
	return {{.CreateCmdVar}}
}

func {{.FlagsFunc}}(cmd *cobra.Command, stepConfig *{{.StepName}}Options) {
	{{- range $key, $value := uniqueName .StepParameters }}
	{{ if isCLIParam $value.Type }}cmd.Flags().{{ $value.Type | flagType }}(&stepConfig.{{ $value.Name | golangName }}, {{ $value.Name | quote }}, {{ $value.Default }}, {{ $value.Description | quote }}){{end}}{{ end }}
	{{- printf "\n" }}
	{{- range $key, $value := .StepParameters }}{{ if $value.Mandatory }}
	cmd.MarkFlagRequired({{ $value.Name | quote }}){{ end }}{{ end }}
}

{{ define "resourceRefs"}}
							{{ "{" }}
								Name: {{- .Name | quote }},
								{{- if .Param }}
								Param: {{ .Param | quote }},
								{{- end }}
								{{- if .Type }}
								Type: {{ .Type | quote }},
								{{- if .Default }}
								Default: {{ .Default | quote }},
								{{- end}}
								{{- end }}
							{{ "}" }},
							{{- nindent 24 ""}}
{{- end -}}

// retrieve step metadata
func {{ .StepName }}Metadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:    {{ .StepName | quote }},
			Aliases: []config.Alias{{ "{" }}{{ range $notused, $alias := .StepAliases }}{{ "{" }}Name: {{ $alias.Name | quote }}, Deprecated: {{ $alias.Deprecated }}{{ "}" }},{{ end }}{{ "}" }},
			Description: {{ .Short | quote }},
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				{{ if .Secrets -}}
				Secrets: []config.StepSecrets{
					{{- range $secrets := .Secrets }}
					{
						{{- if $secrets.Name -}} Name: {{ $secrets.Name | quote }},{{- end }}
						{{- if $secrets.Description -}} Description: {{$secrets.Description | quote}},{{- end }}
						{{- if $secrets.Type -}} Type: {{ $secrets.Type | quote }},{{- end }}
						{{- if $secrets.Aliases -}} Aliases: []config.Alias{ {{- range $i, $a := $secrets.Aliases }} {Name: {{ $a.Name | quote }}, Deprecated: {{$a.Deprecated}}}, {{ end -}}  },{{- end }}
					}, {{ end }}
				},
				{{ end -}}
				{{ if .Resources -}}
				Resources: []config.StepResources{
					{{- range $resource := .Resources }}
					{
						{{- if $resource.Name -}} Name: {{ $resource.Name | quote }},{{- end }}
						{{- if $resource.Description -}} Description: {{ $resource.Description | quote }},{{- end }}
						{{- if $resource.Type -}} Type: {{ $resource.Type | quote }},{{- end }}
						{{- if $resource.Conditions -}} Conditions: []config.Condition{ {{- range $i, $cond := $resource.Conditions }} {ConditionRef: {{ $cond.ConditionRef | quote }}, Params: []config.Param{ {{- range $j, $p := $cond.Params}} { Name: {{ $p.Name | quote }}, Value: {{ $p.Value | quote }} }, {{end -}} } }, {{ end -}} },{{ end }}
					},{{- end }}
				},
				{{ end -}}
				Parameters: []config.StepParameters{
					{{- range $key, $value := .StepParameters }}
					{
						Name:      {{ $value.Name | quote }},
						ResourceRef: []config.ResourceReference{{ "{" }}{{ range $notused, $ref := $value.ResourceRef }}{{ template "resourceRefs" $ref }}{{ end }}{{ "}" }},
						Scope:     []string{{ "{" }}{{ range $notused, $scope := $value.Scope }}{{ $scope | quote }},{{ end }}{{ "}" }},
						Type:      {{ $value.Type | quote }},
						Mandatory: {{ $value.Mandatory }},
						Aliases:   []config.Alias{{ "{" }}{{ range $notused, $alias := $value.Aliases }}{{ "{" }}Name: {{ $alias.Name | quote }}{{ "}" }},{{ end }}{{ "}" }},
						{{ if $value.Default -}} Default:   {{ $value.Default }}, {{- end}}{{ if $value.Conditions }}
						Conditions: []config.Condition{ {{- range $i, $cond := $value.Conditions }} {ConditionRef: {{ $cond.ConditionRef | quote }}, Params: []config.Param{ {{- range $j, $p := $cond.Params}} { Name: {{ $p.Name | quote }}, Value: {{ $p.Value | quote }} }, {{end -}} } }, {{ end -}} },{{- end }}
					},{{ end }}
				},
			},
			{{ if .Containers -}}
			Containers: []config.Container{
				{{- range $container := .Containers }}
				{
					{{- if $container.Name -}} Name: {{ $container.Name | quote }},{{- end }}
					{{- if $container.Image -}} Image: {{ $container.Image | quote }},{{- end }}
					{{- if $container.EnvVars -}} EnvVars: []config.EnvVar{ {{- range $i, $env := $container.EnvVars }} {Name: {{ $env.Name | quote }}, Value: {{ $env.Value | quote }}}, {{ end -}}  },{{- end }}
					{{- if $container.WorkingDir -}} WorkingDir: {{ $container.WorkingDir | quote }},{{- end }}
					{{- if $container.Options -}} Options: []config.Option{ {{- range $i, $option := $container.Options }} {Name: {{ $option.Name | quote }}, Value: {{ $option.Value | quote }}}, {{ end -}} },{{ end }}
					{{- if $container.Conditions -}} Conditions: []config.Condition{ {{- range $i, $cond := $container.Conditions }} {ConditionRef: {{ $cond.ConditionRef | quote }}, Params: []config.Param{ {{- range $j, $p := $cond.Params}} { Name: {{ $p.Name | quote }}, Value: {{ $p.Value | quote }} }, {{end -}} } }, {{ end -}} },{{ end }}
				}, {{ end }}
			},
			{{ end -}}
			{{ if .Sidecars -}}
			Sidecars: []config.Container{
				{{- range $container := .Sidecars }}
				{
					{{- if $container.Name -}} Name: {{ $container.Name | quote }}, {{- end }}
					{{- if $container.Image -}} Image: {{ $container.Image | quote }}, {{- end }}
					{{- if $container.EnvVars -}} EnvVars: []config.EnvVar{ {{- range $i, $env := $container.EnvVars }} {Name: {{ $env.Name | quote }}, Value: {{ $env.Value | quote }}}, {{ end -}}  }, {{- end }}
					{{- if $container.WorkingDir -}} WorkingDir: {{ $container.WorkingDir | quote }}, {{- end }}
					{{- if $container.Options -}} Options: []config.Option{ {{- range $i, $option := $container.Options }} {Name: {{ $option.Name | quote }}, Value: {{ $option.Value | quote }}}, {{ end -}} }, {{- end }}
					{{- if $container.Conditions -}} Conditions: []config.Condition{ {{- range $i, $cond := $container.Conditions }} {ConditionRef: {{ $cond.ConditionRef | quote }}, Params: []config.Param{ {{- range $j, $p := $cond.Params}} { Name: {{ $p.Name | quote }}, Value: {{ $p.Value | quote }} }, {{end -}} } }, {{ end -}} }, {{- end }}
				}, {{ end }}
			},
			{{ end -}}
			{{- if .Outputs.Resources -}}
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{{- range $res := .Outputs.Resources }}
					{
						{{ if $res.Name }}Name: {{ $res.Name | quote }}, {{- end }}
						{{ if $res.Type }}Type: {{ $res.Type | quote }}, {{- end }}
						{{ if $res.Parameters }}Parameters: []map[string]interface{}{ {{- end -}}
						{{ range $i, $p := $res.Parameters }}
							{{ if $p.name}}{"Name": {{ $p.name | quote }}},{{ end -}}
							{{ if $p.fields}}{"fields": []map[string]string{ {{- range $j, $f := $p.fields}} {"name": {{ $f.name | quote }}}, {{end -}} } },{{ end -}}
							{{ if $p.tags}}{"tags": []map[string]string{ {{- range $j, $t := $p.tags}} {"name": {{ $t.name | quote }}}, {{end -}} } },{{ end -}}
						{{ end }}
						{{ if $res.Parameters -}} }, {{- end }}
						{{- if $res.Conditions -}} Conditions: []config.Condition{ {{- range $i, $cond := $res.Conditions }} {ConditionRef: {{ $cond.ConditionRef | quote }}, Params: []config.Param{ {{- range $j, $p := $cond.Params}} { Name: {{ $p.Name | quote }}, Value: {{ $p.Value | quote }} }, {{end -}} } }, {{ end -}} },{{ end }}
					}, {{- end }}
				},
			}, {{- end }}
		},
	}
	return theMetaData
}
`

//StepTestGoTemplate ...
const stepTestGoTemplate = `package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test{{.CobraCmdFuncName}}(t *testing.T) {
	t.Parallel()

	testCmd := {{.CobraCmdFuncName}}()

	// only high level testing performed - details are tested in step generation procedure
	assert.Equal(t, {{ .StepName | quote }}, testCmd.Use, "command name incorrect")

}
`

const stepGoImplementationTemplate = `package cmd
import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/piperutils"
)

type {{.StepName}}Utils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The {{.StepName}}Utils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type {{.StepName}}UtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to {{.StepName}}Utils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// {{.StepName}}UtilsBundle and forward to the implementation of the dependency.
}

func new{{.StepName | title}}Utils() {{.StepName}}Utils {
	utils := {{.StepName}}UtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func {{.StepName}}(config {{ .StepName }}Options, telemetryData *telemetry.CustomData{{ range $notused, $oRes := .OutputResources}}, {{ index $oRes "name" }} *{{ index $oRes "objectname" }}{{ end }}) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := new{{.StepName | title}}Utils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := run{{.StepName | title}}(&config, telemetryData, utils{{ range $notused, $oRes := .OutputResources}}, &{{ index $oRes "name" }}{{ end }})
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func run{{.StepName | title}}(config *{{ .StepName }}Options, telemetryData *telemetry.CustomData, utils {{.StepName}}Utils{{ range $notused, $oRes := .OutputResources}}, {{ index $oRes "name" }} *{{ index $oRes "objectname" }} {{ end }}) error {
	log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")

	// Example of calling methods from external dependencies directly on utils:
	exists, err := utils.FileExists("file.txt")
	if err != nil {
		// It is good practice to set an error category.
		// Most likely you want to do this at the place where enough context is known.
		log.SetErrorCategory(log.ErrorConfiguration)
		// Always wrap non-descriptive errors to enrich them with context for when they appear in the log:
		return fmt.Errorf("failed to check for important file: %w", err)
	}
	if !exists {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("cannot run without important file")
	}

	return nil
}
`

const stepGoImplementationTestTemplate = `package cmd

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

type {{.StepName}}MockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func new{{.StepName | title}}TestsUtils() {{.StepName}}MockUtils {
	utils := {{.StepName}}MockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRun{{.StepName | title}}(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := {{.StepName}}Options{}

		utils := new{{.StepName | title}}TestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := run{{.StepName | title}}(&config, nil, utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := {{.StepName}}Options{}

		utils := new{{.StepName | title}}TestsUtils()

		// test
		err := run{{.StepName | title}}(&config, nil, utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
`

const metadataGeneratedFileName = "metadata_generated.go"
const metadataGeneratedTemplate = `// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import "github.com/SAP/jenkins-library/pkg/config"

// GetStepMetadata return a map with all the step metadata mapped to their stepName
func GetAllStepMetadata() map[string]config.StepData {
	return map[string]config.StepData{
		{{range $stepName := .Steps }} {{ $stepName | quote }}: {{$stepName}}Metadata(),
		{{end}}
	}
}
`

// ProcessMetaFiles generates step coding based on step configuration provided in yaml files
func ProcessMetaFiles(metadataFiles []string, targetDir string, stepHelperData StepHelperData) error {

	allSteps := struct{ Steps []string }{}
	for key := range metadataFiles {

		var stepData config.StepData

		configFilePath := metadataFiles[key]

		metadataFile, err := stepHelperData.OpenFile(configFilePath)
		checkError(err)
		defer metadataFile.Close()

		fmt.Printf("Reading file %v\n", configFilePath)

		err = stepData.ReadPipelineStepData(metadataFile)
		checkError(err)

		stepName := stepData.Metadata.Name
		fmt.Printf("Step name: %v\n", stepName)
		allSteps.Steps = append(allSteps.Steps, stepName)

		for _, parameter := range stepData.Spec.Inputs.Parameters {
			for _, mandatoryIfCase := range parameter.MandatoryIf {
				if mandatoryIfCase.Name == "" || mandatoryIfCase.Value == "" {
					return errors.New("invalid mandatoryIf option")
				}
			}
		}

		osImport := false
		osImport, err = setDefaultParameters(&stepData)
		checkError(err)

		myStepInfo, err := getStepInfo(&stepData, osImport, stepHelperData.ExportPrefix)
		checkError(err)

		step := stepTemplate(myStepInfo, "step", stepGoTemplate)
		err = stepHelperData.WriteFile(filepath.Join(targetDir, fmt.Sprintf("%v_generated.go", stepName)), step, 0644)
		checkError(err)

		test := stepTemplate(myStepInfo, "stepTest", stepTestGoTemplate)
		err = stepHelperData.WriteFile(filepath.Join(targetDir, fmt.Sprintf("%v_generated_test.go", stepName)), test, 0644)
		checkError(err)

		exists, _ := piperutils.FileExists(filepath.Join(targetDir, fmt.Sprintf("%v.go", stepName)))
		if !exists {
			impl := stepImplementation(myStepInfo, "impl", stepGoImplementationTemplate)
			err = stepHelperData.WriteFile(filepath.Join(targetDir, fmt.Sprintf("%v.go", stepName)), impl, 0644)
			checkError(err)
		}

		exists, _ = piperutils.FileExists(filepath.Join(targetDir, fmt.Sprintf("%v_test.go", stepName)))
		if !exists {
			impl := stepImplementation(myStepInfo, "implTest", stepGoImplementationTestTemplate)
			err = stepHelperData.WriteFile(filepath.Join(targetDir, fmt.Sprintf("%v_test.go", stepName)), impl, 0644)
			checkError(err)
		}
	}

	// expose metadata functions
	code := generateCode(allSteps, "metadata", metadataGeneratedTemplate, sprig.HermeticTxtFuncMap())
	err := stepHelperData.WriteFile(filepath.Join(targetDir, metadataGeneratedFileName), code, 0644)
	checkError(err)

	return nil
}

func setDefaultParameters(stepData *config.StepData) (bool, error) {
	//ToDo: custom function for default handling, support all relevant parameter types
	osImportRequired := false
	for k, param := range stepData.Spec.Inputs.Parameters {

		if param.Default == nil {
			switch param.Type {
			case "bool":
				// ToDo: Check if default should be read from env
				param.Default = "false"
			case "int":
				param.Default = "0"
			case "string":
				param.Default = fmt.Sprintf("os.Getenv(\"PIPER_%v\")", param.Name)
				osImportRequired = true
			case "[]string":
				// ToDo: Check if default should be read from env
				param.Default = "[]string{}"
			case "map[string]interface{}":
				// Currently we don't need to set a default here since in this case the default
				// is never used. Needs to be changed in case we enable cli parameter handling
				// for that type.
			default:
				return false, fmt.Errorf("Meta data type not set or not known: '%v'", param.Type)
			}
		} else {
			switch param.Type {
			case "bool":
				boolVal := "false"
				if param.Default.(bool) == true {
					boolVal = "true"
				}
				param.Default = boolVal
			case "int":
				param.Default = fmt.Sprintf("%v", param.Default)
			case "string":
				param.Default = fmt.Sprintf("`%v`", param.Default)
			case "[]string":
				param.Default = fmt.Sprintf("[]string{`%v`}", strings.Join(getStringSliceFromInterface(param.Default), "`, `"))
			case "map[string]interface{}":
				// Currently we don't need to set a default here since in this case the default
				// is never used. Needs to be changed in case we enable cli parameter handling
				// for that type.
			default:
				return false, fmt.Errorf("Meta data type not set or not known: '%v'", param.Type)
			}
		}

		stepData.Spec.Inputs.Parameters[k] = param
	}
	return osImportRequired, nil
}

func getStepInfo(stepData *config.StepData, osImport bool, exportPrefix string) (stepInfo, error) {
	oRes, err := getOutputResourceDetails(stepData)

	return stepInfo{
			StepName:         stepData.Metadata.Name,
			CobraCmdFuncName: fmt.Sprintf("%vCommand", strings.Title(stepData.Metadata.Name)),
			CreateCmdVar:     fmt.Sprintf("create%vCmd", strings.Title(stepData.Metadata.Name)),
			Short:            stepData.Metadata.Description,
			Long:             stepData.Metadata.LongDescription,
			StepParameters:   stepData.Spec.Inputs.Parameters,
			StepAliases:      stepData.Metadata.Aliases,
			FlagsFunc:        fmt.Sprintf("add%vFlags", strings.Title(stepData.Metadata.Name)),
			OSImport:         osImport,
			OutputResources:  oRes,
			ExportPrefix:     exportPrefix,
			StepSecrets:      getSecretFields(stepData),
			Containers:       stepData.Spec.Containers,
			Sidecars:         stepData.Spec.Sidecars,
			Outputs:          stepData.Spec.Outputs,
			Resources:        stepData.Spec.Inputs.Resources,
			Secrets:          stepData.Spec.Inputs.Secrets,
		},
		err
}

func getSecretFields(stepData *config.StepData) []string {
	var secretFields []string

	for _, parameter := range stepData.Spec.Inputs.Parameters {
		if parameter.Secret {
			secretFields = append(secretFields, parameter.Name)
		}
	}
	return secretFields
}

func getOutputResourceDetails(stepData *config.StepData) ([]map[string]string, error) {
	outputResources := []map[string]string{}

	for _, res := range stepData.Spec.Outputs.Resources {
		currentResource := map[string]string{}
		currentResource["name"] = res.Name

		switch res.Type {
		case "piperEnvironment":
			var envResource PiperEnvironmentResource
			envResource.Name = res.Name
			envResource.StepName = stepData.Metadata.Name
			for _, param := range res.Parameters {
				paramSections := strings.Split(fmt.Sprintf("%v", param["name"]), "/")
				category := ""
				name := paramSections[0]
				if len(paramSections) > 1 {
					name = strings.Join(paramSections[1:], "_")
					category = paramSections[0]
					if !contains(envResource.Categories, category) {
						envResource.Categories = append(envResource.Categories, category)
					}
				}
				envParam := PiperEnvironmentParameter{Category: category, Name: name, Type: fmt.Sprint(param["type"])}
				envResource.Parameters = append(envResource.Parameters, envParam)
			}
			def, err := envResource.StructString()
			if err != nil {
				return outputResources, err
			}
			currentResource["def"] = def
			currentResource["objectname"] = envResource.StructName()
			outputResources = append(outputResources, currentResource)
		case "influx":
			var influxResource InfluxResource
			influxResource.Name = res.Name
			influxResource.StepName = stepData.Metadata.Name
			for _, measurement := range res.Parameters {
				influxMeasurement := InfluxMeasurement{Name: fmt.Sprintf("%v", measurement["name"])}
				if fields, ok := measurement["fields"].([]interface{}); ok {
					for _, field := range fields {
						if fieldParams, ok := field.(map[string]interface{}); ok {
							influxMeasurement.Fields = append(influxMeasurement.Fields, InfluxMetric{Name: fmt.Sprintf("%v", fieldParams["name"]), Type: fmt.Sprintf("%v", fieldParams["type"])})
						}
					}
				}

				if tags, ok := measurement["tags"].([]interface{}); ok {
					for _, tag := range tags {
						if tagParams, ok := tag.(map[string]interface{}); ok {
							influxMeasurement.Tags = append(influxMeasurement.Tags, InfluxMetric{Name: fmt.Sprintf("%v", tagParams["name"])})
						}
					}
				}
				influxResource.Measurements = append(influxResource.Measurements, influxMeasurement)
			}
			def, err := influxResource.StructString()
			if err != nil {
				return outputResources, err
			}
			currentResource["def"] = def
			currentResource["objectname"] = influxResource.StructName()
			outputResources = append(outputResources, currentResource)
		}
	}

	return outputResources, nil
}

// MetadataFiles provides a list of all step metadata files
func MetadataFiles(sourceDirectory string) ([]string, error) {

	var metadataFiles []string

	err := filepath.Walk(sourceDirectory, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".yaml" {
			metadataFiles = append(metadataFiles, path)
		}
		return nil
	})
	if err != nil {
		return metadataFiles, nil
	}
	return metadataFiles, nil
}

func isCLIParam(myType string) bool {
	return myType != "map[string]interface{}"
}

func stepTemplate(myStepInfo stepInfo, templateName, goTemplate string) []byte {
	funcMap := sprig.HermeticTxtFuncMap()
	funcMap["flagType"] = flagType
	funcMap["golangName"] = golangNameTitle
	funcMap["title"] = strings.Title
	funcMap["longName"] = longName
	funcMap["uniqueName"] = mustUniqName
	funcMap["isCLIParam"] = isCLIParam

	return generateCode(myStepInfo, templateName, goTemplate, funcMap)
}

func stepImplementation(myStepInfo stepInfo, templateName, goTemplate string) []byte {
	funcMap := sprig.HermeticTxtFuncMap()
	funcMap["title"] = strings.Title
	funcMap["uniqueName"] = mustUniqName

	return generateCode(myStepInfo, templateName, goTemplate, funcMap)
}

func generateCode(dataObject interface{}, templateName, goTemplate string, funcMap template.FuncMap) []byte {
	tmpl, err := template.New(templateName).Funcs(funcMap).Parse(goTemplate)
	checkError(err)

	var generatedCode bytes.Buffer
	err = tmpl.Execute(&generatedCode, dataObject)
	checkError(err)

	return generatedCode.Bytes()
}

func longName(long string) string {
	l := strings.ReplaceAll(long, "`", "` + \"`\" + `")
	l = strings.TrimSpace(l)
	return l
}

func resourceFieldType(fieldType string) string {
	//TODO: clarify why fields are initialized with <nil> and tags are initialized with ''
	if len(fieldType) == 0 || fieldType == "<nil>" {
		return "string"
	}
	return fieldType
}

func golangName(name string) string {
	properName := strings.Replace(name, "Api", "API", -1)
	properName = strings.Replace(properName, "api", "API", -1)
	properName = strings.Replace(properName, "Url", "URL", -1)
	properName = strings.Replace(properName, "Id", "ID", -1)
	properName = strings.Replace(properName, "Json", "JSON", -1)
	properName = strings.Replace(properName, "json", "JSON", -1)
	properName = strings.Replace(properName, "Tls", "TLS", -1)
	return properName
}

func golangNameTitle(name string) string {
	return strings.Title(golangName(name))
}

func flagType(paramType string) string {
	var theFlagType string
	switch paramType {
	case "bool":
		theFlagType = "BoolVar"
	case "int":
		theFlagType = "IntVar"
	case "string":
		theFlagType = "StringVar"
	case "[]string":
		theFlagType = "StringSliceVar"
	default:
		fmt.Printf("Meta data type not set or not known: '%v'\n", paramType)
		os.Exit(1)
	}
	return theFlagType
}

func getStringSliceFromInterface(iSlice interface{}) []string {
	s := []string{}

	t, ok := iSlice.([]interface{})
	if ok {
		for _, v := range t {
			s = append(s, fmt.Sprintf("%v", v))
		}
	} else {
		s = append(s, fmt.Sprintf("%v", iSlice))
	}

	return s
}

func mustUniqName(list []config.StepParameters) ([]config.StepParameters, error) {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		names := []string{}
		dest := []config.StepParameters{}
		var item config.StepParameters
		for i := 0; i < l; i++ {
			item = l2.Index(i).Interface().(config.StepParameters)
			if !piperutils.ContainsString(names, item.Name) {
				names = append(names, item.Name)
				dest = append(dest, item)
			}
		}

		return dest, nil
	default:
		return nil, fmt.Errorf("Cannot find uniq on type %s", tp)
	}
}
