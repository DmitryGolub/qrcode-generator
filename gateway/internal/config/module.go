package config

import (
	"qrcodegen/gateway/config"

	"go.uber.org/fx"
)

var Module = fx.Module("config", fx.Provide(
	config.New,
))
