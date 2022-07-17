package livewire

import "github.com/google/wire"

var WireModule = wire.NewSet(
	NewRenderer,
	ProvideHTMLRenderFuncLivewire,

	wire.Struct(new(NewRouterParams), "*"),
	NewRouter,
)
