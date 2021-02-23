package infra

import "github.com/google/wire"

var Module = wire.NewSet(
	NewPGDB,
	ParseConf,
	NewBBCClient,
	NewSched,
	NewRepo,
)
