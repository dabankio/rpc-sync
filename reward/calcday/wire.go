//+build wireinject

package main

import (
	"bbcsyncer/infra"
	"bbcsyncer/reward"
	"bbcsyncer/sync"

	"github.com/google/wire"
)

func Initialize() (*reward.Calc, error) {
	wire.Build(sync.Module, infra.Module, reward.Module)
	return &reward.Calc{}, nil
}
