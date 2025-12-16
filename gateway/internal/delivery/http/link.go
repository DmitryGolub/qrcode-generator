package http

import (
	pb "qrcodegen/pkg/api/v1/qrcodegen/v1"
	"time"

	"github.com/gofiber/fiber/v2"
)

type LinkHandler struct {
	client pb.QRCodeServiceClient
}

func NewLinkHandler(client pb.QRCodeServiceClient) *LinkHandler {
	return &LinkHandler{client: client}
}

func (h *LinkHandler) CreateLink(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)

	var req struct {
		OriginalURL string `json:"original_url"`
		Name        string `json:"name"`
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	resp, err := h.client.CreateLink(c.Context(), &pb.CreateLinkRequest{
		OriginalUrl: req.OriginalURL,
		Name:        req.Name,
		UserId:      userID,
	})

	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":      resp.Id,
		"message": "Link created successfully",
	})
}

func (h *LinkHandler) GetAllLinks(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	search := c.Query("search")
	sortBy := c.Query("sort_by")
	order := c.Query("order")

	resp, err := h.client.GetAllLinks(c.Context(), &pb.ListLinksRequest{
		UserId: userID,
		Search: search,
		SortBy: sortBy,
		Order:  order,
	})

	if err != nil {
		return err
	}

	links := resp.Links
	if links == nil {
		links = []*pb.Link{}
	}

	return c.JSON(fiber.Map{
		"links":   links,
		"message": "Success",
	})
}

func (h *LinkHandler) GetLink(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	linkID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid link ID")
	}

	resp, err := h.client.GetLink(c.Context(), &pb.GetLinkRequest{
		Id:     int64(linkID),
		UserId: userID,
	})

	if err != nil {
		return err
	}

	return c.JSON(resp)
}

func (h *LinkHandler) EditLink(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	linkID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid link ID")
	}

	var req struct {
		OriginalURL string  `json:"original_url"`
		Color       string  `json:"color"`
		Background  string  `json:"background"`
		Smoothing   float64 `json:"smoothing"`
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	resp, err := h.client.EditLink(c.Context(), &pb.EditLinkRequest{
		Id:          int64(linkID),
		UserId:      userID,
		OriginalUrl: req.OriginalURL,
		Color:       req.Color,
		Background:  req.Background,
		Smoothing:   req.Smoothing,
	})

	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"id":      resp.Id,
		"message": "Link updated successfully",
	})
}

func (h *LinkHandler) DeleteLink(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	linkID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid link ID")
	}

	_, err = h.client.DeleteLink(c.Context(), &pb.DeleteLinkRequest{
		Id:     int64(linkID),
		UserId: userID,
	})

	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *LinkHandler) GetTransitions(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	linkID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid link ID")
	}

	resp, err := h.client.GetTransitions(c.Context(), &pb.GetTransitionsRequest{
		LinkId: int64(linkID),
		UserId: userID,
	})

	if err != nil {
		return err
	}

	type TransitionResponse struct {
		ID        int64     `json:"id"`
		Country   string    `json:"country,omitempty"`
		City      string    `json:"city,omitempty"`
		Referer   string    `json:"referer,omitempty"`
		UserAgent string    `json:"user_agent,omitempty"`
		Browser   string    `json:"browser,omitempty"`
		OS        string    `json:"os,omitempty"`
		CreatedAt time.Time `json:"created_at"`
	}

	transitions := make([]TransitionResponse, 0, len(resp.Transitions))

	for _, t := range resp.Transitions {
		transitions = append(transitions, TransitionResponse{
			ID:        t.Id,
			Country:   t.Country,
			City:      t.City,
			Referer:   t.Referer,
			UserAgent: t.UserAgent,
			Browser:   t.Browser,
			OS:        t.Os,
			CreatedAt: t.CreatedAt.AsTime(),
		})
	}

	return c.JSON(fiber.Map{"transitions": transitions})
}
