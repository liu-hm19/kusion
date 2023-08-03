package provider

// Provider is used to provision Terraform resources.
type Provider struct {
	// The complete provider address.
	URL string
	// The host of the provider registry.
	Host string
	// The namespace of the provider.
	Namespace string
	// The name of the provider.
	Name string
	// The version of the provider.
	Version string
}

// ProviderMeta records the meta info to use provider.
type ProviderMeta struct {
	// The region of provider resources.
	Region string `yaml:"region,omitempty" json:"region,omitempty"`
}
