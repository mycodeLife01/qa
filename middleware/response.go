package middleware

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mycodeLife01/qa/types"
)

const ResponseDataKey = "response_data"

func ResponseHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Next()

		// 异常处理逻辑
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			var appErr *types.AppError
			if errors.As(err, &appErr) {
				// 如果是自定义异常
				c.JSON(http.StatusOK, types.Fail(appErr.Message, appErr.Code))
				if appErr.Err != nil {
					log.Printf("[App Error] Code: %d, Message: %s, Underlying error: %v", appErr.Code, appErr.Message, appErr.Err)
				}
				return
			}
			// 如果是其它未知异常
			c.JSON(http.StatusInternalServerError, types.Fail("Internal Server Error", http.StatusInternalServerError))
			log.Printf("[Unknown Error] %v", err)
			return
		}

		// 检查 handler 是否已经自己写入了响应
		if c.Writer.Written() {
			return
		}
		// 正常返回
		data, exists := c.Get(ResponseDataKey)
		if exists {
			c.JSON(http.StatusOK, types.Success(data))
		} else {
			c.JSON(http.StatusOK, types.Success(nil))
		}
	}
}
