package repository

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
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
	EventRecorder       record.EventRecorder
}
