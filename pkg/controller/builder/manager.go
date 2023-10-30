package builder

import (
	"context"
	"sync"

	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
	"stvz.io/coral/pkg/repository"
	"stvz.io/coral/pkg/repository/github"
)

// Manager is a collection of repositories that will be watched by the builder.
type Manager struct {
	repositories map[string]repository.Repository

	wg       sync.WaitGroup
	syncOnce sync.Once
	sync.Mutex
}

// New returns a new Manager.
func NewManager() *Manager {
	return &Manager{
		repositories: make(map[string]repository.Repository),
	}
}

// AddRepositories adds multiple repositories that will be watched. We
// partition the repositories by name individually right now, but this
// won't scale well once a user starts watching potentially thousands of
// repositories and creating/managing jobs for each of them.  I think
// this is the easiest way to start out, however I think that when we
// start splitting out the builders into individual pods for scale we'll
// have to distribute the repos across the builders somehow.
func (m *Manager) AddWatches(ctx context.Context, token string, repo ...stvziov1.Watch) {
	for _, r := range repo {
		m.AddWatch(ctx, token, r)
	}
}

// AddWatch adds a watch for a repository.
func (m *Manager) AddWatch(ctx context.Context, token string, watch stvziov1.Watch) {
	m.Lock()
	defer m.Unlock()

	// TODO: we are just doing .Name right now, but if we add in different
	// git vendors, we'll also need to add the type as a prefix to prevent
	// collisions.
	if r, ok := m.repositories[watch.FullName()]; ok {
		r.Stop()
	}

	watcher := github.New(&repository.Opts{
		PollIntervalSeconds: *watch.On.PollIntervalSeconds,
		Owner:               *watch.Owner,
		Repo:                *watch.Repo,
		Token:               token,
	})
	m.repositories[watch.FullName()] = watcher

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		watcher.Start(ctx)
	}()
}

// DeleteRepository deletes a watched repository.
func (m *Manager) DeleteRepository(name string) {
	m.Lock()
	defer m.Unlock()

	if r, ok := m.repositories[name]; !ok {
		r.Stop()
		delete(m.repositories, name)
	}
}

// StopAll stops all of the repositories.
func (m *Manager) StopAll() {
	// I probably could use waitgroups here, but...
	m.Lock()
	defer m.Unlock()
	m.syncOnce.Do(func() {
		for _, r := range m.repositories {
			go r.Stop()
		}

		m.wg.Wait()
	})
}
