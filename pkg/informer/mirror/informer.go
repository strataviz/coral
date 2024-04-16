package mirror

import (
	"context"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/credentials"
	"stvz.io/hashring"
)

// type MirrorCache struct {
// 	items map[client.ObjectKey]*stvziov1.Mirror
// 	sync.Mutex
// }

// func (m *MirrorCache) Set(mirror *stvziov1.Mirror) {
// 	if m.items == nil {
// 		m.items = make(map[client.ObjectKey]*stvziov1.Mirror)
// 	}

// 	m.items[client.ObjectKeyFromObject(mirror)] = mirror
// }

// func (m *MirrorCache) Remove(mirror *stvziov1.Mirror) {
// 	delete(m.items, client.ObjectKeyFromObject(mirror))
// }

// func (m *MirrorCache) Items() []*stvziov1.Mirror {
// 	m.Lock()
// 	defer m.Unlock()

// 	items := make([]*stvziov1.Mirror, 0, len(m.items))
// 	for _, mirror := range m.items {
// 		items = append(items, mirror)
// 	}

// 	return items
// }

type Informer struct {
	ServerRing *hashring.Ring
	Keyring    *credentials.Keyring
	Client     client.Client
	Mirrors    map[client.ObjectKey]*stvziov1.Mirror
	Namespace  string
	Labels     labels.Selector
	cache.Cache
}

func SetupWithManager(ctx context.Context, mgr ctrl.Manager, namespace string, lbs labels.Selector) (*Informer, error) {
	informer := &Informer{
		ServerRing: hashring.NewRing(1, nil),
		Keyring:    credentials.NewKeyring(mgr.GetClient()),
		Mirrors:    make(map[client.ObjectKey]*stvziov1.Mirror),
		Client:     mgr.GetClient(),
		Cache:      mgr.GetCache(),
		Namespace:  namespace,
		Labels:     lbs,
	}

	si, err := informer.GetInformerForKind(ctx, schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Secret",
	})
	if err != nil {
		return nil, err
	}

	_, err = si.AddEventHandler(&SecretHandler{
		Keyring: informer.Keyring,
	})
	if err != nil {
		return nil, err
	}

	mi, err := informer.GetInformerForKind(ctx, schema.GroupVersionKind{
		Group:   "stvz.io",
		Version: "v1",
		Kind:    "Mirror",
	})
	if err != nil {
		return nil, err
	}

	_, err = mi.AddEventHandler(&MirrorHandler{
		Keyring: informer.Keyring,
		Mirrors: informer.Mirrors,
	})
	if err != nil {
		return nil, err
	}

	pi, err := informer.GetInformerForKind(ctx, schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Pod",
	})
	if err != nil {
		return nil, err
	}

	_, err = pi.AddEventHandler(&PodHandler{
		Log:        log.FromContext(ctx).WithName("pod-handler"),
		ServerRing: informer.ServerRing,
		Namespace:  informer.Namespace,
		Labels:     informer.Labels,
	})
	if err != nil {
		return nil, err
	}

	return informer, nil
}

func (i *Informer) Start(ctx context.Context) error {
	if err := i.Cache.Start(ctx); err != nil {
		return err
	}

	if !i.Cache.WaitForCacheSync(ctx) {
		return ctx.Err()
	}

	return nil
}
