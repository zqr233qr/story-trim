package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"github/zqr233qr/story-trim/internal/core/domain"
	"github/zqr233qr/story-trim/internal/core/port"
)

type userService struct {
	userRepo  port.UserRepository
	jwtSecret []byte
}

func NewUserService(ur port.UserRepository, secret string) *userService {
	return &userService{
		userRepo:  ur,
		jwtSecret: []byte(secret),
	}
}

func (s *userService) Register(ctx context.Context, username, password string) error {
	// 1. 检查用户是否存在
	existing, _ := s.userRepo.GetByUsername(ctx, username)
	if existing != nil {
		return errors.New("user already exists")
	}

	// 2. 密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return err
	}

	// 3. 存储
	user := &domain.User{
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}
	return s.userRepo.Create(ctx, user)
}

func (s *userService) Login(ctx context.Context, username, password string) (string, error) {
	// 1. 查找用户
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", errors.New("invalid username or password")
	}

	// 2. 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid username or password")
	}

	// 3. 生成 JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": user.ID,
		"exp":    time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString(s.jwtSecret)
}

func (s *userService) ValidateToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}

	userID := uint(claims["userID"].(float64))
	return userID, nil
}
