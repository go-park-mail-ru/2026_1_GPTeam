package jwt_auth

import (
	"github.com/go-park-mail-ru/2026_1_GPTeam/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
)

// ValidateAccessToken проверяет access JWT без обращения к БД (для API-gateway).
func ValidateAccessToken(tokenStr string, secret []byte, jwtVersion string) (bool, int) {
	log := logger.GetLogger()
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Error("unexpected signing method")
			return nil, WrongSigningMethodError
		}
		return secret, nil
	})
	if err != nil {
		return false, -1
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		log.Warn("invalid token (unable to claim payload)")
		return false, -1
	}

	version, ok := claims["version"].(string)
	if !ok {
		log.Warn("invalid token (unable to claim version)")
		return false, -1
	}
	if version != jwtVersion {
		log.Warn("invalid token (invalid version)")
		return false, -1
	}

	userIdFloat, ok := claims["user_id"].(float64)
	userId := int(userIdFloat)
	if !ok {
		log.Warn("invalid token (unable to claim user_id)")
		return false, -1
	}
	return true, userId
}
