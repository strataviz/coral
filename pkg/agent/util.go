package agent

import (
	"context"

	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

func GetImageIdentifiers(ctx context.Context, ims runtime.ImageServiceClient) (map[string]string, error) {
	resp, err := ims.ListImages(ctx, &runtime.ListImagesRequest{})
	if err != nil {
		return nil, err
	}

	ids := make(map[string]string)
	for _, img := range resp.Images {
		for _, tag := range img.RepoTags {
			ids[tag] = img.Id
		}
	}

	return ids, nil
}

func ImageMap(ctx context.Context, ims runtime.ImageServiceClient) (map[string]string, error) {
	resp, err := ims.ListImages(ctx, &runtime.ListImagesRequest{})
	if err != nil {
		return nil, err
	}

	tags := make(map[string]string)
	for _, img := range resp.Images {
		for _, tag := range img.RepoTags {
			tags[tag] = stvziov1.HashedImageLabelKey(tag)
		}
	}

	return tags, nil
}
