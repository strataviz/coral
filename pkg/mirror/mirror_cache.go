package mirror

import (
	"sync"

	"k8s.io/apimachinery/pkg/types"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

type MirrorCache struct {
	items map[types.NamespacedName]*stvziov1.Mirror
	sync.Mutex
}

func NewMirrorCache() *MirrorCache {
	return &MirrorCache{
		items: make(map[types.NamespacedName]*stvziov1.Mirror),
	}
}

func (m *MirrorCache) Add(mirror *stvziov1.Mirror) {
	m.Lock()
	defer m.Unlock()
	key := types.NamespacedName{
		Name:      mirror.Name,
		Namespace: mirror.Namespace,
	}
	m.items[key] = mirror
}

func (m *MirrorCache) Remove(mirror *stvziov1.Mirror) {
	m.Lock()
	defer m.Unlock()
	key := types.NamespacedName{
		Name:      mirror.Name,
		Namespace: mirror.Namespace,
	}
	delete(m.items, key)
}

func (m *MirrorCache) Get() map[types.NamespacedName]*stvziov1.Mirror {
	m.Lock()
	defer m.Unlock()

	return m.items
}
