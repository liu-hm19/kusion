package postgres

import (
	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs/accessories/postgres"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

var _ modules.Generator = &postgresGenerator{}

// postgresGenerator implements the modules.Generator interface.
type postgresGenerator struct {
	project  *apiv1.Project
	stack    *apiv1.Stack
	appName  string
	workload *workload.Workload
	postgres *postgres.PostgreSQL
	ws       *apiv1.Workspace
}

// NewPostgreGenerator returns a new generator for postgres database.
func NewPostgresGenerator(
	project *apiv1.Project,
	stack *apiv1.Stack,
	appName string,
	workload *workload.Workload,
	postgres *postgres.PostgreSQL,
	ws *apiv1.Workspace,
) (modules.Generator, error) {
	// TODO: implement me.
	return nil, nil
}

// NewPostgresGeneratorFunc returns a new generator function for
// generating a new postgres database.
func NewPostgresGeneratorFunc(
	project *apiv1.Project,
	stack *apiv1.Stack,
	appName string,
	workload *workload.Workload,
	postgres *postgres.PostgreSQL,
	ws *apiv1.Workspace,
) modules.NewGeneratorFunc {
	// TODO: implement me.
	return nil
}

// Generate generates a new postgres database instance for the workload.
func (g *postgresGenerator) Generate(spec *apiv1.Intent) error {
	// TODO: implement me.
	return nil
}
