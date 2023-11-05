package watch

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

type Github struct {
	token               string
	log                 logr.Logger
	repo                string
	owner               string
	pollIntervalSeconds int

	stopChan chan struct{}
	stopOnce sync.Once
	sync.Mutex
}

func NewGithubRepo(opts *Opts) Repository {
	return &Github{
		token:               opts.Token,
		log:                 opts.Logger,
		repo:                opts.Repo,
		owner:               opts.Owner,
		pollIntervalSeconds: opts.PollIntervalSeconds,
		stopChan:            make(chan struct{}),
	}
}

func (g *Github) Start(ctx context.Context) {
	g.intervalRun(ctx)

	timer := time.NewTicker(time.Duration(g.pollIntervalSeconds) * time.Second)
	for {
		select {
		case <-g.stopChan:
			g.log.Info("shutting down watcher", "repo", g.repo)
			return
		case <-timer.C:
			g.intervalRun(ctx)
		}
	}
}

func (g *Github) Stop() {
	g.Lock()
	defer g.Unlock()
	g.stopOnce.Do(func() {
		close(g.stopChan)
	})
}

func (g *Github) intervalRun(ctx context.Context) {
	g.log.V(8).Info("fetching events")
	events, err := g.fetch(ctx)
	if err != nil {
		g.log.Error(err, "fetch failed")
		return
	}

	g.log.V(8).Info("processing events")
	for _, event := range events {
		g.log.V(8).Info("processing event", "event", event)
		err := g.process(ctx, event)
		if err != nil {
			g.log.Error(err, "process failed")
			return
		}
	}
}

func (g *Github) fetch(ctx context.Context) ([]*github.Event, error) {
	g.Lock()
	defer g.Unlock()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.token},
	)
	auth := oauth2.NewClient(ctx, ts)
	client := github.NewClient(auth)

	g.log.V(8).Info("logs")

	events, resp, err := client.Activity.ListRepositoryEvents(ctx, g.owner, g.repo, nil)
	if err != nil {
		g.log.Error(err, "search failed")
		return nil, err
	}

	g.log.V(8).Info("github response", "events", events, "resp", resp)

	return events, nil
}

func (g *Github) process(ctx context.Context, events *github.Event) error {
	g.Lock()
	defer g.Unlock()

	// We are interested in:
	// CreateEvent - in conjunction with Payload.RefType == "tag" (using Ref to access the tag name)
	// ReleaseEvent - in conjunction with Payload.Action == "published" (using Payload.Release.TagName to access the tag name)
	// PushEvent - in conjuntion with Payload.Ref as a wildcard match for "refs/heads/*"
	//
	// I should output to the event recorder and add the actor information to the output.
	return nil
}
