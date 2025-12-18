package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService JWT 服务接口
type JWTService interface {
	GenerateToken(userID uint, username string, role string) (*TokenPair, error)
	ValidateToken(tokenString string) (*Claims, error)
	RefreshToken(refreshToken string) (*TokenPair, error)
	RefreshTokenWithUserInfo(refreshToken string, username string, role string) (*TokenPair, error)
	GenerateAccessToken(userID uint, username string, role string) (string, error)
	GenerateRefreshToken(userID uint) (string, error)
}

// Claims JWT 声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Type     string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// TokenPair 令牌对
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // 访问令牌过期时间（秒）
	RefreshIn    int64  `json:"refresh_in"` // 刷新令牌过期时间（秒）
}

// jwtService JWT 服务实现
type jwtService struct {
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	issuer          string
}

// Config JWT 配置
type Config struct {
	SecretKey       string        `mapstructure:"secret_key"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
	Issuer          string        `mapstructure:"issuer"`
}

// NewJWTService 创建 JWT 服务
func NewJWTService(config Config) JWTService {
	return &jwtService{
		secretKey:       []byte(config.SecretKey),
		accessTokenTTL:  config.AccessTokenTTL,
		refreshTokenTTL: config.RefreshTokenTTL,
		issuer:          config.Issuer,
	}
}

// GenerateToken 生成令牌对
func (j *jwtService) GenerateToken(userID uint, username string, role string) (*TokenPair, error) {
	accessToken, err := j.GenerateAccessToken(userID, username, role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := j.GenerateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(j.accessTokenTTL.Seconds()),
		RefreshIn:    int64(j.refreshTokenTTL.Seconds()),
	}, nil
}

// GenerateAccessToken 生成访问令牌
func (j *jwtService) GenerateAccessToken(userID uint, username string, role string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   username,
			Audience:  []string{"prediction-system"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// GenerateRefreshToken 生成刷新令牌
func (j *jwtService) GenerateRefreshToken(userID uint) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Audience:  []string{"prediction-system"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTokenTTL)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken 验证令牌
func (j *jwtService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// RefreshToken 刷新令牌
func (j *jwtService) RefreshToken(refreshToken string) (*TokenPair, error) {
	claims, err := j.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 验证是否为刷新令牌
	if claims.Type != "refresh" {
		return nil, errors.New("invalid refresh token")
	}

	// 注意：刷新令牌只包含 UserID，需要从数据库或其他地方获取用户信息
	// 这里我们返回一个错误，表示需要额外的用户信息
	return nil, errors.New("refresh token requires user lookup - use RefreshTokenWithUserInfo instead")
}

// RefreshTokenWithUserInfo 使用用户信息刷新令牌
func (j *jwtService) RefreshTokenWithUserInfo(refreshToken string, username string, role string) (*TokenPair, error) {
	claims, err := j.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 验证是否为刷新令牌
	if claims.Type != "refresh" {
		return nil, errors.New("invalid refresh token")
	}

	// 生成新的令牌对
	return j.GenerateToken(claims.UserID, username, role)
}

// Validate 实现 Claims 验证接口
func (c Claims) Validate() error {
	// 验证令牌类型
	if c.Type != "access" && c.Type != "refresh" {
		return errors.New("invalid token type")
	}

	// 验证用户ID
	if c.UserID == 0 {
		return errors.New("invalid user ID")
	}

	// 对于访问令牌，验证用户名和角色
	if c.Type == "access" {
		if c.Username == "" {
			return errors.New("username is required for access token")
		}
		if c.Role == "" {
			return errors.New("role is required for access token")
		}
	}

	return nil
}
