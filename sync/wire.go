//+build wireinject

package sync

import "github.com/google/wire"

func InitializeWorker() (*Worker, error) {
	wire.Build(NewWorker, NewRepo, NewPGDB, NewBBCClient, ParseConf)
	return &Worker{}, nil
}
