package v1

import (
	"fmt"
	"hash"

	"k8s.io/apimachinery/pkg/util/dump"
)

func DeepCopyObject(hasher hash.Hash, obj interface{}) {
	hasher.Reset()
	fmt.Fprintf(hasher, "%v", dump.ForHash(obj))
}
