package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/riouske/gophermart/internal/model"
	"github.com/riouske/gophermart/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type AuthService struct {
	userRepo    *repository.UserRepository
	jwtSecret   []byte
	tokenExpiry time.Duration
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		jwtSecret:   []byte(jwtSecret),
		tokenExpiry: 24 * time.Hour, // Token is valid for 24 hours
	}
}

func (s *AuthService) Register(credentials *model.UserCredentials) (*model.User, string, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(credentials.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &model.User{
		Login:    credentials.Login,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", err
	}

	token, err := s.GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) Login(credentials *model.UserCredentials) (*model.User, string, error) {
	user, err := s.userRepo.GetByLogin(credentials.Login)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) GenerateToken(userID int64) (string, error) {
	expiresAt := time.Now().Add(s.tokenExpiry)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s *AuthService) GetUserByID(userID int64) (*model.User, error) {
	return s.userRepo.GetByID(userID)
}
