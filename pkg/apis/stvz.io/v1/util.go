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
	"hash"
	"math/rand" // #nosec

	"github.com/containers/image/v5/transports/alltransports"
	"k8s.io/apimachinery/pkg/util/dump"
)

// NormalizeRepoTag converts an image reference into it's fully explicit form.
func NormalizeRepoTag(r, t string) (string, error) {
	fq := "docker://" + r
	if t != "" {
		fq += ":" + t
	}

	src, err := alltransports.ParseImageName(fq)
	if err != nil {
		return "", err
	}

	return src.DockerReference().String(), nil
}

func DeepCopyObject(hasher hash.Hash, obj interface{}) {
	hasher.Reset()
	fmt.Fprintf(hasher, "%v", dump.ForHash(obj))
}

func HashedImageLabelKey(name string) string {
	hash := ImageHasher(name)
	return fmt.Sprintf("%s/%s", LabelPrefix, hash)
}

func ImageHasher(name string) string {
	hasher := md5.New() // #nosec
	hasher.Write([]byte(name))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

func RandString(n int) string {
	b := make([]rune, n)
	chars := []rune("abcdefghijklmnopqrstuvwxyz1234567890")

	for i := range b {
		b[i] = chars[rand.Intn(len(chars))] // #nosec
	}

	return string(b)
}
