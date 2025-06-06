package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	tokenExpires = 1 * time.Hour
	secretKey    = "NuQu82Q2" //nolint:gosec // secret key is used for authentication (test project)
	authCookie   = "Auth"
)

type contextKey string

const UserIDKey contextKey = "userID"

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type optionsAuthMiddleware struct {
	logger *zap.Logger
}

type OptionAuthMiddleware func(options *optionsAuthMiddleware) error

type AuthMiddleware struct {
	logger *zap.Logger
}

func WithAuthMiddlewareLogger(logger *zap.Logger) OptionAuthMiddleware {
	return func(options *optionsAuthMiddleware) error {
		options.logger = logger.With(zap.String("middleware", "auth"))
		return nil
	}
}

func NewAuthMiddleware(opts ...OptionAuthMiddleware) (*AuthMiddleware, error) {
	options := &optionsAuthMiddleware{}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}
	if options.logger == nil {
		return nil, errors.New("logger is required")
	}
	return &AuthMiddleware{logger: options.logger}, nil
}

// Handler returns chi middleware for authentication.
func (m *AuthMiddleware) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(authCookie)
			if errors.Is(err, http.ErrNoCookie) {
				m.logger.Info("No cookie found, generating new token")
				// generate token and set cookie
				userID := uuid.New().String()
				tokenString, errGenerate := generateToken(userID)
				if errGenerate != nil {
					m.logger.Error("Failed to generate token", zap.Error(errGenerate))
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				http.SetCookie(w, &http.Cookie{
					Name:     authCookie,
					Value:    tokenString,
					MaxAge:   int(tokenExpires.Seconds()),
					Path:     "/",
					HttpOnly: true,
				})

				ctx := r.Context()
				ctx = context.WithValue(ctx, UserIDKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			} else if err != nil {
				m.logger.Error("Failed to get token", zap.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			userID, err := getUserID(cookie.Value)
			if err != nil {
				m.logger.Error("Failed to get userID", zap.Error(err))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// set to request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func generateToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpires)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func getUserID(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(_ *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	return claims.UserID, nil
}
