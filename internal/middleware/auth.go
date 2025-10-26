package middleware

import (
	"log"
	"strings"
	"time"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/appleboy/gin-jwt/v3/core"
	"github.com/gin-gonic/gin"
	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/mycodeLife01/qa/config"
	"github.com/mycodeLife01/qa/internal/dto"
	"github.com/mycodeLife01/qa/internal/model"
	"github.com/mycodeLife01/qa/internal/pkg/api"
	"github.com/mycodeLife01/qa/internal/service"
)

// IdentityKey 定义了存储在 gin.Context 中的用户标识的键名
var identityKey = "username"

// NewAuthMiddleware 创建并配置 gin-jwt 中间件
func NewAuthMiddleware(authService service.AuthService) (*jwt.GinJWTMiddleware, error) {

	// === gin-jwt 配置 ===
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:         "Protected Zone",
		Key:           []byte(config.C.JWT.JWTSecretKey),
		Timeout:       time.Hour,
		MaxRefresh:    time.Hour * 24,
		IdentityKey:   identityKey,
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,

		PayloadFunc: func(data any) gojwt.MapClaims {
			if v, ok := data.(*model.User); ok {
				return gojwt.MapClaims{
					identityKey: v.Username,
				}
			}
			return gojwt.MapClaims{}
		},

		// IdentityHandler 定义了如何从 Payload 中提取用户标识
		IdentityHandler: func(c *gin.Context) any {
			claims := jwt.ExtractClaims(c)
			// 安全地获取 username，避免 panic
			username, ok := claims[identityKey].(string)
			if !ok {
				log.Printf("identityHandler: failed to extract username from claims")
				return nil
			}
			return &model.User{
				Username: username,
			}
		},

		// Authenticator 是核心函数，在用户尝试登录时调用
		// 参数 c *gin.Context 允许你从请求中获取用户名和密码
		// 返回值 interface{} 将被传递给 PayloadFunc；error 表示认证失败
		Authenticator: func(c *gin.Context) (any, error) {
			var loginVals dto.LoginRequest
			if err := c.ShouldBindJSON(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			username := loginVals.Username
			password := loginVals.Password

			loginUser, _ := authService.IsValidUser(username, password)
			if loginUser != nil {
				// 返回指针类型，与 payloadFunc 中的类型断言匹配
				return loginUser, nil
			}
			return nil, jwt.ErrFailedAuthentication
		},

		// Authorizator 验证从 Token 解析出的用户是否有权限访问请求的资源
		// payload 是 IdentityHandler 返回的值 (这里是 uint)
		// 返回 true 表示授权成功
		Authorizer: func(c *gin.Context, data any) bool {
			exist, err := authService.IsExistUser(data.(*model.User).Username)
			if err != nil {
				return false
			}
			if !exist {
				return false
			}
			role, err := authService.GetUserRoleByName(data.(*model.User).Username)
			if err != nil {
				return false
			}
			if role != "banned" && role != "" {
				requestPath := c.Request.URL.Path
				router := strings.Split(requestPath, "/")[1]
				if router == "user" {
					if role == "admin" {
						return true
					} else {
						return false
					}
				}
				return true
			}
			return false
		},

		// Unauthorized 定义了认证失败或授权失败时的响应
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code": code,
				"msg":  message,
				"data": nil,
			})
		},

		LogoutResponse: func(c *gin.Context) {
			claims := jwt.ExtractClaims(c)
			if len(claims) > 0 {
				username := claims[identityKey].(string)
				logoutInfo := map[string]any{"logoutUsername": username, "success": true}
				c.Set(api.ResponseDataKey, logoutInfo)
			} else {
				logoutInfo := map[string]any{"logoutUsername": "", "success": false}
				c.Set(api.ResponseDataKey, logoutInfo)
			}

		},
		LoginResponse: func(c *gin.Context, token *core.Token) {
			response := gin.H{
				"access_token": token.AccessToken,
				"token_type":   token.TokenType,
				"expires_in":   token.ExpiresIn(),
			}
			c.Set(api.ResponseDataKey, response)
		},
	})

	if err != nil {
		return nil, err
	}

	return authMiddleware, nil
}
