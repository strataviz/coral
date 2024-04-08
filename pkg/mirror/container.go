package mirror

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func GetRepositoryTags(ctx context.Context, auth *runtime.AuthConfig, registry string, name string) ([]string, error) {
	name = registry + "/" + name

	ref, err := alltransports.ParseImageName("docker://" + name)
	if err != nil {
		return nil, err
	}

	// TODO: make tls verify configurable
	sys := SystemContext(auth, false)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	l, err := docker.GetRepositoryTags(ctx, sys, ref)
	if err != nil {
		if strings.Contains(err.Error(), "repository name not known to registry") {
			return []string{}, nil
		}
		return nil, err
	}

	return l, nil
}

func Copy(ctx context.Context, auth *runtime.AuthConfig, registry string, name string) error {
	src, err := alltransports.ParseImageName("docker://" + name)
	if err != nil {
		return err
	}

	dest, err := alltransports.ParseImageName("docker://" + registry + "/" + name)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	pctx, err := signature.NewPolicyContext(&signature.Policy{
		Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()},
	})
	if err != nil {
		return err
	}

	// TODO: make tls verify configurable
	sctx := SystemContext(auth, true)
	dctx := SystemContext(auth, false)

	_, err = copy.Image(ctx, pctx, dest, src, &copy.Options{
		RemoveSignatures:   true,
		ReportWriter:       io.Discard,
		ImageListSelection: copy.CopySystemImage,
		SourceCtx:          sctx,
		DestinationCtx:     dctx,
		PreserveDigests:    true,
	})

	return err
}

func SystemContext(auth *runtime.AuthConfig, tlsVerify bool) *types.SystemContext {
	var ob types.OptionalBool
	if tlsVerify {
		ob = types.OptionalBoolTrue
	} else {
		ob = types.OptionalBoolFalse
	}

	return &types.SystemContext{
		DockerInsecureSkipTLSVerify: ob,
		DockerAuthConfig: &types.DockerAuthConfig{
			Username:      auth.Username,
			Password:      auth.Password,
			IdentityToken: auth.IdentityToken,
		},
		DockerBearerRegistryToken: auth.RegistryToken,
	}
}
