package create

import (
	"fmt"
	"os"
	"path/filepath"

	"kusionstack.io/kusion/pkg/cmd/stack/util"
)

type Options struct {
	Name        string
	ProjectDir  string
	StackDir    string
	RefStackDir string
	ConfigPath  string
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) Complete(args []string) error {
	name, err := util.GetNameFromArgs(args)
	if err != nil {
		return err
	}
	o.Name = name

	if o.ProjectDir == "" {
		projectDir, err := os.Getwd()
		if err != nil {
			return err
		}
		o.ProjectDir = projectDir
	}

	o.StackDir = filepath.Join(o.ProjectDir, o.Name)

	return nil
}

func (o *Options) Validate() error {
	if err := util.ValidateName(o.Name); err != nil {
		return err
	}

	if err := util.ValidateRefStackDir(o.RefStackDir); err != nil {
		return err
	}

	if err := util.ValidateConfigPath(o.ConfigPath); err != nil {
		return err
	}

	return nil
}

func (o *Options) Run() error {
	if err := util.CreateStackWithRefAndConfigFile(o.StackDir, o.RefStackDir, o.ConfigPath); err != nil {
		return err
	}

	fmt.Printf("Created stack '%s' successfully\n", o.Name)

	return nil
}
