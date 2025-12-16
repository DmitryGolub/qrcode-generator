package main

import (
	"qrcodegen/gateway/internal/client/grpc"
	"qrcodegen/gateway/internal/config"
	"qrcodegen/gateway/internal/delivery"
	"qrcodegen/gateway/internal/logger"
	"qrcodegen/gateway/internal/middleware"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	fx.New(
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{
				Logger: log.WithOptions(zap.IncreaseLevel(zapcore.ErrorLevel)),
			}
		}),

		config.Module,
		logger.Module,
		middleware.Module,
		grpc.Module,
		delivery.Module,
	).Run()
}
