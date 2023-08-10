package database

// Database defines the delivery artifact of relational database service (rds)
// provided by different cloud vendors for the application.
type Database struct {
	// The specific cloud vendor that provides the rds.
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// The database engine to use.
	Engine string `json:"engine,omitempty" yaml:"engine,omitempty"`
	// The database engine version to use.
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	// The type of the rds instance.
	InstanceType string `json:"instanceType,omitempty" yaml:"instanceType,omitempty"`
	// The allocated storage size of the rds instance in GB.
	Size int `json:"size,omitempty" yaml:"size,omitempty"`
	// The edition of the rds instance.
	Category string `json:"category,omitempty" yaml:"category,omitempty"`
	// The operation account for the rds instance.
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	// The list of IP addresses allowed to access the rds instance.
	SecurityIPs []string `json:"securityIPs,omitempty" yaml:"securityIPs,omitempty"`
	// The virtual switch ID to launch alicloud rds instance.
	AlicloudVSwitchID string `json:"alicloudVSwitchID,omitempty" yaml:"alicloudVSwitchID,omitempty"`
	// Whether the host address of the rds instance will route through the public network or
	// the internal network of the cloud vendor.
	PrivateRouting bool `json:"privateRouting,omitempty" yaml:"privateRouting,omitempty"`
	// The diversified rds configuration items from different cloud vendors.
	ExtraMap map[string]string `json:"extraMap,omitempty" yaml:"extraMap,omitempty"`
}
