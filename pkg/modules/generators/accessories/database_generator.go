package database

import (
	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	database "kusionstack.io/kusion/pkg/modules/inputs/accessories"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

var _ modules.Generator = &databaseGenerator{}

type databaseGenerator struct {
	project  *apiv1.Project
	stack    *apiv1.Stack
	appName  string
	workload *workload.Workload
	database map[string]*database.Database
	ws       *apiv1.Workspace
}

func NewDatabaseGenerator(
	project *apiv1.Project,
	stack *apiv1.Stack,
	appName string,
	workload *workload.Workload,
	database map[string]*database.Database,
	ws *apiv1.Workspace,
) (modules.Generator, error) {
	// TODO: implement me.
	return nil, nil
}

func NewDatabaseGeneratorFunc(
	project *apiv1.Project,
	stack *apiv1.Stack,
	appName string,
	workload *workload.Workload,
	database map[string]*database.Database,
	ws *apiv1.Workspace,
) modules.NewGeneratorFunc {
	// TODO: implement me.
	return nil
}

func (g *databaseGenerator) Generate(spec *apiv1.Intent) error {
	// TODO: implement me.
	return nil
}
