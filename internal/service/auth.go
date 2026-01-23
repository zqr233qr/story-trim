package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zqr233qr/story-trim/internal/errno"
	"github.com/zqr233qr/story-trim/internal/model"
	"github.com/zqr233qr/story-trim/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo          repository.AuthRepositoryInterface
	pointsService PointsServiceInterface
	jwtSecret     []byte
}

// NewAuthService 创建认证服务。
func NewAuthService(repo repository.AuthRepositoryInterface, pointsService PointsServiceInterface, secret string) *AuthService {
	return &AuthService{
		repo:          repo,
		pointsService: pointsService,
		jwtSecret:     []byte(secret),
	}
}

// Register 注册新用户并赠送积分。
func (s *AuthService) Register(ctx context.Context, username, password string) error {
	existing, _ := s.repo.GetByUsername(ctx, username)
	if existing != nil {
		return errno.ErrBookExist
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.User{
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return err
	}

	if err := s.pointsService.GrantRegisterBonus(ctx, user.ID, 100); err != nil {
		_ = s.repo.DeleteByID(ctx, user.ID)
		return err
	}
	return nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return "", errno.ErrAuthNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errno.ErrAuthWrongPwd
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": user.ID,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) ValidateToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return 0, errno.ErrAuthToken
	}

	if !token.Valid {
		return 0, errno.ErrAuthToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errno.ErrAuthToken
	}

	userID := uint(claims["userID"].(float64))
	return userID, nil
}

type AuthServiceInterface interface {
	Register(ctx context.Context, username, password string) error
	Login(ctx context.Context, username, password string) (string, error)
	ValidateToken(tokenString string) (uint, error)
}
