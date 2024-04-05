package v1

import (
	"crypto/md5" // #nosec
	"fmt"
	"hash"
	"math/rand" // #nosec

	"k8s.io/apimachinery/pkg/util/dump"
)

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
