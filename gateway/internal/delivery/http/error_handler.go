package http

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewGlobalErrorHandler(logger *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		resp := ErrorResponse{
			Success: false,
			Error: ErrorDetails{
				Code:    "INTERNAL_ERROR",
				Message: "Internal Server Error",
			},
		}

		var e *fiber.Error
		if errors.As(err, &e) {
			code = e.Code
			resp.Error.Code = "HTTP_ERROR"
			resp.Error.Message = e.Message
		}

		if st, ok := status.FromError(err); ok {
			code = grpcStatusToHTTP(st.Code())
			resp.Error.Code = st.Code().String()
			resp.Error.Message = st.Message()
		}

		if code >= 500 {
			logger.Error("Request failed",
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
				zap.Int("status", code),
				zap.Error(err),
			)
		}

		return c.Status(code).JSON(resp.Error)
	}
}

func grpcStatusToHTTP(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
