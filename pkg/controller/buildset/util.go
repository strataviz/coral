package buildset

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getDeploymentNamespacedName(o client.Object) types.NamespacedName {
	return types.NamespacedName{
		Name:      o.GetName(),
		Namespace: o.GetNamespace(),
	}
}
