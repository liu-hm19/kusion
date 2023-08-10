package component

import (
	"kusionstack.io/kusion/pkg/models/appconfiguration/accessories/database"
	"kusionstack.io/kusion/pkg/models/appconfiguration/component/workload"
)

type Component struct {
	Job                *workload.Job                `yaml:"job,omitempty" json:"job,omitempty"`
	LongRunningService *workload.LongRunningService `yaml:"longRunningService,omitempty" json:"longRunningService,omitempty"`

	// List of Workload supporting accessory. Accessory defines various runtime capabilities and operation functionalities.
	// Database defines the relational database service provided by cloud vendor.
	Database database.Database `yaml:"database,omitempty" json:"database,omitempty"`

	// Variables for Day-2 Operation.

	// Variables for Workload scheduling.

	// Other metadata info

	// Labels and annotations can be used to attach arbitrary metadata as key-value pairs to resources.
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}
