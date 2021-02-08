package reward

import "github.com/google/wire"

var Module = wire.NewSet(
	NewCalc,
	NewRepo,
	NewHandler,
)
