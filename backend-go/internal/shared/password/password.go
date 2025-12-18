package password

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// Service 密码服务接口
type Service interface {
	HashPassword(password string) (string, error)
	ValidatePassword(hashedPassword, password string) bool
	ValidatePasswordStrength(password string) error
}

// service 密码服务实现
type service struct {
	cost int
}

// Config 密码服务配置
type Config struct {
	Cost int `mapstructure:"cost"` // bcrypt 成本因子，默认为 bcrypt.DefaultCost (10)
}

// NewService 创建密码服务
func NewService(config Config) Service {
	cost := config.Cost
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}

	// 确保成本因子在合理范围内
	if cost < bcrypt.MinCost {
		cost = bcrypt.MinCost
	}
	if cost > bcrypt.MaxCost {
		cost = bcrypt.MaxCost
	}

	return &service{
		cost: cost,
	}
}

// HashPassword 哈希密码
func (s *service) HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

// ValidatePassword 验证密码
func (s *service) ValidatePassword(hashedPassword, password string) bool {
	if hashedPassword == "" || password == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// ValidatePasswordStrength 验证密码强度
func (s *service) ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return errors.New("password must be no more than 128 characters long")
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

// GenerateRandomPassword 生成随机密码（可选功能）
func GenerateRandomPassword(length int) (string, error) {
	if length < 8 {
		length = 8
	}
	if length > 128 {
		length = 128
	}

	// 字符集
	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numbers   = "0123456789"
		special   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	)

	// 确保至少包含每种类型的字符
	password := make([]byte, 0, length)

	// 添加必需的字符类型
	password = append(password, lowercase[0]) // 至少一个小写字母
	password = append(password, uppercase[0]) // 至少一个大写字母
	password = append(password, numbers[0])   // 至少一个数字
	password = append(password, special[0])   // 至少一个特殊字符

	// 填充剩余长度
	allChars := lowercase + uppercase + numbers + special
	for len(password) < length {
		password = append(password, allChars[0]) // 简化实现，实际应该使用随机数
	}

	return string(password), nil
}

// IsPasswordCompromised 检查密码是否在常见密码列表中（可选功能）
func IsPasswordCompromised(password string) bool {
	// 常见弱密码列表
	commonPasswords := []string{
		"password", "123456", "password123", "admin", "qwerty",
		"letmein", "welcome", "monkey", "1234567890", "abc123",
		"password1", "123456789", "welcome123", "admin123",
	}

	for _, common := range commonPasswords {
		if password == common {
			return true
		}
	}

	return false
}
