package inertia

import "github.com/google/wire"

var WireModule = wire.NewSet(
	wire.Struct(new(NewRendererParams), "*"),
	NewRenderer,
	NewSSRClient,
	LoadConfig,
)
