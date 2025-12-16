package delivery

import (
	"qrcodegen/gateway/config"
	"qrcodegen/gateway/internal/delivery/http"
	"qrcodegen/gateway/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

func NewRouter(
	cfg *config.Config,
	logger *zap.Logger,
	authMiddleware *middleware.Auth,
	authHandler *http.AuthHandler,
	linkHandler *http.LinkHandler,
	qrHandler *http.QRHandler,
	redirectHandler *http.RedirectHandler,
) *fiber.App {

	app := fiber.New(fiber.Config{
		ErrorHandler: http.NewGlobalErrorHandler(logger),
		ProxyHeader:  fiber.HeaderXForwardedFor,
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowCredentials: true,
	}))

	app.Get("/redirect/:hash", redirectHandler.Redirect)

	api := app.Group("/api/v1")
	api.Post("/login", authHandler.Login)
	api.Post("/register", authHandler.Register)
	api.Post("/logout", authHandler.Logout)

	protected := api.Group("/", authMiddleware.Protected())

	protected.Post("/links/create", linkHandler.CreateLink)
	protected.Get("/links", linkHandler.GetAllLinks)
	protected.Get("/links/:id", linkHandler.GetLink)
	protected.Patch("/links/:id", linkHandler.EditLink)
	protected.Delete("/links/:id", linkHandler.DeleteLink)
	protected.Get("/links/:id/transitions", linkHandler.GetTransitions)

	protected.Post("/qrcode", qrHandler.GenerateQR)
	protected.Get("/links/:id/download", qrHandler.DownloadQR)

	return app
}
