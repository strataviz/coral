package v1

import (
	"crypto/md5" // #nosec
	"fmt"
)

const (
	LabelPrefix = "image.stvz.io"
)

// TODO: get rid of this, it's only used in tests at the moment.
func (i *ImageSpecImages) GetRepoTag(tag string) string {
	return fmt.Sprintf("%s:%s", *i.Name, tag)
}

func (i *ImageSpecImages) GetLabel(tag string) string {
	hasher := md5.New() // #nosec
	hasher.Write([]byte(i.GetRepoTag(tag)))
	return fmt.Sprintf("%s/%x", LabelPrefix, hasher.Sum(nil))
}

// TODO: This is only used in tests, remove.
func (i *Image) GetImages() []string {
	var images []string
	for _, image := range i.Spec.Images {
		for _, tag := range image.Tags {
			images = append(images, image.GetRepoTag(tag))
		}
	}
	return images
}

func (i *Image) GetStatusData() []ImageData {
	data := make([]ImageData, 0)
	for _, image := range i.Spec.Images {
		for _, tag := range image.Tags {
			data = append(data, ImageData{
				Name:  image.GetRepoTag(tag),
				Label: image.GetLabel(tag),
			})
		}
	}

	return data
}
