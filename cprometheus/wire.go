package cprometheus

import "github.com/google/wire"

var WireModule = wire.NewSet(
	NewRouter,
	LoadConfig,
)
