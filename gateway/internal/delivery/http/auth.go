package http

import (
	"time"

	pb "qrcodegen/pkg/api/v1/qrcodegen/v1"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	client pb.QRCodeServiceClient
}

func NewAuthHandler(client pb.QRCodeServiceClient) *AuthHandler {
	return &AuthHandler{
		client: client,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req struct {
		Name           string `json:"name"`
		Email          string `json:"email"`
		Password       string `json:"password"`
		SecondPassword string `json:"second_password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Password != req.SecondPassword {
		return fiber.NewError(fiber.StatusBadRequest, "Passwords do not match")
	}

	_, err := h.client.Register(c.Context(), &pb.RegisterRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User registered successfully"})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	resp, err := h.client.Login(c.Context(), &pb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		return err
	}

	c.Cookie(&fiber.Cookie{
		Name:     "jwt_token",
		Value:    resp.Token,
		HTTPOnly: true,
		Secure:   false, // set true in prod
		SameSite: "Lax",
		Expires:  time.Now().Add(24 * time.Hour),
	})

	return c.SendStatus(fiber.StatusOK)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	c.ClearCookie("jwt_token")
	return c.SendStatus(fiber.StatusOK)
}
