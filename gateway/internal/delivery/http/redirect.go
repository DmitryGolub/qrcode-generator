package http

import (
	"strings"

	pb "qrcodegen/pkg/api/v1/qrcodegen/v1"

	"github.com/gofiber/fiber/v2"
)

type RedirectHandler struct {
	client pb.QRCodeServiceClient
}

func NewRedirectHandler(client pb.QRCodeServiceClient) *RedirectHandler {
	return &RedirectHandler{client: client}
}

func (h *RedirectHandler) Redirect(c *fiber.Ctx) error {
	hash := c.Params("hash")
	if hash == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Hash required")
	}

	clientIP := c.Get("X-Real-Ip")
	if clientIP == "" {
		clientIP = c.Get("X-Forwarded-For")
	}
	if clientIP == "" {
		clientIP = c.IP()
	}
	if strings.Contains(clientIP, ",") {
		clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
	}

	resp, err := h.client.GetRedirect(c.Context(), &pb.GetRedirectRequest{
		Hash:      hash,
		Ip:        clientIP,
		UserAgent: c.Get("User-Agent"),
		Referer:   c.Get("Referer"),
	})

	if err != nil {
		return err
	}

	return c.Redirect(resp.OriginalUrl, fiber.StatusFound)
}
