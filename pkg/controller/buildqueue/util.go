package buildqueue

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TODO: combine some of these into functions that can be used in both
// the observer and reconciler.

func getStatefulSetNamespacedName(o client.Object) types.NamespacedName {
	return types.NamespacedName{
		Name:      fmt.Sprintf("%s-nats", o.GetName()),
		Namespace: o.GetNamespace(),
	}
}

func getServiceNamespacedName(o client.Object) types.NamespacedName {
	return types.NamespacedName{
		Name:      fmt.Sprintf("%s-nats", o.GetName()),
		Namespace: o.GetNamespace(),
	}
}

func getHeadlessServiceNamespacedName(o client.Object) types.NamespacedName {
	return types.NamespacedName{
		Name:      fmt.Sprintf("%s-nats-headless", o.GetName()),
		Namespace: o.GetNamespace(),
	}
}

func getConfigMapNamespacedName(o client.Object) types.NamespacedName {
	return types.NamespacedName{
		Name:      fmt.Sprintf("%s-nats", o.GetName()),
		Namespace: o.GetNamespace(),
	}
}
