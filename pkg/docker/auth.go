package docker

import (
	"context"

	typesRegistry "github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/registry"
	"gopkg.in/yaml.v2"
)

// GetAuthConfig returns the AuthConfig for a Docker registry from the Docker credential helper
func (c *client) GetAuthConfig(ctx context.Context, registryURL string, checkCredentialsStore bool) (*typesRegistry.AuthConfig, error) {
	isDefaultRegistry, serverAddress, err := c.GetRegistryEndpoint(ctx, registryURL)
	if err != nil {
		return nil, err
	}

	return getDefaultAuthConfig(checkCredentialsStore, serverAddress, isDefaultRegistry)
}

// GetRegistryEndpoint retrieves the correct registry url
func (c *client) GetRegistryEndpoint(ctx context.Context, registryURL string) (bool, string, error) {
	authServer := c.getOfficialServer(ctx)
	if registryURL == "" || registryURL == "hub.docker.com" {
		registryURL = authServer
	}

	return registryURL == authServer, registryURL, nil
}

func (c *client) getOfficialServer(ctx context.Context) string {
	// The daemon `/info` endpoint informs us of the default registry being
	// used. This is essential in cross-platforms environment, where for
	// example a Linux client might be interacting with a Windows daemon, hence
	// the default registry URL might be Windows specific.
	serverAddress := registry.IndexServer
	if info, err := c.Info(ctx); err != nil {
		// Only report the warning if we're in debug mode to prevent nagging during engine initialization workflows
		// log.Warnf("Warning: failed to get default registry endpoint from daemon (%v). Using system default: %s", err, serverAddress)
	} else if info.IndexServerAddress == "" {
		// log.Warnf("Warning: Empty registry endpoint from daemon. Using system default: %s", serverAddress)
	} else {
		serverAddress = info.IndexServerAddress
	}

	return serverAddress
}

func getDefaultAuthConfig(checkCredStore bool, serverAddress string, isDefaultRegistry bool) (*typesRegistry.AuthConfig, error) {
	var authconfig typesRegistry.AuthConfig
	var err error

	if !isDefaultRegistry {
		serverAddress = registry.ConvertToHostname(serverAddress)
	}

	if checkCredStore {
		configfile, err := LoadDockerConfig()
		if configfile != nil && err == nil {
			authconfigOrig, err := configfile.GetAuthConfig(serverAddress)
			if err != nil {
				authconfig.ServerAddress = serverAddress
				return &authconfig, err
			}

			// convert
			err = convert(authconfigOrig, &authconfig)
			if err != nil {
				authconfig.ServerAddress = serverAddress
				return &authconfig, err
			}
		}
	}

	authconfig.ServerAddress = serverAddress
	return &authconfig, err
}

// convert converts the old object into the new object through yaml serialization / deserialization
func convert(old interface{}, new interface{}) error {
	o, err := yaml.Marshal(old)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(o, new); err != nil {
		return err
	}
	return nil
}
