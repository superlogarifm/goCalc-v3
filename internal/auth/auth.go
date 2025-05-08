package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type AuthService struct {
	secretKey     []byte
	tokenDuration time.Duration
}

func NewAuthService(secretKey string, tokenDuration time.Duration) (*AuthService, error) {
	if secretKey == "" {
		return nil, errors.New("secret key cannot be empty")
	}
	if tokenDuration <= 0 {
		return nil, errors.New("token duration must be positive")
	}
	return &AuthService{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
	}, nil
}

func (s *AuthService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func (s *AuthService) CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (s *AuthService) GenerateToken(userID uint, login string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,                                 // Subject (ID пользователя)
		"login": login,                                  // Логин пользователя
		"exp":   time.Now().Add(s.tokenDuration).Unix(), // Время истечения токена
		"iat":   time.Now().Unix(),                      // Время создания токена
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

// ValidateToken проверяет JWT токен и возвращает ID пользователя и логин, если токен валиден.
func (s *AuthService) ValidateToken(tokenString string) (userID uint, login string, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, "", ErrTokenExpired
		}
		return 0, "", ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		sub, subOk := claims["sub"].(float64)
		login, loginOk := claims["login"].(string)
		exp, expOk := claims["exp"].(float64)

		if !subOk || !loginOk || !expOk {
			return 0, "", ErrInvalidToken
		}

		if int64(exp) < time.Now().Unix() {
			return 0, "", ErrTokenExpired
		}

		return uint(sub), login, nil
	}

	return 0, "", ErrInvalidToken
}
