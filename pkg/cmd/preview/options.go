package preview

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"

	"github.com/pkg/errors"

	compilecmd "kusionstack.io/kusion/pkg/cmd/compile"
	"kusionstack.io/kusion/pkg/cmd/spec"
	"kusionstack.io/kusion/pkg/engine/backend"
	"kusionstack.io/kusion/pkg/engine/operation"
	opsmodels "kusionstack.io/kusion/pkg/engine/operation/models"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/generator"
	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/status"
	"kusionstack.io/kusion/pkg/util/pretty"
)

const jsonOutput = "json"

type Options struct {
	compilecmd.Options
	Flags
	backend.BackendOps
}

type Flags struct {
	Operator     string
	Detail       bool
	All          bool
	NoStyle      bool
	Output       string
	SpecFile     string
	IgnoreFields []string
}

func NewPreviewOptions() *Options {
	return &Options{
		Options: *compilecmd.NewCompileOptions(),
	}
}

func (o *Options) Complete(args []string) {
	o.Options.Complete(args)
}

func (o *Options) Validate() error {
	if err := o.Options.Validate(); err != nil {
		return err
	}
	if o.Output != "" && o.Output != jsonOutput {
		return errors.New("invalid output type, supported types: json")
	}
	if err := o.ValidateSpecFile(); err != nil {
		return err
	}
	return nil
}

func (o *Options) ValidateSpecFile() error {
	if o.SpecFile == "" {
		return nil
	}

	// calculate the absolute path of the specFile
	var absSF string
	if o.WorkDir == "" {
		absSF, _ = filepath.Abs(o.SpecFile)
	} else if filepath.IsAbs(o.SpecFile) {
		absSF = o.SpecFile
	} else {
		absSF = filepath.Join(o.WorkDir, o.SpecFile)
	}

	fi, err := os.Stat(absSF)
	if err != nil {
		if os.IsNotExist(err) {
			err = fmt.Errorf("spec file not exist")
		}
		return err
	}

	if fi.IsDir() || !fi.Mode().IsRegular() {
		return fmt.Errorf("spec file must be a regular file")
	}

	// calculate the relative path between absWD and absSF,
	// if absSF is not located in the directory or subdirectory specified by absWD,
	// an error will be returned
	absWD, _ := filepath.Abs(o.WorkDir)
	rel, err := filepath.Rel(absWD, absSF)
	if err != nil {
		return err
	}
	if rel[:3] == ".."+string(filepath.Separator) {
		return fmt.Errorf("the spec file must be located in the working directory or its subdirectories")
	}

	// set the spec file to the absolute path for further processing
	o.SpecFile = absSF
	return nil
}

func (o *Options) Run() error {
	// Set no style
	if o.NoStyle || o.Output == jsonOutput {
		pterm.DisableStyling()
		pterm.DisableColor()
	}
	// Parse project and stack of work directory
	project, stack, err := projectstack.DetectProjectAndStack(o.WorkDir)
	if err != nil {
		return err
	}

	options := &generator.Options{
		IsKclPkg:    o.IsKclPkg,
		WorkDir:     o.WorkDir,
		Filenames:   o.Filenames,
		Settings:    o.Settings,
		Arguments:   o.Arguments,
		Overrides:   o.Overrides,
		DisableNone: o.DisableNone,
		OverrideAST: o.OverrideAST,
		NoStyle:     o.NoStyle,
	}

	// Generate Intent
	var sp *models.Intent
	if o.SpecFile != "" {
		sp, err = spec.GenerateSpecFromFile(o.SpecFile)
	} else if o.Output == jsonOutput {
		sp, err = spec.GenerateSpec(options, project, stack)
	} else {
		sp, err = spec.GenerateSpecWithSpinner(options, project, stack)
	}
	if err != nil {
		return err
	}

	// return immediately if no resource found in stack
	// todo: if there is no resource, should still do diff job; for now, if output is json format, there is no hint
	if sp == nil || len(sp.Resources) == 0 {
		if o.Output != jsonOutput {
			fmt.Println(pretty.GreenBold("\nNo resource found in this stack."))
		}
		return nil
	}

	// Get state storage from backend config to manage state
	stateStorage, err := backend.BackendFromConfig(project.Backend, o.BackendOps, o.WorkDir)
	if err != nil {
		return err
	}

	// Compute changes for preview
	changes, err := Preview(o, stateStorage, sp, project, stack)
	if err != nil {
		return err
	}

	if o.Output == jsonOutput {
		var previewChanges []byte
		previewChanges, err = json.Marshal(changes)
		if err != nil {
			return fmt.Errorf("json marshal preview changes failed as %w", err)
		}
		fmt.Println(string(previewChanges))
		return nil
	}

	if changes.AllUnChange() {
		fmt.Println("All resources are reconciled. No diff found")
		return nil
	}

	// Summary preview table
	changes.Summary(os.Stdout)

	// Detail detection
	if o.Detail {
		for {
			target, err := changes.PromptDetails()
			if err != nil {
				return err
			}
			if target == "" { // Cancel option
				break
			}
			changes.OutputDiff(target)
		}
	}

	return nil
}

// The Preview function calculates the upcoming actions of each resource
// through the execution Kusion Engine, and you can customize the
// runtime of engine and the state storage through `runtime` and
// `storage` parameters.
//
// Example:
//
//	o := NewPreviewOptions()
//	stateStorage := &states.FileSystemState{
//	    Path: filepath.Join(o.WorkDir, states.KusionState)
//	}
//	kubernetesRuntime, err := runtime.NewKubernetesRuntime()
//	if err != nil {
//	    return err
//	}
//
//	changes, err := Preview(o, kubernetesRuntime, stateStorage,
//	    planResources, project, stack, os.Stdout)
//	if err != nil {
//	    return err
//	}
func Preview(
	o *Options,
	storage states.StateStorage,
	planResources *models.Intent,
	project *projectstack.Project,
	stack *projectstack.Stack,
) (*opsmodels.Changes, error) {
	log.Info("Start compute preview changes ...")

	// Construct the preview operation
	pc := &operation.PreviewOperation{
		Operation: opsmodels.Operation{
			OperationType: opsmodels.ApplyPreview,
			Stack:         stack,
			StateStorage:  storage,
			IgnoreFields:  o.IgnoreFields,
			ChangeOrder:   &opsmodels.ChangeOrder{StepKeys: []string{}, ChangeSteps: map[string]*opsmodels.ChangeStep{}},
		},
	}

	log.Info("Start call pc.Preview() ...")

	// parse cluster in arguments
	cluster := o.Arguments["cluster"]
	rsp, s := pc.Preview(&operation.PreviewRequest{
		Request: opsmodels.Request{
			Tenant:   project.Tenant,
			Project:  project,
			Stack:    stack,
			Operator: o.Operator,
			Spec:     planResources,
			Cluster:  cluster,
		},
	})
	if status.IsErr(s) {
		return nil, fmt.Errorf("preview failed.\n%s", s.String())
	}

	return opsmodels.NewChanges(project, stack, rsp.Order), nil
}
