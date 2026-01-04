package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github/zqr233qr/story-trim/internal/domain"
	"github/zqr233qr/story-trim/pkg/config"
)

type UserService struct {
	db     *gorm.DB
	config config.AuthConfig
}

func NewUserService(db *gorm.DB, cfg config.AuthConfig) *UserService {
	return &UserService{
		db:     db,
		config: cfg,
	}
}

// Register 注册新用户
func (s *UserService) Register(username, password string) (*domain.User, error) {
	// 1. 检查是否存在
	var count int64
	s.db.Model(&domain.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return nil, errors.New("username already exists")
	}

	// 2. 加密密码
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 3. 创建用户
	user := &domain.User{
		Username: username,
		Password: string(hashedPwd),
		Role:     domain.RoleUser,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// Login 登录并返回 Token
func (s *UserService) Login(username, password string) (string, error) {
	var user domain.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("invalid credentials")
		}
		return "", err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// 生成 Token
	return s.GenerateToken(user.ID, string(user.Role))
}

// GenerateToken 生成 JWT
func (s *UserService) GenerateToken(userID uint, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"role": role,
		"exp":  time.Now().Add(time.Duration(s.config.TokenDuration) * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}
