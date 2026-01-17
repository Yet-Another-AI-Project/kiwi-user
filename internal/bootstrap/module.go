package bootstrap

import "go.uber.org/fx"

var Module = fx.Provide(
	NewBootStrap,
)
