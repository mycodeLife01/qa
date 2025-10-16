package service

import (
	"log"
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/mycodeLife01/qa/models"
	"github.com/mycodeLife01/qa/types"
	"gorm.io/gorm"
)

type AuthService struct {
	userService *UserService
	DB          *gorm.DB
}

var identityKey = "username"

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{userService: NewUserService(db), DB: db}
}

func (as *AuthService) InitJwtMiddleware() *jwt.GinJWTMiddleware {
	authMiddleware, err := jwt.New(as.initParams())
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}
	errInit := authMiddleware.MiddlewareInit()
	if errInit != nil {
		log.Fatal("authMiddleware.MiddlewareInit() Error:" + errInit.Error())
	}
	return authMiddleware
}

func (as *AuthService) initParams() *jwt.GinJWTMiddleware {

	return &jwt.GinJWTMiddleware{
		Realm:           "test zone",
		Key:             []byte("secret key"),
		Timeout:         time.Hour,
		MaxRefresh:      time.Hour,
		IdentityKey:     identityKey,
		PayloadFunc:     as.payloadFunc(),
		IdentityHandler: as.identityHandler(),
		Authenticator:   as.authenticator(),
		Authorizer:      as.authorizator(),
		Unauthorized:    as.unauthorized(),
		LogoutResponse:  as.logoutResponse(),
		TokenLookup:     "header: Authorization, query: token, cookie: jwt",
		TokenHeadName:   "Bearer",
		TimeFunc:        time.Now,
	}
}

func (as *AuthService) payloadFunc() func(data any) gojwt.MapClaims {
	return func(data any) gojwt.MapClaims {
		log.Printf("payloadFunc data: %v", data)
		if v, ok := data.(*types.LoginResponse); ok {
			return gojwt.MapClaims{
				identityKey: v.Username,
			}
		}
		return gojwt.MapClaims{}
	}
}

func (as *AuthService) identityHandler() func(c *gin.Context) any {
	return func(c *gin.Context) any {
		claims := jwt.ExtractClaims(c)
		// 安全地获取 username，避免 panic
		username, ok := claims[identityKey].(string)
		if !ok {
			log.Printf("identityHandler: failed to extract username from claims")
			return nil
		}
		return &types.LoginResponse{
			Username: username,
		}
	}
}

func (as *AuthService) authenticator() func(c *gin.Context) (any, error) {
	return func(c *gin.Context) (any, error) {
		var loginVals types.LoginRequest
		// 使用 ShouldBindJSON 明确指定 JSON 绑定，避免自动选择绑定器的问题
		if err := c.ShouldBindJSON(&loginVals); err != nil {
			// 打印详细的绑定错误信息，方便调试
			log.Printf("Login bind error: %v", err)
			return "", jwt.ErrMissingLoginValues
		}
		username := loginVals.Username
		password := loginVals.Password

		// 打印登录尝试信息（生产环境中应该移除密码日志）
		log.Printf("Login attempt - username: %s, password length: %d", username, len(password))

		loginUser, _ := as.userService.IsValidUser(username, password)
		if loginUser != nil {
			// 返回指针类型，与 payloadFunc 中的类型断言匹配
			return &types.LoginResponse{Username: loginUser.Username, Email: loginUser.Email}, nil
		}
		return nil, jwt.ErrFailedAuthentication
	}
}

func (as *AuthService) authorizator() func(c *gin.Context, data any) bool {
	return func(c *gin.Context, data any) bool {
		log.Printf("authorizator data: %v", data)
		if v, ok := data.(*types.LoginResponse); ok && v.Username != "" {
			return true
		}
		return false
	}
}

func (as *AuthService) unauthorized() func(c *gin.Context, code int, message string) {
	return func(c *gin.Context, code int, message string) {
		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
		})
	}
}

func (as *AuthService) logoutResponse() func(c *gin.Context) {
	return func(c *gin.Context) {
		// This demonstrates that claims are now accessible during logout
		claims := jwt.ExtractClaims(c)
		user, exists := c.Get(identityKey)

		response := gin.H{
			"code":    http.StatusOK,
			"message": "Successfully logged out",
		}

		// Show that we can access user information during logout
		if len(claims) > 0 {
			response["logged_out_user"] = claims[identityKey]
		}
		if exists {
			response["user_info"] = user.(*models.User).Username
		}

		c.JSON(http.StatusOK, response)
	}
}
