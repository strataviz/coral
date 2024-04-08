// Copyright 2024 Coral Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"crypto/md5" // #nosec
	"fmt"
)

const (
	LabelPrefix = "image.stvz.io"
)

// TODO: get rid of this, it's only used in tests at the moment.
func (i *RepositorySpec) GetRepoTag(tag string) string {
	return fmt.Sprintf("%s:%s", *i.Name, tag)
}

func (i *RepositorySpec) GetLabel(tag string) string {
	hasher := md5.New() // #nosec
	hasher.Write([]byte(i.GetRepoTag(tag)))
	return fmt.Sprintf("%s/%x", LabelPrefix, hasher.Sum(nil))
}

// TODO: This is only used in tests, remove.
func (i *Image) GetImages() []string {
	var images []string
	for _, image := range i.Spec.Repositories {
		for _, tag := range image.Tags {
			images = append(images, image.GetRepoTag(tag))
		}
	}
	return images
}

func (i *Image) GetStatusData() []ImageData {
	data := make([]ImageData, 0)
	for _, image := range i.Spec.Repositories {
		for _, tag := range image.Tags {
			data = append(data, ImageData{
				Name:  image.GetRepoTag(tag),
				Label: image.GetLabel(tag),
			})
		}
	}

	return data
}
