package utils

import (
	"errors"
	"fmt"
	"hive-admin-go/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// minJWTSecretLength JWT 签名密钥最小长度（字节）；不足时拒绝启动，避免弱密钥签发凭证
const minJWTSecretLength = 32

var jwtSecret []byte

// InitJWT 校验并加载 JWT 签名密钥；密钥过短返回错误，由启动流程终止服务。
func InitJWT() error {
	secret := config.AppConfig.JWT.Secret
	if len(secret) < minJWTSecretLength {
		return fmt.Errorf("jwt secret 长度不足 %d 字节，拒绝启动", minJWTSecretLength)
	}
	jwtSecret = []byte(secret)
	return nil
}

type Claims struct {
	UserID     string `json:"userId"`
	PwdVersion int    `json:"pwdVersion"`
	JTI        string `json:"jti"`
	jwt.RegisteredClaims
}

// GenerateToken 签发登录 token，有效期 expireHours 小时，携带签发时刻的密码版本号，
// 供认证中间件在口令变更后立即判定旧世代凭证失效；jti 是本枚凭证的唯一标识，
// 供登出撤销写入黑名单。
func GenerateToken(userID string, pwdVersion, expireHours int) (string, error) {
	expireTime := time.Now().Add(time.Duration(expireHours) * time.Hour)
	claims := Claims{
		UserID:     userID,
		PwdVersion: pwdVersion,
		JTI:        GenerateUUID(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// SignShortLivedToken 签发短时 JWT token，复用项目 JWT 密钥，供需要"短期凭证 + 无状态校验"的场景使用。
// claims 必须由调用方提供并自行设置 ExpiresAt、IssuedAt、Issuer 等字段；本函数不修改 claims 内容。
// 与 GenerateToken 不同，本函数不绑定登录态，调用方决定 payload 内容（taskId、邮箱、回调标识等）和有效期。
// 典型场景：文件预览/下载授权、邮件验证链接、回调 URL 签名、跨服务临时凭证。
// 安全边界：token 在有效期内可被持有者重放使用，调用方应结合业务校验（如对应记录是否仍存在）缓解风险；
// 需要真正一次性失效时由调用方自行引入数据库存储。
func SignShortLivedToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseShortLivedToken 解析并校验短时 token，调用方传入自己实现的 jwt.Claims 实例（指针）接收 payload。
// 校验流程：JWT 签名 + 过期时间 + 标准 claims 有效性；不校验黑名单，不强制 userId。
// 调用方拿到 claims 后必须按业务规则做进一步校验（如对应业务记录是否仍存在、用户是否仍拥有该资源），
// 避免 token 在有效期内被他人持有后越权访问。
func ParseShortLivedToken(tokenString string, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("invalid short-lived token")
	}
	return nil
}
