package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/t-boris/ecommerce/tokens"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("token")
		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no authorization header"})
			c.Abort()
			return
		}
		claims, err := tokens.ValidateToken(clientToken)
		if err != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return
		}
		c.Set("email", claims.Email)
		c.Set("uid", claims.UID)
		c.Next()
	}
}
