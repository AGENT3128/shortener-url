package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/AGENT3128/shortener-url/internal/app/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	tokenExpires = 1 * time.Hour
	secretKey    = "NuQu82Q2"
	authCookie   = "Auth"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// AuthMiddleware
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(authCookie)
		if errors.Is(err, http.ErrNoCookie) {
			logger.Log.Info("No cookie found, generating new token")
			// generate token and set cookie
			userID := uuid.New().String()
			tokenString, err := generateToken(userID)
			if err != nil {
				logger.Log.Error("Failed to generate token", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				c.Abort()
				return
			}

			c.SetCookie(authCookie, tokenString, int(tokenExpires.Seconds()), "/", "", false, true)
			c.Set("userID", userID)
			c.Next()
			return
		} else if err != nil {
			logger.Log.Error("Failed to get token", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		userID, err := getUserID(token)
		if err != nil {
			logger.Log.Error("Failed to get userID", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		// set to request context
		c.Set("userID", userID)
		c.Next()
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
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
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
