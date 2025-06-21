package log

import (
	"fmt"
	"time"

	"prodLoaderREST/internal/api/middlewares/requestid"

	"github.com/gin-gonic/gin"
)

func Logging(param gin.LogFormatterParams) string {
	return fmt.Sprintf("[%s] %s | %d | %s | %s | %s %s\n| ",
		param.Request.Context().Value(requestid.ContextKeyRequestID),
		param.TimeStamp.Format(time.DateTime),
		param.StatusCode,
		param.Latency,
		param.ClientIP,
		param.Method,
		param.Path,
	)
}
