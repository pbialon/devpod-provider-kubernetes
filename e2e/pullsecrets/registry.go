package pullsecrets

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

type ContainerRegistry interface {
	Login()
	Logout()
	PublicImageName() string
	PrivateImageName() string
}

type Registry struct {
	Username string
	Password string
	Server   string
}

func (r *Registry) Login() {
	cmd := exec.Command(
		"docker",
		"login",
		r.Server,
		"--username", r.Username,
		"--password", r.Password,
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(fmt.Sprintf("failed to get stdin pipe: %v", err))
	}

	go func() {
		defer stdin.Close()
		_, _ = stdin.Write([]byte(r.Password))
	}()

	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("failed to login to Docker: %v, output: %s", err, output))
	}
}

func (r *Registry) Logout() {
	cmd := exec.Command("docker", "logout", r.Server)
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("failed to logout of Docker: %v, output: %s", err, output))
	}
}

type AWSRegistry struct{ Registry }

func (r *AWSRegistry) imageName(basename string) string {
	return path.Join(r.Server, basename)
}

func (r *AWSRegistry) PrivateImageName() string {
	return r.imageName("private-test-image")
}

func (r *AWSRegistry) PublicImageName() string {
	return r.imageName("public-test-image")
}

type GithubRegistry struct{ Registry }

func (r *GithubRegistry) imageName(basename string) string {
	return path.Join("ghcr.io/loft-sh/devpod-provider-kubernetes/", basename)
}

func (r *GithubRegistry) PrivateImageName() string {
	return r.imageName("private-test-image")
}

func (r *GithubRegistry) PublicImageName() string {
	return r.imageName("public-test-image")
}

type DockerHubRegistry struct{ Registry }

func (r *DockerHubRegistry) imageName(basename string) string {
	return path.Join(r.Username, basename)
}

func (r *DockerHubRegistry) PrivateImageName() string {
	return r.imageName("private-test-image")
}

func (r *DockerHubRegistry) PublicImageName() string {
	return r.imageName("public-test-image")
}

func RegistryFromEnv() (ContainerRegistry, error) {
	dockerUsername := os.Getenv("DOCKER_USERNAME")
	dockerPassword := os.Getenv("DOCKER_PASSWORD")
	containerRegistry := os.Getenv("CONTAINER_REGISTRY")

	if dockerUsername == "" || dockerPassword == "" {
		return nil, fmt.Errorf("DOCKER_USERNAME and/or DOCKER_PASSWORD are not set")
	}

	reg := &Registry{
		Username: dockerUsername,
		Password: dockerPassword,
		Server:   containerRegistry,
	}

	if strings.Contains(containerRegistry, "amazonaws.com") {
		return &AWSRegistry{*reg}, nil
	}
	if strings.Contains(containerRegistry, "ghcr.io") {
		return &GithubRegistry{*reg}, nil
	}
	if containerRegistry == "" || containerRegistry == "docker.io" {
		return &DockerHubRegistry{*reg}, nil
	}

	return nil, fmt.Errorf("unsupported registry: %s", containerRegistry)
}
