package agent

import (
	"context"

	runtime "k8s.io/cri-api/pkg/apis/runtime/v1"
	"stvz.io/coral/pkg/util"
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

func ImageHashMap(ctx context.Context, ims runtime.ImageServiceClient) (map[string]string, error) {
	resp, err := ims.ListImages(ctx, &runtime.ListImagesRequest{})
	if err != nil {
		return nil, err
	}

	tags := make(map[string]string)
	for _, img := range resp.Images {
		for _, tag := range img.RepoTags {
			key := util.HashedImageLabelKey(tag)
			tags[key] = tag
		}
	}

	return tags, nil
}
