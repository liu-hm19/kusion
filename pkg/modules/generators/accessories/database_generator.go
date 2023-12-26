package database

import (
	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/generators/accessories/mysql"
	"kusionstack.io/kusion/pkg/modules/generators/accessories/postgres"
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
	return &databaseGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
		database: database,
		ws:       ws,
	}, nil
}

func NewDatabaseGeneratorFunc(
	project *apiv1.Project,
	stack *apiv1.Stack,
	appName string,
	workload *workload.Workload,
	database map[string]*database.Database,
	ws *apiv1.Workspace,
) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewDatabaseGenerator(project, stack, appName, workload, database, ws)
	}
}

func (g *databaseGenerator) Generate(spec *apiv1.Intent) error {
	if spec.Resources == nil {
		spec.Resources = make(apiv1.Resources, 0)
	}

	if len(g.database) > 0 {
		var gfs []modules.NewGeneratorFunc

		for dbKey, db := range g.database {
			switch db.Header.Type {
			case database.TypeMySQL:
				gfs = append(gfs, mysql.NewMySQLGeneratorFunc(
					g.project, g.stack, g.appName, g.workload, db.MySQL, g.ws, dbKey))
			case database.TypePostgreSQL:
				gfs = append(gfs, postgres.NewPostgresGeneratorFunc(
					g.project, g.stack, g.appName, g.workload, db.PostgreSQL, g.ws, dbKey))
			}
		}

		if err := modules.CallGenerators(spec, gfs...); err != nil {
			return err
		}
	}
	return nil
}
