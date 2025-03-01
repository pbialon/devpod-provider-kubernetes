package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/loft-sh/devpod-provider-kubernetes/pkg/docker"
	perrors "github.com/pkg/errors"
	k8sv1 "k8s.io/api/core/v1"
)

func (k *KubernetesDriver) EnsurePullSecret(
	ctx context.Context,
	pullSecretName string,
	dockerImage string,
) (bool, error) {
	k.Log.Debugf("Ensure pull secrets")

	host, err := GetRegistryFromImageName(dockerImage)
	if err != nil {
		return false, fmt.Errorf("get registry from image name: %w", err)
	}

	dockerCredentials, err := docker.GetAuthConfig(host)
	if err != nil || dockerCredentials == nil || dockerCredentials.Username == "" || dockerCredentials.Secret == "" {
		k.Log.Debugf("Couldn't retrieve credentials for registry: %s", host)
		return false, nil
	}

	if k.secretExists(ctx, pullSecretName) {
		if !k.shouldRecreateSecret(ctx, dockerCredentials, pullSecretName, host) {
			k.Log.Debugf("Pull secret '%s' already exists and is up to date", pullSecretName)
			return false, nil
		}

		k.Log.Debugf("Pull secret '%s' already exists, but is outdated. Recreating...", pullSecretName)
		err := k.DeletePullSecret(ctx, pullSecretName)
		if err != nil {
			return false, err
		}
	}

	err = k.createPullSecret(ctx, pullSecretName, dockerCredentials)
	if err != nil {
		return false, err
	}

	k.Log.Infof("Pull secret '%s' created", pullSecretName)
	return true, nil
}

func (k *KubernetesDriver) ReadSecretContents(
	ctx context.Context,
	pullSecretName string,
	host string,
) (string, error) {
	args := []string{
		"get",
		"secret",
		pullSecretName,
		"-o", "json",
	}

	out, err := k.buildCmd(ctx, args).CombinedOutput()
	if err != nil {
		return "", perrors.Wrapf(err, "delete pull secret: %s", string(out))
	}

	var secret k8sv1.Secret
	err = json.Unmarshal(out, &secret)
	if err != nil {
		return "", perrors.Wrap(err, "unmarshal secret")
	}

	return DecodeAuthTokenFromPullSecret(secret, host)
}

func (k *KubernetesDriver) DeletePullSecret(
	ctx context.Context,
	pullSecretName string) error {
	if !k.secretExists(ctx, pullSecretName) {
		return nil
	}

	args := []string{
		"delete",
		"secret",
		pullSecretName,
	}

	out, err := k.buildCmd(ctx, args).CombinedOutput()
	if err != nil {
		return perrors.Wrapf(err, "delete pull secret: %s", string(out))
	}

	return nil
}

func (k *KubernetesDriver) shouldRecreateSecret(ctx context.Context, dockerCredentials *docker.Credentials, pullSecretName, host string) bool {
	existingAuthToken, err := k.ReadSecretContents(ctx, pullSecretName, host)
	if err != nil {
		return true
	}
	return existingAuthToken != dockerCredentials.AuthToken()
}

func (k *KubernetesDriver) secretExists(
	ctx context.Context,
	pullSecretName string,
) bool {
	args := []string{
		"get",
		"secret",
		pullSecretName,
	}

	_, err := k.buildCmd(ctx, args).CombinedOutput()
	if err != nil {
		return false
	}
	return true
}

func (k *KubernetesDriver) createPullSecret(
	ctx context.Context,
	pullSecretName string,
	dockerCredentials *docker.Credentials,
) error {

	authToken := dockerCredentials.AuthToken()
	email := "noreply@loft.sh"

	encodedSecretData, err := PreparePullSecretData(dockerCredentials.ServerURL, authToken, email)
	if err != nil {
		return perrors.Wrap(err, "prepare pull secret data")
	}

	args := []string{
		"create",
		"secret",
		"generic",
		pullSecretName,
		"--type", string(k8sv1.SecretTypeDockerConfigJson),
		"--from-literal", encodedSecretData,
	}

	out, err := k.buildCmd(ctx, args).CombinedOutput()
	if err != nil {
		return perrors.Wrapf(err, "create pull secret: %s", string(out))
	}

	return nil
}
