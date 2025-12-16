package middleware

import (
	"strconv"

	"qrcodegen/gateway/config"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type Auth struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewAuth(cfg *config.Config, logger *zap.Logger) *Auth {
	return &Auth{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *Auth) Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Cookies("jwt_token")
		if tokenString == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(a.cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			a.logger.Debug("Invalid token", zap.Error(err))
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid claims")
		}

		if sub, ok := claims["sub"].(string); ok {
			userID, _ := strconv.ParseInt(sub, 10, 64)
			c.Locals("userID", userID)
		} else {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid sub")
		}

		return c.Next()
	}
}
