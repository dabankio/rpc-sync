//+build wireinject

package main

import (
	"bbcsyncer/infra"
	"bbcsyncer/sync"

	"github.com/google/wire"
)

func InitializeWorker() (*sync.Worker, error) {
	wire.Build(sync.Module, infra.Module)
	return &sync.Worker{}, nil
}
