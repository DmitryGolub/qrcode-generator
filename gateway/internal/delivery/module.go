package delivery

import (
	"context"

	"qrcodegen/gateway/config"
	"qrcodegen/gateway/internal/delivery/http"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"delivery",
	fx.Provide(
		http.NewAuthHandler,
		http.NewLinkHandler,
		http.NewQRHandler,
		http.NewRedirectHandler,
		NewRouter,
	),
	fx.Invoke(registerHooks),
)

func registerHooks(lifecycle fx.Lifecycle, cfg *config.Config, app *fiber.App, logger *zap.Logger) {
	addr := cfg.HTTPPort

	lifecycle.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			logger.Info("Starting HTTP Gateway", zap.String("address", addr))

			go func() {
				if err := app.Listen(addr); err != nil {
					logger.Fatal("Fiber Gateway Listen failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			logger.Info("Stopping HTTP Gateway")
			return app.Shutdown()
		},
	})
}
