package middleware

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/internal/pkg/api"
)

func ResponseHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Next()

		// 异常处理逻辑
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			var appErr *api.AppError
			if errors.As(err, &appErr) {
				// 如果是自定义异常
				c.JSON(http.StatusOK, api.Fail(appErr.Message, appErr.Code))
				if appErr.Err != nil {
					log.Printf("[App Error] Code: %d, Message: %s, Underlying error: %v", appErr.Code, appErr.Message, appErr.Err)
				}
				return
			}
			// 如果是其它未知异常
			c.JSON(http.StatusInternalServerError, api.Fail("Internal Server Error", http.StatusInternalServerError))
			log.Printf("[Unknown Error] %v", err)
			return
		}

		// 检查 handler 是否已经自己写入了响应
		if c.Writer.Written() {
			return
		}

		// 检查是否是 404（路由不存在）
		if c.Writer.Status() == http.StatusNotFound {
			c.JSON(http.StatusNotFound, api.Fail("Route not found", http.StatusNotFound))
			return
		}

		// 检查是否是 405（方法不允许）
		if c.Writer.Status() == http.StatusMethodNotAllowed {
			c.JSON(http.StatusMethodNotAllowed, api.Fail("Method not allowed", http.StatusMethodNotAllowed))
			return
		}

		// 正常返回
		data, exists := c.Get(api.ResponseDataKey)
		if exists {
			c.JSON(http.StatusOK, api.Success(data))
		} else {
			c.JSON(http.StatusOK, api.Success(nil))
		}
	}
}
