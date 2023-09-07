package pullsecrets

// PullSecretConfig defines a pull secret that should be created by DevSpace
type PullSecretConfig struct {
	// Name is the pull secret name to deploy
	Name string `yaml:"name,omitempty" json:"name,omitempty"`

	// The registry to create the image pull secret for.
	// Empty string == docker hub
	// e.g. gcr.io
	Registry string `yaml:"registry,omitempty" json:"registry" jsonschema:"required"`

	// The username of the registry. If this is empty, devspace will try
	// to receive the auth data from the local docker
	Username string `yaml:"username,omitempty" json:"username,omitempty"`

	// The password to use for the registry. If this is empty, devspace will
	// try to receive the auth data from the local docker
	Password string `yaml:"password,omitempty" json:"password,omitempty"`

	// The optional email to use
	Email string `yaml:"email,omitempty" json:"email,omitempty"`

	// The secret to create
	Secret string `yaml:"secret,omitempty" json:"secret,omitempty"`

	// The service account to add the secret to
	ServiceAccounts []string `yaml:"serviceAccounts,omitempty" json:"serviceAccounts,omitempty"`
}
