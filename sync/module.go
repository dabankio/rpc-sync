package sync

import "github.com/google/wire"

var Module = wire.NewSet(
	NewWorker,
	NewRepo,
)
