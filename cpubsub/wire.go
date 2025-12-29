package cpubsub

import "github.com/google/wire"

// WireModuleLocal defines the wire module for providing LocalPubSub
var WireModuleLocal = wire.NewSet(
	NewLocalPubSub,
)

var WireModuleRedis = wire.NewSet(
	NewRedisPubSub,
)
