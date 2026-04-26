package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": data})
}

func OKMsg(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": msg})
}

func Fail(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"code": code, "message": msg})
}

func CurrentUser(c *gin.Context) *User {
	val, exists := c.Get("currentUser")
	if !exists {
		return nil
	}
	u, ok := val.(*User)
	if !ok {
		return nil
	}
	return u
}
