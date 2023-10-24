package builder

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

// Watcher is a struct that contains the logic for watching a github repository
// and sending build jobs to the controller for processing.  I still need to
// figure out how to keep state between the github event stream and what the
// watcher has already submitted jobs for.
//
// I think we may just be able to record the last event ID that we've seen and
// use that to determine if we've already submitted a job for that event.  From
// appearances, it looks like the event IDs are sequential, so we should be able
// skip events less than the last event ID we've seen.
type Watcher struct {
	PollInterval int
	Owner        string
	Repo         string
	Token        string
	Logger       logr.Logger
}

func (w *Watcher) Start(ctx context.Context) {
	w.intervalRun(ctx)

	timer := time.NewTicker(time.Duration(w.PollInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			w.Logger.Info("shutting down watcher", "repo", w.Repo)
			return
		case <-timer.C:
			w.intervalRun(ctx)
		}
	}
}

func (w *Watcher) intervalRun(ctx context.Context) {
	w.Logger.V(8).Info("fetching events")
	events, err := w.fetch(ctx)
	if err != nil {
		w.Logger.Error(err, "fetch failed")
		return
	}

	w.Logger.V(8).Info("processing events")
	for _, event := range events {
		w.Logger.V(8).Info("processing event", "event", event)
		err := w.process(ctx, event)
		if err != nil {
			w.Logger.Error(err, "process failed")
			return
		}
	}
}

func (w *Watcher) fetch(ctx context.Context) ([]*github.Event, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: w.Token},
	)
	auth := oauth2.NewClient(ctx, ts)
	client := github.NewClient(auth)

	w.Logger.V(8).Info("logs")

	events, resp, err := client.Activity.ListRepositoryEvents(ctx, w.Owner, w.Repo, nil)
	if err != nil {
		w.Logger.Error(err, "search failed")
		return nil, err
	}

	w.Logger.V(8).Info("github response", "events", events, "resp", resp)

	return events, nil
}

func (w *Watcher) process(ctx context.Context, events *github.Event) error {
	// We are interested in:
	// CreateEvent - in conjunction with Payload.RefType == "tag" (using Ref to access the tag name)
	// ReleaseEvent - in conjunction with Payload.Action == "published" (using Payload.Release.TagName to access the tag name)
	// PushEvent - in conjuntion with Payload.Ref as a wildcard match for "refs/heads/*"
	//
	// I should output to the event recorder and add the actor information to the output.
	return nil
}
