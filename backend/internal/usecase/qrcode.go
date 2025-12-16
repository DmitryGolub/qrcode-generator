package usecase

import (
	"context"

	"qrcodegen/internal/dto"
	"qrcodegen/pkg/qrcode"
)

type QRUseCase struct{}

func NewQRUseCase() *QRUseCase { return &QRUseCase{} }

func (uc *QRUseCase) Generate(ctx context.Context, req dto.GenerateQRCodeRequest) ([]byte, error) {
	return qrcode.GeneratePNG(req.URL, req.Color, req.Background, req.Smoothing)
}
