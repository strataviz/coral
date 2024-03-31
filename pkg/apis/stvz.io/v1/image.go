package v1

import (
	"fmt"
)

func (i *ImageSpecImages) GetFullTaggedName(tag string) string {
	return fmt.Sprintf("%s:%s", *i.Name, tag)
}

// replace with the image hasher in util.
// func (i *Image) GetImageHash(tag string) string {
// 	hasher := md5.New()
// 	hasher.Write([]byte(i.GetFullTaggedName(tag)))
// 	return fmt.Sprintf("%x", hasher.Sum(nil))
// }

// TODO: Only used with testing.  remove
// func (i *Image) GetImageLabelKey(tag string) string {
// 	return fmt.Sprintf("image.stvz.io/%s", i.GetImageHash(tag))
// }

// TODO: This is only used in tests, remove.
func (i *Image) GetImages() []string {
	var images []string
	for _, image := range i.Spec.Images {
		for _, tag := range image.Tags {
			images = append(images, image.GetFullTaggedName(tag))
		}
	}
	return images
}
