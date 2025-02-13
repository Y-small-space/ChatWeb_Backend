package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 结构体包含了用户身份信息和注册的标准 JWT 声明
type Claims struct {
	// UserID 用户的唯一标识符
	UserID string `json:"user_id"`
	// 使用 jwt.RegisteredClaims 来包含标准的 JWT 声明（如过期时间、签发时间等）
	jwt.RegisteredClaims
}

// GenerateToken 生成一个新的 JWT token
// 输入:
//   - userID: 用户的唯一标识符
//   - secret: 用于签名的密钥
//   - expireHours: token 的过期时间，单位为小时
//
// 输出:
//   - token 字符串: 生成的 JWT token
//   - error: 错误信息（如果有的话）
func GenerateToken(userID string, secret string, expireHours int) (string, error) {
	// 创建自定义的 Claims，其中包含用户ID和注册的 JWT 声明（如过期时间、签发时间）
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			// 设置 token 的过期时间为当前时间 + expireHours 小时
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expireHours))),
			// 设置 token 的签发时间为当前时间
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	// 使用 HS256 算法生成带有 Claims 的 JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 将 token 使用提供的 secret 进行签名并返回签名后的 token 字符串
	return token.SignedString([]byte(secret))
}

// ParseToken 解析 JWT token 并验证其有效性
// 输入:
//   - tokenString: 要解析的 token 字符串
//   - secret: 用于验证签名的密钥
//
// 输出:
//   - claims: 解析得到的 Claims 信息
//   - error: 错误信息（如果有的话）
func ParseToken(tokenString string, secret string) (*Claims, error) {
	// 使用提供的 secret 解析 token，并将其映射为 Claims 类型
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 返回密钥，用于验证签名
		return []byte(secret), nil
	})

	// 如果解析时出现错误，返回错误
	if err != nil {
		return nil, err
	}

	// 验证 token 是否有效，并返回解析出的 Claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	// 如果 token 无效，则返回签名无效错误
	return nil, jwt.ErrSignatureInvalid
}
