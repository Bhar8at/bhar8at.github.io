package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Next()

		// executes before the function is returned
		defer func() {
			// recover captures the panic
			if err := recover(); err != nil {
				c.HTML(http.StatusInternalServerError, "errorT.html", gin.H{
					"error":   "500 Internal Server Error",
					"message": "An unexpected error occured, try again later.",
				})
			}
			c.Abort()
		}()
	}
}
