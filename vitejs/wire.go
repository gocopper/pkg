package vitejs

import "github.com/google/wire"

var WireModule = wire.NewSet(
	NewAssets,
	LoadConfig,
)
