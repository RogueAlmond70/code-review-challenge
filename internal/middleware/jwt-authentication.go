package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	conf "github.com/RogueAlmond70/code-review-challenge/internal/config"
	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
)

var (
	jwtSecret string
	cfg       conf.Config
	logger    *zap.Logger
)

var ErrInvalidFormat = errors.New("invalid authorization header format")

func init() {
	jwtSecret = cfg.JWTToken
	if jwtSecret == "" {
		log.Fatal("Environment variable MY_JWT_TOKEN not set")
	}
}

// GenerateJWT generates a JWT with standard claims.
func GenerateJWT(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"authorized": true,
		"user":       username,
		"exp":        time.Now().Add(30 * time.Minute).Unix(),
	})

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		logger.Error("Error signing JWT", zap.Error(err))
		return "", fmt.Errorf("error signing JWT: %w", err)
	}
	return tokenString, nil
}

// IsAuthorized is middleware that verifies JWT in the Authorization header.
func IsAuthorized(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			logger.Error("Authorization header missing")
			return
		}

		tokenString, err := extractBearerToken(authHeader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			logger.Error("Unauthorized")
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				logger.Error("unexpected signing method")
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			logger.Error("invalid token")
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func extractBearerToken(authHeader string) (string, error) {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		logger.Error("unable to extract bearer token", zap.Error(ErrInvalidFormat))
		return "", fmt.Errorf("unable to extract bearer token: %w", ErrInvalidFormat)
	}
	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

// HomePage is a protected route for testing authorization.
func HomePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Super secret info")
}
