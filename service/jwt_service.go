package service

import (
	"log"
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v3"
	"github.com/gin-gonic/gin"
	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/mycodeLife01/qa/config"
	"github.com/mycodeLife01/qa/models"
	"github.com/mycodeLife01/qa/types"
	"gorm.io/gorm"
)

type JwtService struct {
	userService *UserService
	DB          *gorm.DB
}

var identityKey = "username"

func NewJwtService(db *gorm.DB) *JwtService {
	return &JwtService{userService: NewUserService(db), DB: db}
}

func (js *JwtService) InitJwtMiddleware() (*jwt.GinJWTMiddleware, error) {
	authMiddleware, err := jwt.New(js.initParams())
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
		return nil, err
	}
	errInit := authMiddleware.MiddlewareInit()
	if errInit != nil {
		log.Fatal("authMiddleware.MiddlewareInit() Error:" + errInit.Error())
		return nil, errInit
	}
	return authMiddleware, nil
}

func (js *JwtService) initParams() *jwt.GinJWTMiddleware {

	return &jwt.GinJWTMiddleware{
		Realm:           "test zone",
		Key:             []byte(config.C.JWT.JWTSecretKey),
		Timeout:         time.Hour,
		MaxRefresh:      time.Hour,
		IdentityKey:     identityKey,
		PayloadFunc:     js.payloadFunc(),
		IdentityHandler: js.identityHandler(),
		Authenticator:   js.authenticator(),
		Authorizer:      js.authorizer(),
		Unauthorized:    js.unauthorized(),
		LogoutResponse:  js.logoutResponse(),
		TokenLookup:     "header: Authorization, query: token, cookie: jwt",
		TokenHeadName:   "Bearer",
		TimeFunc:        time.Now,
	}
}

func (js *JwtService) payloadFunc() func(data any) gojwt.MapClaims {
	return func(data any) gojwt.MapClaims {
		if v, ok := data.(*models.User); ok {
			return gojwt.MapClaims{
				identityKey: v.Username,
			}
		}
		return gojwt.MapClaims{}
	}
}

func (js *JwtService) identityHandler() func(c *gin.Context) any {
	return func(c *gin.Context) any {
		claims := jwt.ExtractClaims(c)
		// 安全地获取 username，避免 panic
		username, ok := claims[identityKey].(string)
		if !ok {
			log.Printf("identityHandler: failed to extract username from claims")
			return nil
		}
		return &models.User{
			Username: username,
		}
	}
}

func (js *JwtService) authenticator() func(c *gin.Context) (any, error) {
	return func(c *gin.Context) (any, error) {
		var loginVals types.LoginRequest
		// 使用 ShouldBindJSON 明确指定 JSON 绑定，避免自动选择绑定器的问题
		if err := c.ShouldBindJSON(&loginVals); err != nil {
			return "", jwt.ErrMissingLoginValues
		}
		username := loginVals.Username
		password := loginVals.Password

		loginUser, _ := js.userService.IsValidUser(username, password)
		if loginUser != nil {
			// 返回指针类型，与 payloadFunc 中的类型断言匹配
			// return &types.LoginResponse{Username: loginUser.Username, Email: loginUser.Email}, nil
			return loginUser, nil
		}
		return nil, jwt.ErrFailedAuthentication
	}
}

func (js *JwtService) authorizer() func(c *gin.Context, data any) bool {
	return func(c *gin.Context, data any) bool {
		if v, ok := data.(*models.User); ok && v.Username != "" {
			return true
		}
		return false
	}
}

func (js *JwtService) unauthorized() func(c *gin.Context, code int, message string) {
	return func(c *gin.Context, code int, message string) {
		c.JSON(code, gin.H{
			"code": code,
			"msg":  message,
			"data": nil,
		})
	}
}

func (js *JwtService) logoutResponse() func(c *gin.Context) {
	return func(c *gin.Context) {
		// This demonstrates that claims are now accessible during logout
		claims := jwt.ExtractClaims(c)
		user, exists := c.Get(identityKey)

		response := gin.H{
			"code": http.StatusOK,
			"msg":  "Successfully logged out",
			"data": gin.H{},
		}

		// Show that we can access user information during logout
		if len(claims) > 0 {
			response["data"].(gin.H)["logged_out_user"] = claims[identityKey]
		}
		if exists {
			response["data"].(gin.H)["user_info"] = user.(*models.User).Username
		}

		c.JSON(http.StatusOK, response)
	}
}
