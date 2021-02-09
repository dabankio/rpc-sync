//+build wireinject

package main

import (
	"bbcsyncer/infra"
	"bbcsyncer/pow"
	"bbcsyncer/reward"
	"bbcsyncer/sync"

	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(
		infra.Module,
		sync.Module,
		reward.Module,
		pow.Module,
		NewApp, NewJobs, NewRouter)
	return &App{}, nil
}
