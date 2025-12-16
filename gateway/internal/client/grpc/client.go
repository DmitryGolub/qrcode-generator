package grpc

import (
	"context"

	"qrcodegen/gateway/config"
	pb "qrcodegen/pkg/api/v1/qrcodegen/v1"

	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var Module = fx.Module("grpc_client", fx.Provide(NewQRCodeClient))

func NewQRCodeClient(lc fx.Lifecycle, cfg *config.Config) (pb.QRCodeServiceClient, error) {
	target := cfg.GRPCTarget

	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return conn.Close()
		},
	})

	return pb.NewQRCodeServiceClient(conn), nil
}
