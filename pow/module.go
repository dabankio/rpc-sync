package pow

import "github.com/google/wire"

var Module = wire.NewSet(
	NewHandler,
	NewRepo,
)
