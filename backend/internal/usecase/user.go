package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"qrcodegen/config"
	"qrcodegen/internal/dto"
	"qrcodegen/pkg/jwt"
	"qrcodegen/internal/repository/postgres"
	sqldb "qrcodegen/sqlc/generated"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserUseCase struct {
	repo postgres.Repository
	cfg  *config.Config
}

func NewUserUseCase(repo postgres.Repository, cfg *config.Config) *UserUseCase {
	return &UserUseCase{repo: repo, cfg: cfg}
}

func (uc *UserUseCase) Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	_, err := uc.repo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	params := sqldb.CreateUserParams{
		Name:           req.Name,
		Email:          req.Email,
		HashedPassword: string(hashedPassword),
	}
	_, err = uc.repo.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &dto.RegisterResponse{Message: "User registered successfully"}, nil
}

func (uc *UserUseCase) Login(ctx context.Context, req dto.LoginRequest) (string, time.Time, error) {
	user, err := uc.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", time.Time{}, ErrInvalidCredentials
		}
		return "", time.Time{}, fmt.Errorf("failed to get user by email: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password))
	if err != nil {
		return "", time.Time{}, ErrInvalidCredentials
	}

	token, expirationTime, err := jwt.Sign(user, uc.cfg)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return token, expirationTime, nil
}
