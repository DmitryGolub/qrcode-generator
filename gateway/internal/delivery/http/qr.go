package http

import (
	"fmt"
	"qrcodegen/gateway/config"
	pb "qrcodegen/pkg/api/v1/qrcodegen/v1"

	"github.com/gofiber/fiber/v2"
)

type QRHandler struct {
	client  pb.QRCodeServiceClient
	baseURL string
}

func NewQRHandler(client pb.QRCodeServiceClient, cfg *config.Config) *QRHandler {
	return &QRHandler{
		client:  client,
		baseURL: cfg.AppBaseURL,
	}
}

func (h *QRHandler) GenerateQR(c *fiber.Ctx) error {
	var req struct {
		URL        string  `json:"url"`
		Color      string  `json:"color"`
		Background string  `json:"background"`
		Smoothing  float64 `json:"smoothing"`
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	resp, err := h.client.GenerateQRCode(c.Context(), &pb.GenerateQRCodeRequest{
		Url:        req.URL,
		Color:      req.Color,
		Background: req.Background,
		Smoothing:  req.Smoothing,
	})

	if err != nil {
		return err
	}

	c.Set("Content-Type", "image/png")
	return c.Send(resp.Image)
}

func (h *QRHandler) DownloadQR(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	linkID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid link ID")
	}

	linkResp, err := h.client.GetLink(c.Context(), &pb.GetLinkRequest{
		Id:     int64(linkID),
		UserId: userID,
	})
	if err != nil {
		return err
	}

	redirectURL := fmt.Sprintf("%s/redirect/%s", h.baseURL, linkResp.Hash)

	qrResp, err := h.client.GenerateQRCode(c.Context(), &pb.GenerateQRCodeRequest{
		Url:        redirectURL,
		Color:      linkResp.Color,
		Background: linkResp.Background,
		Smoothing:  linkResp.Smoothing,
	})
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("qr-%d.png", linkID)
	c.Set("Content-Type", "image/png")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(qrResp.Image)
}
