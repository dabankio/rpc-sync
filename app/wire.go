//+build wireinject

package main

import (
	"bbcsyncer/infra"
	"bbcsyncer/reward"
	"bbcsyncer/sync"

	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(infra.Module, sync.Module, reward.Module,
		NewApp, NewJobs)
	return &App{}, nil
}
