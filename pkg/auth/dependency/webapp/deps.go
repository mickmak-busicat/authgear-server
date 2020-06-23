package webapp

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(ValidateProviderImpl), "*"),
	wire.Struct(new(RenderProviderImpl), "*"),
	wire.Struct(new(StateStoreImpl), "*"),
	wire.Bind(new(StateStore), new(*StateStoreImpl)),
	wire.Struct(new(StateProviderImpl), "*"),
	wire.Bind(new(StateProvider), new(*StateProviderImpl)),
	wire.Struct(new(URLProvider), "*"),

	wire.Struct(new(CSPMiddleware), "*"),
	wire.Struct(new(CSRFMiddleware), "*"),
	wire.Struct(new(AuthEntryPointMiddleware), "*"),
	wire.Struct(new(StateMiddleware), "*"),
)
