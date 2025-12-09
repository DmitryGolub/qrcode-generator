package http

import (
	"fmt"
	"net/http"
	"time"

	"qrcodegen/gateway/config"
	pb "qrcodegen/pkg/api/v1/qrcodegen/v1"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GatewayHandler struct {
	client  pb.QRCodeServiceClient
	baseURL string
}

func NewGatewayHandler(client pb.QRCodeServiceClient, cfg *config.Config) *GatewayHandler {
	return &GatewayHandler{
		client:  client,
		baseURL: cfg.AppBaseURL,
	}
}

// --- Auth ---

func (h *GatewayHandler) Register(c *fiber.Ctx) error {
	var req struct {
		Name           string `json:"name"`
		Email          string `json:"email"`
		Password       string `json:"password"`
		SecondPassword string `json:"second_password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Password != req.SecondPassword {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Passwords do not match"})
	}

	_, err := h.client.Register(c.Context(), &pb.RegisterRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		return h.mapError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User registered successfully"})
}

func (h *GatewayHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	resp, err := h.client.Login(c.Context(), &pb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		return h.mapError(c, err)
	}

	// Устанавливаем HTTP-Only Cookie
	c.Cookie(&fiber.Cookie{
		Name:     "jwt_token",
		Value:    resp.Token,
		HTTPOnly: true,
		Secure:   false, // Set to true in production (HTTPS)
		SameSite: "Lax",
		Expires:  time.Now().Add(24 * time.Hour), // Лучше вынести в конфиг
	})

	return c.SendStatus(fiber.StatusOK)
}

func (h *GatewayHandler) Logout(c *fiber.Ctx) error {
	c.ClearCookie("jwt_token")
	return c.SendStatus(fiber.StatusOK)
}

// --- Links ---

func (h *GatewayHandler) CreateLink(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)

	var req struct {
		OriginalURL string `json:"original_url"`
		Name        string `json:"name"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	resp, err := h.client.CreateLink(c.Context(), &pb.CreateLinkRequest{
		OriginalUrl: req.OriginalURL,
		Name:        req.Name,
		UserId:      userID,
	})

	if err != nil {
		return h.mapError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":      resp.Id,
		"message": "Link created successfully",
	})
}

func (h *GatewayHandler) GetAllLinks(c *fiber.Ctx) error {
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
		return h.mapError(c, err)
	}

	// Если ссылок нет, возвращаем пустой массив вместо null
	links := resp.Links
	if links == nil {
		links = []*pb.Link{}
	}

	return c.JSON(fiber.Map{
		"links":   links,
		"message": "Success",
	})
}

func (h *GatewayHandler) GetLink(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	linkID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid link ID"})
	}

	resp, err := h.client.GetLink(c.Context(), &pb.GetLinkRequest{
		Id:     int64(linkID),
		UserId: userID,
	})

	if err != nil {
		return h.mapError(c, err)
	}

	return c.JSON(resp)
}

func (h *GatewayHandler) EditLink(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	linkID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid link ID"})
	}

	var req struct {
		OriginalURL string  `json:"original_url"`
		Color       string  `json:"color"`
		Background  string  `json:"background"`
		Smoothing   float64 `json:"smoothing"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
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
		return h.mapError(c, err)
	}

	return c.JSON(fiber.Map{
		"id":      resp.Id,
		"message": "Link updated successfully",
	})
}

func (h *GatewayHandler) DeleteLink(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	linkID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid link ID"})
	}

	_, err = h.client.DeleteLink(c.Context(), &pb.DeleteLinkRequest{
		Id:     int64(linkID),
		UserId: userID,
	})

	if err != nil {
		return h.mapError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// --- Analytics ---

func (h *GatewayHandler) GetTransitions(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	linkID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid link ID"})
	}

	resp, err := h.client.GetTransitions(c.Context(), &pb.GetTransitionsRequest{
		LinkId: int64(linkID),
		UserId: userID,
	})

	if err != nil {
		return h.mapError(c, err)
	}

	transitions := resp.Transitions
	if transitions == nil {
		transitions = []*pb.Transition{}
	}

	return c.JSON(fiber.Map{"transitions": transitions})
}

// --- QR Code & Download ---

func (h *GatewayHandler) GenerateQR(c *fiber.Ctx) error {
	var req struct {
		URL        string  `json:"url"`
		Color      string  `json:"color"`
		Background string  `json:"background"`
		Smoothing  float64 `json:"smoothing"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	resp, err := h.client.GenerateQRCode(c.Context(), &pb.GenerateQRCodeRequest{
		Url:        req.URL,
		Color:      req.Color,
		Background: req.Background,
		Smoothing:  req.Smoothing,
	})

	if err != nil {
		return h.mapError(c, err)
	}

	c.Set("Content-Type", "image/png")
	return c.Send(resp.Image)
}

func (h *GatewayHandler) DownloadQR(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	linkID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid link ID"})
	}

	// 1. Получаем настройки ссылки
	linkResp, err := h.client.GetLink(c.Context(), &pb.GetLinkRequest{
		Id:     int64(linkID),
		UserId: userID,
	})
	if err != nil {
		return h.mapError(c, err)
	}

	// 2. Формируем URL для редиректа
	// h.baseURL должен быть, например, "http://localhost:8080"
	redirectURL := fmt.Sprintf("%s/redirect/%s", h.baseURL, linkResp.Hash)

	// 3. Генерируем QR код через gRPC
	qrResp, err := h.client.GenerateQRCode(c.Context(), &pb.GenerateQRCodeRequest{
		Url:        redirectURL,
		Color:      linkResp.Color,
		Background: linkResp.Background,
		Smoothing:  linkResp.Smoothing,
	})
	if err != nil {
		return h.mapError(c, err)
	}

	// 4. Отдаем файл
	filename := fmt.Sprintf("qr-%d.png", linkID)
	c.Set("Content-Type", "image/png")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(qrResp.Image)
}

// --- Redirect ---

func (h *GatewayHandler) Redirect(c *fiber.Ctx) error {
	hash := c.Params("hash")
	if hash == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Hash required")
	}

	resp, err := h.client.GetRedirect(c.Context(), &pb.GetRedirectRequest{
		Hash:      hash,
		Ip:        c.IP(),
		UserAgent: c.Get("User-Agent"),
		Referer:   c.Get("Referer"),
	})

	if err != nil {
		// Если ссылка не найдена, можно вернуть 404 или редирект на 404 страницу фронта
		return c.Status(fiber.StatusNotFound).SendString("Link not found")
	}

	return c.Redirect(resp.OriginalUrl, fiber.StatusFound)
}

// --- Helpers ---

func (h *GatewayHandler) mapError(c *fiber.Ctx, err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var httpStatus int
	switch st.Code() {
	case codes.OK:
		httpStatus = http.StatusOK
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
	case codes.PermissionDenied:
		httpStatus = http.StatusForbidden
	case codes.NotFound:
		httpStatus = http.StatusNotFound
	case codes.AlreadyExists:
		httpStatus = http.StatusConflict
	case codes.Internal:
		httpStatus = http.StatusInternalServerError
	case codes.Unavailable:
		httpStatus = http.StatusServiceUnavailable
	default:
		httpStatus = http.StatusInternalServerError
	}

	return c.Status(httpStatus).JSON(fiber.Map{
		"error": st.Message(),
	})
}
