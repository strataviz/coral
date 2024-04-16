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
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func GetRepositoryTags(ctx context.Context, auth *runtime.AuthConfig, registry string, name string) ([]string, error) {
	name = registry + "/" + name

	ref, err := alltransports.ParseImageName(name)
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

func Copy(ctx context.Context, auth *runtime.AuthConfig, src string, dest string) error {
	log := log.FromContext(ctx)

	sref, err := alltransports.ParseImageName(src)
	if err != nil {
		log.Error(err, "failed to parse source image name", "name", src)
		return err
	}

	dref, err := alltransports.ParseImageName(dest)
	if err != nil {
		log.Error(err, "failed to parse dest image name", "name", dest)
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

	sctx := SystemContext(auth, true)
	dctx := SystemContext(auth, false)

	_, err = copy.Image(ctx, pctx, dref, sref, &copy.Options{
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
	if auth == nil {
		auth = &runtime.AuthConfig{}
	}

	var skipTLSVerify types.OptionalBool
	if !tlsVerify {
		skipTLSVerify = types.OptionalBoolTrue
	} else {
		skipTLSVerify = types.OptionalBoolFalse
	}

	return &types.SystemContext{
		DockerInsecureSkipTLSVerify: skipTLSVerify,
		DockerAuthConfig: &types.DockerAuthConfig{
			Username:      auth.Username,
			Password:      auth.Password,
			IdentityToken: auth.IdentityToken,
		},
		DockerBearerRegistryToken: auth.RegistryToken,
	}
}
