package main

import (
	"context"
	"qrcodegen/gateway/config"
	"qrcodegen/gateway/internal/client/grpc"
	"qrcodegen/gateway/internal/delivery/http"
	"qrcodegen/gateway/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			// 1. Предоставляем конфиг
			config.New,
			// 2. Клиенты и хендлеры теперь автоматически получат конфиг
			grpc.NewQRCodeClient,
			http.NewGatewayHandler,
			NewFiberApp,
		),
		fx.Invoke(func(lc fx.Lifecycle, app *fiber.App, cfg *config.Config) {
			// Запуск сервера через Lifecycle
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go app.Listen(cfg.HTTPPort)
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return app.Shutdown()
				},
			})
		}),
	).Run()
}

func NewFiberApp(handler *http.GatewayHandler, cfg *config.Config) *fiber.App {
	app := fiber.New()
	app.Use(cors.New())

	app.Get("/redirect/:hash", handler.Redirect)
	api := app.Group("/api/v1")
	api.Post("/login", handler.Login)
	api.Post("/register", handler.Register)

	auth := api.Group("/", middleware.Auth(cfg.JWTSecret)) 
	
	auth.Post("/links/create", handler.CreateLink)
	auth.Get("/links", handler.GetAllLinks)
	auth.Get("/links/:id", handler.GetLink)
	auth.Patch("/links/:id", handler.EditLink)
	auth.Delete("/links/:id", handler.DeleteLink)
	
	auth.Post("/qrcode", handler.GenerateQR)
	auth.Get("/links/:id/download", handler.DownloadQR)
	
	auth.Get("/links/:id/transitions", handler.GetTransitions)

	return app
}