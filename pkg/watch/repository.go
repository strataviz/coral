package watch

import (
	"context"

	"github.com/go-logr/logr"
)

type Repository interface {
	Start(context.Context)
	Stop()
}

type Opts struct {
	PollIntervalSeconds int
	Owner               string
	Repo                string
	Token               string
	Logger              logr.Logger
}
