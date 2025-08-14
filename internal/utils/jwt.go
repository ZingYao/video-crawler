package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims 自定义 JWT 声明结构
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsAdmin  *bool  `json:"is_admin,omitempty"`
	jwt.RegisteredClaims
}

// JWTManager JWT 管理器
type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
}

// NewJWTManager 创建新的 JWT 管理器
func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
	}
}

// GenerateToken 生成 JWT token
func (j *JWTManager) GenerateToken(userID, username string, isAdmin bool) (string, error) {
	isAdminPtr := &isAdmin
	if !isAdmin {
		isAdminPtr = nil
	}
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		IsAdmin:  isAdminPtr,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "video-crawler",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ParseToken 解析 JWT token
func (j *JWTManager) ParseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateToken 验证 token 是否有效
func (j *JWTManager) ValidateToken(tokenString string) bool {
	_, err := j.ParseToken(tokenString)
	return err == nil
}

// RefreshToken 刷新 token
func (j *JWTManager) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查 token 是否即将过期（比如还有1小时过期）
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", errors.New("token is not near expiration")
	}

	// 生成新的 token
	return j.GenerateToken(claims.UserID, claims.Username, *claims.IsAdmin)
}

// GetTokenInfo 获取 token 中的信息
func (j *JWTManager) GetTokenInfo(tokenString string) (map[string]interface{}, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"user_id":  claims.UserID,
		"username": claims.Username,
		"is_admin": claims.IsAdmin,
		"exp":      claims.ExpiresAt.Time,
		"iat":      claims.IssuedAt.Time,
		"iss":      claims.Issuer,
		"sub":      claims.Subject,
	}, nil
}

// ExtractTokenFromHeader 从 Authorization header 中提取 token
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is empty")
	}

	// 检查 Bearer token 格式
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", errors.New("invalid authorization header format")
	}

	return authHeader[7:], nil
}
