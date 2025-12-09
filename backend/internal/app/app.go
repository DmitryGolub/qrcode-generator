package app

import (
	"context"
	"net"

	"qrcodegen/config"
	"qrcodegen/internal/handler/grpc"
	"qrcodegen/pkg/database"
	"qrcodegen/pkg/geo"
	"qrcodegen/internal/repository/postgres"
	"qrcodegen/internal/usecase"
	pb "qrcodegen/pkg/api/v1/qrcodegen/v1"

	"github.com/rs/zerolog/log"
	"go.uber.org/fx"
	googlegrpc "google.golang.org/grpc"
)

func New() *fx.App {
	return fx.New(
		fx.Provide(
			config.New,
			database.NewDBPool,
			postgres.NewRepository,
			geo.NewGeoResolver,
			usecase.NewUserUseCase,
			usecase.NewLinkUseCase,
			usecase.NewQRUseCase,
			grpc.NewHandler,
			NewGRPCServer,
		),
		fx.Invoke(
			func(lifecycle fx.Lifecycle, srv *googlegrpc.Server, cfg *config.Config) {
				lifecycle.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						lis, err := net.Listen("tcp", cfg.GRPCServerAddress)
						if err != nil {
							return err
						}
						log.Info().Msgf("Starting gRPC server on %s", cfg.GRPCServerAddress)
						go srv.Serve(lis)
						return nil
					},
					OnStop: func(ctx context.Context) error {
						srv.GracefulStop()
						return nil
					},
				})
			},
		),
	)
}

func NewGRPCServer(handler *grpc.Handler) *googlegrpc.Server {
	s := googlegrpc.NewServer()
	pb.RegisterQRCodeServiceServer(s, handler)
	return s
}
