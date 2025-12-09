package grpc

import (
	"context"
	"errors"

	"qrcodegen/internal/dto"
	"qrcodegen/internal/usecase"
	pb "qrcodegen/pkg/api/v1/qrcodegen/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	pb.UnimplementedQRCodeServiceServer
	userUC *usecase.UserUseCase
	linkUC *usecase.LinkUseCase
	qrUC   *usecase.QRUseCase
}

func NewHandler(userUC *usecase.UserUseCase, linkUC *usecase.LinkUseCase, qrUC *usecase.QRUseCase) *Handler {
	return &Handler{
		userUC: userUC,
		linkUC: linkUC,
		qrUC:   qrUC,
	}
}


func (h *Handler) Register(ctx context.Context, req *pb.RegisterRequest) (*emptypb.Empty, error) {
	_, err := h.userUC.Register(ctx, dto.RegisterRequest{
		Name:           req.Name,
		Email:          req.Email,
		Password:       req.Password,
		SecondPassword: req.Password,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (h *Handler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	token, _, err := h.userUC.Login(ctx, dto.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "Invalid credentials")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.LoginResponse{Token: token}, nil
}

func (h *Handler) CreateLink(ctx context.Context, req *pb.CreateLinkRequest) (*pb.Link, error) {
	resp, err := h.linkUC.CreateLink(ctx, dto.CreateLinkRequest{
		OriginalURL: req.OriginalUrl,
		Name:        req.Name,
	}, req.UserId)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Link{
		Id:          resp.ID,
		OriginalUrl: req.OriginalUrl,
		Name:        req.Name,
	}, nil
}

func (h *Handler) GetAllLinks(ctx context.Context, req *pb.ListLinksRequest) (*pb.ListLinksResponse, error) {
	var resp *dto.GetAllLinksResponse
	var err error

	if req.Search != "" {
		resp, err = h.linkUC.SearchLinksByName(ctx, req.UserId, req.Search)
	} else {
		resp, err = h.linkUC.GetAllLinks(ctx, req.UserId)
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	by := usecase.SortByCreatedAt
	if req.SortBy == "transitions" {
		by = usecase.SortByTransitions
	}
	ord := usecase.SortDesc
	if req.Order == "asc" {
		ord = usecase.SortAsc
	}

	sortedLinks := h.linkUC.SortLinks(resp.Links, by, ord)

	pbLinks := make([]*pb.Link, len(sortedLinks))
	for i, l := range sortedLinks {
		pbLinks[i] = &pb.Link{
			Id:               l.ID,
			OriginalUrl:      l.OriginalURL,
			Name:             l.Name,
			CreatedAt:        timestamppb.New(l.CreatedAt),
			TransitionsCount: l.Transitions,
		}
	}

	return &pb.ListLinksResponse{Links: pbLinks}, nil
}

func (h *Handler) GetLink(ctx context.Context, req *pb.GetLinkRequest) (*pb.Link, error) {
	link, err := h.linkUC.GetLinkByID(ctx, req.Id, req.UserId)
	if err != nil {
		if errors.Is(err, usecase.ErrLinkNotFound) {
			return nil, status.Error(codes.NotFound, "Link not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	var smoothing float64
	if link.Smoothing != nil {
		smoothing = *link.Smoothing
	}

	return &pb.Link{
		Id:          link.ID,
		OriginalUrl: link.OriginalURL,
		Hash:        link.Hash,
		Name:        link.Name,
		CreatedAt:   timestamppb.New(link.CreatedAt),
		Color:       link.Color,
		Background:  link.Background,
		Smoothing:   smoothing,
	}, nil
}

func (h *Handler) EditLink(ctx context.Context, req *pb.EditLinkRequest) (*pb.Link, error) {
	resp, err := h.linkUC.EditLink(ctx, req.Id, req.UserId, dto.EditLinkRequest{
		OriginalURL: req.OriginalUrl,
		Color:       req.Color,
		Background:  req.Background,
		Smoothing:   req.Smoothing,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrLinkNotFound) {
			return nil, status.Error(codes.NotFound, "Link not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Link{
		Id: resp.ID,
	}, nil
}

func (h *Handler) DeleteLink(ctx context.Context, req *pb.DeleteLinkRequest) (*emptypb.Empty, error) {
	err := h.linkUC.DeleteLink(ctx, req.Id, req.UserId)
	if err != nil {
		if errors.Is(err, usecase.ErrLinkNotFound) {
			return nil, status.Error(codes.NotFound, "Link not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (h *Handler) GetTransitions(ctx context.Context, req *pb.GetTransitionsRequest) (*pb.GetTransitionsResponse, error) {
	resp, err := h.linkUC.GetTransitions(ctx, req.LinkId, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbTransitions := make([]*pb.Transition, len(resp.Transitions))
	for i, t := range resp.Transitions {
		pbTransitions[i] = &pb.Transition{
			Id:        t.ID,
			Country:   derefString(t.Country),
			City:      derefString(t.City),
			Referer:   derefString(t.Referer),
			UserAgent: derefString(t.UserAgent),
			Browser:   derefString(t.Browser),
			Os:        derefString(t.OS),
			CreatedAt: timestamppb.New(t.CreatedAt),
		}
	}

	return &pb.GetTransitionsResponse{Transitions: pbTransitions}, nil
}

func (h *Handler) GenerateQRCode(ctx context.Context, req *pb.GenerateQRCodeRequest) (*pb.GenerateQRCodeResponse, error) {
	pngBytes, err := h.qrUC.Generate(ctx, dto.GenerateQRCodeRequest{
		URL:        req.Url,
		Color:      req.Color,
		Background: req.Background,
		Smoothing:  req.Smoothing,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to generate QR code")
	}

	return &pb.GenerateQRCodeResponse{Image: pngBytes}, nil
}

func (h *Handler) GetRedirect(ctx context.Context, req *pb.GetRedirectRequest) (*pb.GetRedirectResponse, error) {
	url, err := h.linkUC.Redirect(ctx, req.Hash, req.Referer, req.UserAgent, req.Ip)
	if err != nil {
		if errors.Is(err, usecase.ErrLinkNotFound) {
			return nil, status.Error(codes.NotFound, "Link not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetRedirectResponse{OriginalUrl: url}, nil
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
