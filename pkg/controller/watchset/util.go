package watchset

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

func getDeploymentNamespacedName(o client.Object) types.NamespacedName {
	return types.NamespacedName{
		Name:      o.GetName(),
		Namespace: o.GetNamespace(),
	}
}

func getSecretNamespacedName(o *stvziov1.WatchSet) types.NamespacedName {
	return types.NamespacedName{
		Name:      *o.Spec.SecretName,
		Namespace: o.GetNamespace(),
	}
}
