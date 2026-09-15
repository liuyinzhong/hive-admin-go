package utils

import (
	"errors"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const (
	// passwordMinLength 密码最小长度（字符数）
	passwordMinLength = 8
	// passwordMaxLength 密码最大长度（字节数）。bcrypt 只消费前 72 字节，
	// 超长部分会被静默截断导致不同密码命中同一哈希，必须在此拦截。
	passwordMaxLength = 72
)

// HashPassword 生成密码的 bcrypt 哈希。哈希串自含盐与成本因子，不需要独立的盐字段。
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword 校验明文密码与 bcrypt 哈希是否匹配。
func VerifyPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// ValidatePassword 校验密码强度：长度不小于 8 个字符且不超过 72 字节，
// 并至少包含小写字母、大写字母、数字、特殊字符四类中的两类。
func ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < passwordMinLength {
		return errors.New("密码长度不能少于 8 位")
	}
	if len(password) > passwordMaxLength {
		return errors.New("密码长度不能超过 72 字节")
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= '0' && r <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	classes := 0
	for _, present := range []bool{hasLower, hasUpper, hasDigit, hasSpecial} {
		if present {
			classes++
		}
	}
	if classes < 2 {
		return errors.New("密码须至少包含字母、数字、特殊字符中的两类")
	}
	return nil
}
