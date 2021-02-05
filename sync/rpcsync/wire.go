//+build wireinject

package main

import (
	"bbcsyncer/infra"
	"bbcsyncer/sync"

	"github.com/google/wire"
)

func InitializeWorker() (*sync.Worker, error) {
	wire.Build(sync.NewWorker, sync.NewRepo, infra.NewPGDB, infra.NewBBCClient, infra.ParseConf)
	return &sync.Worker{}, nil
}
